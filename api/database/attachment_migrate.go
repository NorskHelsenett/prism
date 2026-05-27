package database

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MigrationReport summarises what an attachment migration pass did or would do.
type MigrationReport struct {
	VulnerabilitiesScanned int
	VulnerabilitiesChanged int
	LegacyBlobsConverted   int
	LegacyBlobsMissing     int
	DataURIsConverted      int
	UnchangedExisting      int
	DecodeFailures         int
	Errors                 []string
}

func (r MigrationReport) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "vulns scanned         : %d\n", r.VulnerabilitiesScanned)
	fmt.Fprintf(&b, "vulns rewritten       : %d\n", r.VulnerabilitiesChanged)
	fmt.Fprintf(&b, "legacy /api/blob/<id> : %d converted, %d missing in db\n",
		r.LegacyBlobsConverted, r.LegacyBlobsMissing)
	fmt.Fprintf(&b, "inline data: URIs     : %d converted\n", r.DataURIsConverted)
	fmt.Fprintf(&b, "already-scoped URLs   : %d left alone\n", r.UnchangedExisting)
	fmt.Fprintf(&b, "decode failures       : %d\n", r.DecodeFailures)
	if len(r.Errors) > 0 {
		fmt.Fprintf(&b, "errors (%d):\n", len(r.Errors))
		for _, e := range r.Errors {
			fmt.Fprintf(&b, "  - %s\n", e)
		}
	}
	return b.String()
}

var (
	// Matches the legacy global blob URL pattern. Captures the filename
	// segment so we can look up the bytes in the old image_data table.
	legacyBlobURLRegex = regexp.MustCompile(`/api/blob/([a-zA-Z0-9\-]+(?:\.[a-zA-Z0-9]+)?)`)

	// Matches inline data: URIs in markdown image syntax. Captures the MIME
	// and the base64 payload so we can re-encode + persist them.
	dataURIRegex = regexp.MustCompile(`data:([a-zA-Z0-9.+/-]+);base64,([A-Za-z0-9+/=]+)`)

	// Matches per-vuln scoped URLs that already point at the new endpoint —
	// to detect already-migrated references so we can leave them alone.
	scopedURLRegex = regexp.MustCompile(`/api/vulnerability/(\d+)/attachments/([a-zA-Z0-9\-]+)`)
)

// migrationCandidatePredicate selects only vulnerabilities whose markdown
// could possibly contain something the normaliser would rewrite. The
// previous implementation loaded every JSONData row into memory on every
// boot — and on prod-sized data that was ~1 GiB of RSS for a no-op pass.
// SQLite executes this LIKE-scan against the JSON column without
// deserialising into Go, so it's cheap even without an index.
const migrationCandidatePredicate = `vulnerability LIKE '%/api/blob/%' OR vulnerability LIKE '%data:image%'`

// migrationBatchSize caps the number of JSONData rows held in memory at any
// one time during the batched scan. Each batch runs in its own transaction
// so we don't accumulate gorm session state across the whole walk.
const migrationBatchSize = 50

// MigrateAllAttachments walks vulnerabilities whose evidence/remediation
// markdown contains legacy /api/blob/<id> URLs or inline data: URIs and
// rewrites them into per-vuln attachment rows. Idempotent: running it twice
// is a near-no-op (just the predicate scan + a small cleanup pass).
//
// With dryRun=true nothing is written; the report reflects what would change.
func MigrateAllAttachments(dryRun bool) (MigrationReport, error) {
	var report MigrationReport

	// Pre-scan: cheap COUNT to decide whether to do any row loading at all.
	// Idempotent boots return early here — bounded memory + bounded time.
	var candidateCount int64
	if err := db.Model(&JSONData{}).Where(migrationCandidatePredicate).Count(&candidateCount).Error; err != nil {
		return report, fmt.Errorf("count migration candidates: %w", err)
	}
	if candidateCount == 0 {
		if !dryRun {
			pruneOrphanLegacyImages(&report)
		}
		return report, nil
	}

	if dryRun {
		var batch []JSONData
		err := db.Where(migrationCandidatePredicate).
			FindInBatches(&batch, migrationBatchSize, func(_ *gorm.DB, _ int) error {
				report.VulnerabilitiesScanned += len(batch)
				for i := range batch {
					v := batch[i]
					changed, err := normaliseVulnAttachments(db, &v, &report, true)
					if err != nil {
						report.Errors = append(report.Errors, fmt.Sprintf("vuln %d: %v", v.ID, err))
						continue
					}
					if changed {
						report.VulnerabilitiesChanged++
					}
				}
				return nil
			}).Error
		if err != nil {
			return report, fmt.Errorf("dry-run scan: %w", err)
		}
		return report, nil
	}

	// Commit mode: one transaction per batch. We still amortise fsyncs across
	// migrationBatchSize inserts/updates without holding the entire walk's
	// state in memory.
	var batch []JSONData
	err := db.Where(migrationCandidatePredicate).
		FindInBatches(&batch, migrationBatchSize, func(_ *gorm.DB, batchNumber int) error {
			processed := report.VulnerabilitiesScanned
			report.VulnerabilitiesScanned += len(batch)
			return db.Transaction(func(tx *gorm.DB) error {
				for i := range batch {
					v := batch[i]
					changed, err := normaliseVulnAttachments(tx, &v, &report, false)
					if err != nil {
						report.Errors = append(report.Errors, fmt.Sprintf("vuln %d: %v", v.ID, err))
						continue
					}
					if changed {
						report.VulnerabilitiesChanged++
						if err := tx.Save(&v).Error; err != nil {
							report.Errors = append(report.Errors,
								fmt.Sprintf("save vuln %d: %v", v.ID, err))
						}
					}
				}
				log.Printf("attachment migration: batch %d done (%d cumulative; rewritten=%d, legacy converted=%d, data-uris converted=%d)",
					batchNumber, processed+len(batch),
					report.VulnerabilitiesChanged,
					report.LegacyBlobsConverted,
					report.DataURIsConverted)
				return nil
			})
		}).Error
	if err != nil {
		return report, fmt.Errorf("migration walk: %w", err)
	}

	pruneOrphanLegacyImages(&report)
	return report, nil
}

// pruneOrphanLegacyImages clears the legacy image_data table when no
// vulnerability markdown references /api/blob/ anymore. The migration has
// already copied every referenced row's bytes into vulnerability_attachments,
// so the remaining rows are guaranteed to be orphans. This reverses the
// "database doubled" symptom observed on prod after the v1 migration left
// the legacy rows behind as a safety net.
func pruneOrphanLegacyImages(report *MigrationReport) {
	var stillReferenced int64
	if err := db.Model(&JSONData{}).Where("vulnerability LIKE '%/api/blob/%'").Count(&stillReferenced).Error; err != nil {
		report.Errors = append(report.Errors, fmt.Sprintf("count legacy refs: %v", err))
		return
	}
	if stillReferenced > 0 {
		log.Printf("attachment migration: %d vuln(s) still reference /api/blob/ — keeping legacy image_data for diagnosis", stillReferenced)
		return
	}
	var legacyRows int64
	if err := db.Model(&ImageData{}).Count(&legacyRows).Error; err != nil {
		report.Errors = append(report.Errors, fmt.Sprintf("count legacy rows: %v", err))
		return
	}
	if legacyRows == 0 {
		return
	}
	// Unscoped().Delete drops the rows hard (no soft-delete tombstone). They
	// were never accessed through any route that respects gorm.DeletedAt
	// anyway — image_data has no model-level filtering on it.
	if err := db.Unscoped().Where("1 = 1").Delete(&ImageData{}).Error; err != nil {
		report.Errors = append(report.Errors, fmt.Sprintf("prune legacy: %v", err))
		return
	}
	log.Printf("attachment migration: pruned %d orphan image_data rows", legacyRows)
}

// MigrateVulnAttachments runs the same normalisation as the batch migration
// but for a single in-memory vulnerability — used at PUT/PATCH save time so
// new content (pasted data: URIs, copied legacy URLs) is persisted as
// attachments before the row hits the database.
//
// Mutates v in place. Returns whether any rewrite happened so the caller
// can decide whether to re-save. Always commits new attachment rows
// (no dry-run here).
func MigrateVulnAttachments(v *JSONData) (bool, error) {
	var report MigrationReport
	changed, err := normaliseVulnAttachments(db, v, &report, false)
	if err != nil {
		return false, err
	}
	if len(report.Errors) > 0 {
		return changed, fmt.Errorf("normalise: %s", report.Errors[0])
	}
	return changed, nil
}

// normaliseVulnAttachments inspects the JSON-encoded vulnerability payload,
// extracts evidence + remediation, rewrites image references, and folds the
// result back into v.Vulnerability. Returns whether any change was made.
// All DB work flows through tx so the caller can batch many vulns into a
// single transaction.
func normaliseVulnAttachments(tx *gorm.DB, v *JSONData, report *MigrationReport, dryRun bool) (bool, error) {
	if len(v.Vulnerability) == 0 {
		return false, nil
	}
	var payload map[string]any
	if err := json.Unmarshal(v.Vulnerability, &payload); err != nil {
		return false, fmt.Errorf("unmarshal: %w", err)
	}

	changed := false
	for _, field := range []string{"evidence", "remediation"} {
		raw, ok := payload[field].(string)
		if !ok || raw == "" {
			continue
		}
		rewritten, err := normaliseMarkdownReferences(tx, v.ID, v.FoundBy, raw, report, dryRun)
		if err != nil {
			return false, fmt.Errorf("%s: %w", field, err)
		}
		if rewritten != raw {
			payload[field] = rewritten
			changed = true
		}
	}

	if !changed {
		return false, nil
	}
	merged, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("re-marshal: %w", err)
	}
	v.Vulnerability = merged
	return true, nil
}

// normaliseMarkdownReferences scans a markdown string and rewrites every
// legacy or inline image reference into a scoped attachment URL. Each
// distinct reference creates at most one attachment row (deduplicated within
// the same vulnerability). All DB work goes through tx.
func normaliseMarkdownReferences(
	tx *gorm.DB,
	vulnID uint,
	foundBy string,
	markdown string,
	report *MigrationReport,
	dryRun bool,
) (string, error) {
	// First count already-scoped URLs so the report is accurate even when
	// nothing else changes.
	report.UnchangedExisting += len(scopedURLRegex.FindAllString(markdown, -1))

	// In-memory dedup for this single pass: a vuln that references the same
	// legacy blob twice should produce exactly one attachment row.
	legacyMapping := map[string]string{}
	dataURIMapping := map[string]string{}

	// Rewrite inline data: URIs first (longer match, otherwise the legacy
	// regex's broad character class can swallow part of base64 payloads).
	markdown = dataURIRegex.ReplaceAllStringFunc(markdown, func(match string) string {
		sub := dataURIRegex.FindStringSubmatch(match)
		if len(sub) != 3 {
			return match
		}
		if mapped, ok := dataURIMapping[match]; ok {
			return mapped
		}
		mime := normaliseMime(sub[1])
		payload := sub[2]
		bytes, err := base64.StdEncoding.DecodeString(payload)
		if err != nil {
			report.DecodeFailures++
			report.Errors = append(report.Errors,
				fmt.Sprintf("vuln %d: data URI base64: %v", vulnID, err))
			return match
		}
		if !AllowedAttachmentMime(mime) || !VerifyAttachmentMagic(mime, bytes) {
			report.DecodeFailures++
			report.Errors = append(report.Errors,
				fmt.Sprintf("vuln %d: data URI MIME %q failed verification", vulnID, mime))
			return match
		}
		url, err := persistMigrationAttachment(tx, vulnID, foundBy, "inline-image", mime, bytes, dryRun)
		if err != nil {
			report.DecodeFailures++
			report.Errors = append(report.Errors,
				fmt.Sprintf("vuln %d: persist inline image: %v", vulnID, err))
			return match
		}
		report.DataURIsConverted++
		dataURIMapping[match] = url
		return url
	})

	// Now rewrite legacy /api/blob/<filename> references.
	markdown = legacyBlobURLRegex.ReplaceAllStringFunc(markdown, func(match string) string {
		filename := match[len("/api/blob/"):]
		if mapped, ok := legacyMapping[filename]; ok {
			return mapped
		}
		var img ImageData
		if err := tx.Where("filename = ?", filename).First(&img).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				report.LegacyBlobsMissing++
				return match
			}
			report.Errors = append(report.Errors,
				fmt.Sprintf("vuln %d: load legacy blob %s: %v", vulnID, filename, err))
			return match
		}
		mime := SniffAttachmentMime(img.Data)
		if !AllowedAttachmentMime(mime) || !VerifyAttachmentMagic(mime, img.Data) {
			report.DecodeFailures++
			report.Errors = append(report.Errors,
				fmt.Sprintf("vuln %d: legacy blob %s failed sniff/magic", vulnID, filename))
			return match
		}
		url, err := persistMigrationAttachment(tx, vulnID, foundBy, filename, mime, img.Data, dryRun)
		if err != nil {
			report.DecodeFailures++
			report.Errors = append(report.Errors,
				fmt.Sprintf("vuln %d: persist legacy blob %s: %v", vulnID, filename, err))
			return match
		}
		report.LegacyBlobsConverted++
		legacyMapping[filename] = url
		return url
	})

	return markdown, nil
}

type RegenReport struct {
	Total       int
	Regenerated int
	Skipped     int
	Errors      []string
}

func (r RegenReport) String() string {
	out := fmt.Sprintf("regenerated %d/%d proxies (skipped %d)", r.Regenerated, r.Total, r.Skipped)
	if len(r.Errors) > 0 {
		out += fmt.Sprintf(" — %d errors", len(r.Errors))
	}
	return out
}

// RegenerateAllProxies walks every existing attachment and re-encodes its
// proxy using the current Settings.AttachmentMaxEdge. Triggered from the
// admin Settings panel when the operator changes the resolution tier.
// Idempotent: re-running with the same setting produces equivalent bytes.
//
// The original is the source of truth; this function never touches
// OriginalData. Wrapped in a single transaction for the same fsync reason
// as MigrateAllAttachments.
func RegenerateAllProxies() (RegenReport, error) {
	var report RegenReport
	maxEdge := EffectiveAttachmentMaxEdge()

	// Count up front so the report's Total is accurate; the prior
	// db.Find(&attachments) loaded every blob's bytes into Go memory
	// simultaneously, which was the other plausible source of the
	// ~1 GiB RSS spike on prod.
	var total int64
	if err := db.Model(&VulnerabilityAttachment{}).Count(&total).Error; err != nil {
		return report, fmt.Errorf("count attachments: %w", err)
	}
	report.Total = int(total)

	const batchSize = 25
	var batch []VulnerabilityAttachment
	err := db.FindInBatches(&batch, batchSize, func(_ *gorm.DB, batchNumber int) error {
		return db.Transaction(func(tx *gorm.DB) error {
			for i := range batch {
				a := batch[i]
				if AttachmentKind(a.Mime) != "image" || len(a.OriginalData) == 0 {
					report.Skipped++
					continue
				}
				proxy, proxyMime, err := EncodeAttachmentProxy(a.OriginalData, maxEdge)
				if err != nil {
					report.Errors = append(report.Errors,
						fmt.Sprintf("attachment %d (key=%s): %v", a.ID, a.Key, err))
					continue
				}
				if err := tx.Model(&VulnerabilityAttachment{}).
					Where("id = ?", a.ID).
					Updates(map[string]any{"proxy_data": proxy, "proxy_mime": proxyMime}).Error; err != nil {
					report.Errors = append(report.Errors,
						fmt.Sprintf("attachment %d: update: %v", a.ID, err))
					continue
				}
				report.Regenerated++
			}
			log.Printf("proxy regen: batch %d done (regenerated=%d, skipped=%d, errors=%d)",
				batchNumber, report.Regenerated, report.Skipped, len(report.Errors))
			return nil
		})
	}).Error
	if err != nil {
		return report, fmt.Errorf("regen walk: %w", err)
	}
	return report, nil
}

// persistMigrationAttachment creates a VulnerabilityAttachment row from raw
// bytes during migration. In dryRun mode it returns a stable simulated URL
// without touching the database, so the caller can still measure the
// rewrite shape and counts. All writes go through tx so callers can batch
// the whole migration into one transaction.
func persistMigrationAttachment(
	tx *gorm.DB,
	vulnID uint,
	foundBy string,
	filename string,
	mime string,
	data []byte,
	dryRun bool,
) (string, error) {
	var (
		proxy     []byte
		proxyMime string
	)
	if AttachmentKind(mime) == "image" {
		var err error
		proxy, proxyMime, err = EncodeAttachmentProxy(data, EffectiveAttachmentMaxEdge())
		if err != nil {
			return "", fmt.Errorf("encode proxy: %w", err)
		}
	}
	key := uuid.New().String()
	url := "/api/vulnerability/" + strconv.FormatUint(uint64(vulnID), 10) + "/attachments/" + key
	if dryRun {
		return url, nil
	}
	att := &VulnerabilityAttachment{
		VulnerabilityID: vulnID,
		Key:             key,
		Filename:        filename,
		Mime:            mime,
		OriginalData:    data,
		ProxyData:       proxy,
		ProxyMime:       proxyMime,
		UploadedBy:      foundBy,
	}
	if err := tx.Create(att).Error; err != nil {
		return "", err
	}
	return url, nil
}
