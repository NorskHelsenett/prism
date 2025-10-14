package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"prism/config"
	"prism/database"
	"prism/tests/integration"
)

func main() {
	log.Println("Starting database seeding...")

	// Initialize configuration
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	database.InitDB()

	// Clean existing data
	log.Println("Cleaning existing test data...")
	if err := integration.CleanDatabase(); err != nil {
		log.Printf("Warning: Failed to clean database: %v", err)
	}

	// Seed fixtures
	fixturesDir := os.Getenv("FIXTURES_DIR")
	if fixturesDir == "" {
		fixturesDir = "./fixtures"
	}

	projectsFile := filepath.Join(fixturesDir, "projects.csv")
	vulnsFile := filepath.Join(fixturesDir, "vulnerabilities.csv")

	log.Println("Seeding database with fixtures...")
	if err := integration.SeedDatabase(projectsFile, vulnsFile); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	log.Println("✓ Database seeding completed successfully!")

	// Print summary
	projects, _ := integration.LoadProjects(projectsFile)
	vulns, _ := integration.LoadVulnerabilities(vulnsFile)
	users, _ := integration.LoadUsers(filepath.Join(filepath.Dir(projectsFile), "users.csv"))
	apiKeys, _ := integration.LoadAPIKeys()

	fmt.Printf("\nSeeded Data Summary:\n")
	fmt.Printf("  Projects: %d\n", len(projects))
	fmt.Printf("  Vulnerabilities: %d\n", len(vulns))
	fmt.Printf("  Users: %d\n", len(users))
	fmt.Printf("  API Keys: %d\n", len(apiKeys))
	fmt.Println("\nDatabase is ready for testing!")
}
