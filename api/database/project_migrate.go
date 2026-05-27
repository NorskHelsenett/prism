package database

import (
	"fmt"
	"log"
)

// EnsureProjectColumns adds the StartDate / EndDate / Color / GroupID / SortOrder
// columns to project_data when GORM's AutoMigrate fails to do so. We've seen this
// happen on long-lived prod databases where the table predates the new fields and
// AutoMigrate silently skips the ALTER (its error path is unchecked at the call
// site). This catch-up uses SQLite's tolerant ALTER TABLE ADD COLUMN behaviour
// and is idempotent — already-present columns are detected via PRAGMA table_info
// before any ALTER is attempted.
func EnsureProjectColumns() error {
	required := []struct {
		Name string
		DDL  string
	}{
		{"start_date", "ALTER TABLE project_data ADD COLUMN start_date DATE"},
		{"end_date", "ALTER TABLE project_data ADD COLUMN end_date DATE"},
		{"color", "ALTER TABLE project_data ADD COLUMN color TEXT"},
		{"group_id", "ALTER TABLE project_data ADD COLUMN group_id INTEGER"},
		{"sort_order", "ALTER TABLE project_data ADD COLUMN sort_order INTEGER DEFAULT 0"},
	}

	existing, err := tableColumns("project_data")
	if err != nil {
		return fmt.Errorf("read project_data columns: %w", err)
	}

	for _, c := range required {
		if existing[c.Name] {
			continue
		}
		if err := db.Exec(c.DDL).Error; err != nil {
			return fmt.Errorf("add column %s: %w", c.Name, err)
		}
		log.Printf("EnsureProjectColumns: added %s", c.Name)
	}

	// Best-effort index on group_id for the planning views' lookups.
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_project_data_group_id ON project_data(group_id)").Error; err != nil {
		log.Printf("EnsureProjectColumns: could not create group_id index: %v", err)
	}

	return nil
}

func tableColumns(table string) (map[string]bool, error) {
	rows, err := db.Raw(fmt.Sprintf("PRAGMA table_info(%s)", table)).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols := map[string]bool{}
	for rows.Next() {
		var (
			cid     int
			name    string
			ctype   string
			notnull int
			dflt    any
			pk      int
		)
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return nil, err
		}
		cols[name] = true
	}
	return cols, rows.Err()
}
