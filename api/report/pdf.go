// Package report renders an immutable PDF from a frozen
// models.ReportVersionPayload. The PDF is generated once at publish
// time and stored on the ReportVersion row; subsequent downloads serve
// those bytes unchanged so historic versions stay reproducible.
package report

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"

	"prism/models"
)

// RenderPDF turns a frozen payload into PDF bytes.
func RenderPDF(payload models.ReportVersionPayload) ([]byte, error) {
	cfg := config.NewBuilder().
		WithPageSize(pagesize.A4).
		Build()

	m := maroto.New(cfg)

	addCover(m, payload)
	addExecutiveSummary(m, payload)
	addProjects(m, payload)
	addFindings(m, payload)

	doc, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("generate pdf: %w", err)
	}
	return doc.GetBytes(), nil
}

func addCover(m core.Maroto, p models.ReportVersionPayload) {
	m.AddRows(
		row.New(60),
		row.New(20).Add(
			text.NewCol(12, p.Title, props.Text{
				Size:  24,
				Align: align.Center,
				Style: fontstyle.Bold,
			}),
		),
		row.New(10).Add(
			text.NewCol(12, fmt.Sprintf("Version %d", p.Version), props.Text{
				Size:  14,
				Align: align.Center,
			}),
		),
		row.New(8).Add(
			text.NewCol(12, p.PublishedAt.Format("2 January 2006"), props.Text{
				Size:  12,
				Align: align.Center,
				Color: &props.Color{Red: 90, Green: 90, Blue: 90},
			}),
		),
		row.New(8).Add(
			text.NewCol(12, "Published by "+p.PublishedBy, props.Text{
				Size:  10,
				Align: align.Center,
				Color: &props.Color{Red: 130, Green: 130, Blue: 130},
			}),
		),
	)
}

func addExecutiveSummary(m core.Maroto, p models.ReportVersionPayload) {
	m.AddRows(
		row.New(20),
		sectionHeader("Executive Summary"),
	)
	for _, line := range wrapParagraphs(stripMarkdown(p.ExecutiveSummary)) {
		m.AddRow(6, text.NewCol(12, line, props.Text{Size: 11, Align: align.Left}))
	}
}

func addProjects(m core.Maroto, p models.ReportVersionPayload) {
	m.AddRows(
		row.New(10),
		sectionHeader("Projects in scope"),
	)
	for _, proj := range p.Projects {
		m.AddRow(6, text.NewCol(12, "• "+proj.Name, props.Text{Size: 11}))
	}
}

func addFindings(m core.Maroto, p models.ReportVersionPayload) {
	m.AddRows(
		row.New(10),
		sectionHeader(fmt.Sprintf("Findings (%d)", len(p.Findings))),
	)
	if len(p.Findings) == 0 {
		m.AddRow(8, text.NewCol(12, "No findings included in this report.", props.Text{
			Size:  11,
			Style: fontstyle.Italic,
			Color: &props.Color{Red: 120, Green: 120, Blue: 120},
		}))
		return
	}
	for i, f := range p.Findings {
		m.AddRow(8)
		m.AddRow(8,
			text.NewCol(8, fmt.Sprintf("%d. %s", i+1, f.Title), props.Text{
				Size:  13,
				Style: fontstyle.Bold,
			}),
			text.NewCol(4, severityLabel(f.Severity), props.Text{
				Size:  11,
				Align: align.Right,
				Style: fontstyle.Bold,
				Color: severityColor(f.Severity),
			}),
		)
		meta := fmt.Sprintf("Status: %s   ·   Project: %s", or(f.Status, "Reported"), or(f.ProjectName, "—"))
		m.AddRow(5, text.NewCol(12, meta, props.Text{
			Size:  9,
			Color: &props.Color{Red: 110, Green: 110, Blue: 110},
		}))
		for _, line := range wrapParagraphs(stripMarkdown(f.Summary)) {
			m.AddRow(5, text.NewCol(12, line, props.Text{Size: 10}))
		}
	}
}

func sectionHeader(label string) core.Row {
	return row.New(10).Add(
		text.NewCol(12, label, props.Text{
			Size:  16,
			Style: fontstyle.Bold,
			Color: &props.Color{Red: 30, Green: 30, Blue: 30},
		}),
	)
}

func severityLabel(sev string) string {
	if sev == "" {
		return "—"
	}
	return strings.ToUpper(sev)
}

func severityColor(sev string) *props.Color {
	switch strings.ToLower(strings.TrimSpace(sev)) {
	case "critical":
		return &props.Color{Red: 153, Green: 0, Blue: 0}
	case "high":
		return &props.Color{Red: 204, Green: 51, Blue: 0}
	case "medium":
		return &props.Color{Red: 204, Green: 153, Blue: 0}
	case "low":
		return &props.Color{Red: 0, Green: 102, Blue: 153}
	case "info", "informational":
		return &props.Color{Red: 90, Green: 90, Blue: 90}
	default:
		return &props.Color{Red: 60, Green: 60, Blue: 60}
	}
}

func or(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}

// stripMarkdown is a tiny markdown→plain converter for body text in PDF.
// Removes headings, emphasis markers, code fences, image syntax, and turns
// links into "text (url)". Good enough for executive summaries and finding
// summaries; full markdown rendering belongs in the web UI.
var (
	mdImage   = regexp.MustCompile(`!\[([^\]]*)\]\([^)]*\)`)
	mdLink    = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	mdCode    = regexp.MustCompile("`([^`]+)`")
	mdFence   = regexp.MustCompile("(?s)```[a-zA-Z0-9]*\\n(.*?)```")
	mdHeading = regexp.MustCompile(`(?m)^#{1,6}\s+`)
	mdBold    = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	mdItalic  = regexp.MustCompile(`\*([^*]+)\*|_([^_]+)_`)
	mdList    = regexp.MustCompile(`(?m)^\s*[-*]\s+`)
	mdQuote   = regexp.MustCompile(`(?m)^>\s?`)
)

func stripMarkdown(s string) string {
	s = mdImage.ReplaceAllString(s, "")
	s = mdLink.ReplaceAllString(s, "$1 ($2)")
	s = mdFence.ReplaceAllString(s, "$1")
	s = mdCode.ReplaceAllString(s, "$1")
	s = mdHeading.ReplaceAllString(s, "")
	s = mdBold.ReplaceAllString(s, "$1")
	s = mdItalic.ReplaceAllStringFunc(s, func(m string) string {
		inner := strings.Trim(m, "*_")
		return inner
	})
	s = mdList.ReplaceAllString(s, "• ")
	s = mdQuote.ReplaceAllString(s, "")
	return s
}

// wrapParagraphs splits text into paragraphs by blank lines, then by single
// newlines; each entry becomes one PDF row. Very long lines are passed
// through — maroto wraps inside the text cell.
func wrapParagraphs(s string) []string {
	if strings.TrimSpace(s) == "" {
		return []string{""}
	}
	parts := strings.Split(s, "\n")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			out = append(out, "")
			continue
		}
		out = append(out, trimmed)
	}
	return out
}
