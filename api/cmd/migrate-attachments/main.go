// migrate-attachments rewrites legacy /api/blob/<uuid> and inline data: URI
// image references in every vulnerability's evidence + remediation markdown
// into per-vulnerability scoped attachment rows.
//
// Reads CONFIG_PATH the same way the API does (so it picks up the same DB
// path). Idempotent: re-running after a successful pass is a no-op.
//
// Usage:
//
//	CONFIG_PATH=./config.yaml go run ./cmd/migrate-attachments --dry-run
//	CONFIG_PATH=./config.yaml go run ./cmd/migrate-attachments --commit
//
// One of --dry-run or --commit is required.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"prism/config"
	"prism/database"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "scan only; report what would change without writing")
	commit := flag.Bool("commit", false, "actually write attachment rows and rewrite markdown")
	flag.Parse()

	if *dryRun == *commit {
		fmt.Fprintln(os.Stderr, "exactly one of --dry-run or --commit is required")
		flag.Usage()
		os.Exit(2)
	}

	if err := config.LoadConfig(); err != nil {
		log.Fatalf("config: %v", err)
	}
	database.InitDB()

	mode := "dry-run"
	if *commit {
		mode = "commit"
	}
	fmt.Printf("== attachment migration (%s) ==\n", mode)

	report, err := database.MigrateAllAttachments(*dryRun)
	if err != nil {
		log.Fatalf("migrate: %v", err)
	}
	fmt.Println(report.String())

	if len(report.Errors) > 0 {
		os.Exit(1)
	}
}
