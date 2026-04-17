package database

import (
	"errors"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"prism/models"
)

const notePreviewLength = 180

func initNotesFTS() {
	// FTS5 virtual table (external content pointing at the notes table).
	// Storing only the columns we want to search; `content` lets FTS5 pull
	// the indexed text from the main table without duplicating storage.
	if err := db.Exec(`
		CREATE VIRTUAL TABLE IF NOT EXISTS notes_fts USING fts5(
			title, content,
			content='notes', content_rowid='id',
			tokenize = 'unicode61 remove_diacritics 2'
		);
	`).Error; err != nil {
		log.Printf("notes_fts create failed (search will fall back to LIKE): %v", err)
		return
	}

	triggers := []string{
		`DROP TRIGGER IF EXISTS notes_ai;`,
		`DROP TRIGGER IF EXISTS notes_ad;`,
		`DROP TRIGGER IF EXISTS notes_au;`,
		`CREATE TRIGGER notes_ai AFTER INSERT ON notes BEGIN
			INSERT INTO notes_fts(rowid, title, content) VALUES (new.id, new.title, new.content);
		END;`,
		`CREATE TRIGGER notes_ad AFTER DELETE ON notes BEGIN
			INSERT INTO notes_fts(notes_fts, rowid, title, content) VALUES('delete', old.id, old.title, old.content);
		END;`,
		`CREATE TRIGGER notes_au AFTER UPDATE ON notes BEGIN
			INSERT INTO notes_fts(notes_fts, rowid, title, content) VALUES('delete', old.id, old.title, old.content);
			INSERT INTO notes_fts(rowid, title, content) VALUES (new.id, new.title, new.content);
		END;`,
	}
	for _, stmt := range triggers {
		if err := db.Exec(stmt).Error; err != nil {
			log.Printf("notes_fts trigger failed: %v", err)
		}
	}
}

// Tags are Bear-style inline #hashtags. Extracted fresh on every save so the
// stored `tags` column stays in sync with the content.
var tagRegex = regexp.MustCompile(`(?:^|\s)#([a-zA-Z][\w\-\/]*)`)

func ExtractTags(content string) []string {
	stripped := stripCodeAndUrls(content)
	matches := tagRegex.FindAllStringSubmatch(stripped, -1)
	seen := map[string]struct{}{}
	out := []string{}
	for _, m := range matches {
		tag := strings.ToLower(m[1])
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	sort.Strings(out)
	return out
}

var codeFenceRegex = regexp.MustCompile("(?s)```.*?```")
var inlineCodeRegex = regexp.MustCompile("`[^`]*`")
var mdLinkRegex = regexp.MustCompile(`\[([^\]]*)\]\(([^)]*)\)`)
var bareUrlRegex = regexp.MustCompile(`https?://\S+`)

func stripCodeAndUrls(s string) string {
	s = codeFenceRegex.ReplaceAllString(s, " ")
	s = inlineCodeRegex.ReplaceAllString(s, " ")
	s = mdLinkRegex.ReplaceAllString(s, "$1")
	s = bareUrlRegex.ReplaceAllString(s, " ")
	return s
}

// deriveTitle pulls the first non-empty line, stripped of leading markdown
// heading markers, and truncates to a reasonable length.
func DeriveTitle(content string) string {
	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		line = strings.TrimLeft(line, "# ")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if len(line) > 120 {
			line = line[:120]
		}
		return line
	}
	return ""
}

var markdownStripRegex = regexp.MustCompile(`(?m)^#+\s+|[*_~` + "`" + `>\[\]()!]|!\[[^\]]*\]\([^)]*\)`)

func DerivePreview(content string) string {
	// Remove the first line (it becomes the title) so the preview doesn't
	// just repeat it.
	lines := strings.SplitN(content, "\n", 2)
	body := ""
	if len(lines) == 2 {
		body = lines[1]
	}
	body = stripCodeAndUrls(body)
	body = markdownStripRegex.ReplaceAllString(body, "")
	body = strings.ReplaceAll(body, "\n", " ")
	body = regexp.MustCompile(`\s+`).ReplaceAllString(body, " ")
	body = strings.TrimSpace(body)
	if len(body) > notePreviewLength {
		body = body[:notePreviewLength] + "…"
	}
	return body
}

func tagsToString(tags []string) string {
	return strings.Join(tags, ",")
}

func tagsFromString(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}

func toListItem(n models.Note) models.NoteListItem {
	item := models.NoteListItem{
		ID:        n.ID,
		Title:     n.Title,
		Preview:   n.Preview,
		Tags:      tagsFromString(n.Tags),
		UpdatedAt: n.UpdatedAt,
	}
	if n.DeletedAt.Valid {
		t := n.DeletedAt.Time
		item.DeletedAt = &t
	}
	return item
}

func CreateNote(userEmail, content string) (*models.Note, error) {
	note := &models.Note{
		UserEmail: userEmail,
		Content:   content,
		Title:     DeriveTitle(content),
		Preview:   DerivePreview(content),
		Tags:      tagsToString(ExtractTags(content)),
	}
	if err := db.Create(note).Error; err != nil {
		return nil, err
	}
	return note, nil
}

func GetNote(userEmail string, id uint) (*models.Note, error) {
	var note models.Note
	err := db.Where("id = ? AND user_email = ?", id, userEmail).First(&note).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &note, nil
}

type ListNotesOptions struct {
	Query string
	Tag   string
	Trash bool
}

func ListNotes(userEmail string, opts ListNotesOptions) ([]models.NoteListItem, error) {
	var notes []models.Note
	q := db.Model(&models.Note{}).Where("user_email = ?", userEmail)
	if opts.Trash {
		q = q.Unscoped().Where("deleted_at IS NOT NULL")
	}

	if opts.Query != "" {
		ids, err := searchNoteIDs(userEmail, opts.Query, opts.Trash)
		if err != nil {
			return nil, err
		}
		if len(ids) == 0 {
			return []models.NoteListItem{}, nil
		}
		q = q.Where("id IN ?", ids)
	}

	if opts.Tag != "" {
		tag := strings.ToLower(opts.Tag)
		// stored as comma-joined lowercased tags; match whole tokens to avoid
		// "work" matching "workshop"
		q = q.Where("(',' || tags || ',') LIKE ?", "%,"+tag+",%")
	}

	order := "updated_at DESC"
	if opts.Trash {
		order = "deleted_at DESC"
	}
	if err := q.Order(order).Limit(500).Find(&notes).Error; err != nil {
		return nil, err
	}
	out := make([]models.NoteListItem, 0, len(notes))
	for _, n := range notes {
		out = append(out, toListItem(n))
	}
	return out, nil
}

func searchNoteIDs(userEmail, query string, trash bool) ([]uint, error) {
	// Build an FTS5 MATCH expression. We quote each term and suffix * so
	// typing "auth" matches "authentication". Strip characters that would
	// break the FTS parser.
	sanitized := regexp.MustCompile(`[^\w\s]`).ReplaceAllString(query, " ")
	parts := strings.Fields(sanitized)
	if len(parts) == 0 {
		return nil, nil
	}
	for i, p := range parts {
		parts[i] = `"` + p + `"*`
	}
	matchExpr := strings.Join(parts, " ")

	rows, err := db.Raw(`
		SELECT n.id FROM notes n
		JOIN notes_fts fts ON fts.rowid = n.id
		WHERE notes_fts MATCH ? AND n.user_email = ?
		  AND (? OR n.deleted_at IS NULL)
		  AND (NOT ? OR n.deleted_at IS NOT NULL)
	`, matchExpr, userEmail, trash, trash).Rows()
	if err != nil {
		// FTS may not be available; fall back to LIKE.
		return likeSearchNoteIDs(userEmail, query, trash)
	}
	defer rows.Close()
	ids := []uint{}
	for rows.Next() {
		var id uint
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func likeSearchNoteIDs(userEmail, query string, trash bool) ([]uint, error) {
	like := "%" + query + "%"
	q := db.Model(&models.Note{}).Where("user_email = ?", userEmail).
		Where("title LIKE ? OR content LIKE ?", like, like)
	if trash {
		q = q.Unscoped().Where("deleted_at IS NOT NULL")
	}
	var notes []models.Note
	if err := q.Limit(500).Find(&notes).Error; err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(notes))
	for _, n := range notes {
		ids = append(ids, n.ID)
	}
	return ids, nil
}

func ListAllTags(userEmail string) ([]string, error) {
	var rows []struct{ Tags string }
	err := db.Model(&models.Note{}).
		Where("user_email = ? AND tags <> ''", userEmail).
		Select("tags").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	out := []string{}
	for _, r := range rows {
		for _, t := range strings.Split(r.Tags, ",") {
			if t == "" {
				continue
			}
			if _, ok := seen[t]; ok {
				continue
			}
			seen[t] = struct{}{}
			out = append(out, t)
		}
	}
	sort.Strings(out)
	return out, nil
}

func UpdateNote(userEmail string, id uint, content string) (*models.Note, error) {
	note, err := GetNote(userEmail, id)
	if err != nil {
		return nil, err
	}
	note.Content = content
	note.Title = DeriveTitle(content)
	note.Preview = DerivePreview(content)
	note.Tags = tagsToString(ExtractTags(content))
	note.UpdatedAt = time.Now()
	if err := db.Save(note).Error; err != nil {
		return nil, err
	}
	return note, nil
}

func SoftDeleteNote(userEmail string, id uint) error {
	res := db.Where("id = ? AND user_email = ?", id, userEmail).Delete(&models.Note{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func RestoreNote(userEmail string, id uint) (*models.Note, error) {
	res := db.Unscoped().Model(&models.Note{}).
		Where("id = ? AND user_email = ? AND deleted_at IS NOT NULL", id, userEmail).
		Update("deleted_at", nil)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	var note models.Note
	if err := db.Where("id = ? AND user_email = ?", id, userEmail).First(&note).Error; err != nil {
		return nil, err
	}
	return &note, nil
}

func HardDeleteNote(userEmail string, id uint) error {
	// Only allow hard delete of already-trashed rows. This matches the UX
	// where the trash view is the only place that exposes permanent delete.
	res := db.Unscoped().
		Where("id = ? AND user_email = ? AND deleted_at IS NOT NULL", id, userEmail).
		Delete(&models.Note{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func EmptyTrash(userEmail string) (int64, error) {
	res := db.Unscoped().
		Where("user_email = ? AND deleted_at IS NOT NULL", userEmail).
		Delete(&models.Note{})
	return res.RowsAffected, res.Error
}
