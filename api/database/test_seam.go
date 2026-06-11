package database

import "gorm.io/gorm"

// SetDBForTest swaps the package-level db handle and returns the previous
// one. Only meant to be called from tests in other packages (notably
// prism/routes) that need to point the data layer at an in-memory sqlite.
// Pair every call with a defer to restore.
func SetDBForTest(testDB *gorm.DB) *gorm.DB {
	prev := db
	db = testDB
	return prev
}

// DBForTest exposes the package-level db handle for tests that need to seed
// rows directly through GORM without going through the typed helpers.
func DBForTest() *gorm.DB {
	return db
}
