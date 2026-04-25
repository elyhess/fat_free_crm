package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func main() {
	_ = godotenv.Load()

	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	dsn := buildDSN()
	migrationsDir := findMigrationsDir()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("failed to set dialect: %v", err)
	}

	switch command {
	case "up":
		err = goose.Up(db, migrationsDir)
	case "up-one":
		err = goose.UpByOne(db, migrationsDir)
	case "down":
		err = goose.Down(db, migrationsDir)
	case "status":
		err = goose.Status(db, migrationsDir)
	case "version":
		err = goose.Version(db, migrationsDir)
	case "create":
		if len(args) < 1 {
			log.Fatal("create requires a migration name")
		}
		err = goose.Create(db, migrationsDir, args[0], "sql")
	case "redo":
		err = goose.Redo(db, migrationsDir)
	case "fix":
		err = goose.Fix(migrationsDir)
	case "mark-applied":
		// Mark baseline as applied without running it (for existing databases).
		err = markBaselineApplied(db, migrationsDir)
	default:
		usage()
		os.Exit(1)
	}

	if err != nil {
		log.Fatalf("migration %s failed: %v", command, err)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: migrate <command> [args]

Commands:
  up             Apply all pending migrations
  up-one         Apply the next pending migration
  down           Roll back the last migration
  status         Show migration status
  version        Show current migration version
  create <name>  Create a new migration file
  redo           Roll back and re-apply the last migration
  mark-applied   Mark baseline migration as applied (for existing databases)
`)
}

func buildDSN() string {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}

	host := envOr("DB_HOST", "localhost")
	port := envOr("DB_PORT", "5432")
	user := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	dbName := envOr("DB_DATABASE", "fat_free_crm_elyhess_development")
	sslMode := envOr("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s dbname=%s sslmode=%s", host, port, dbName, sslMode)
	if user != "" {
		dsn += fmt.Sprintf(" user=%s", user)
	}
	if password != "" {
		dsn += fmt.Sprintf(" password=%s", password)
	}
	return dsn
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func findMigrationsDir() string {
	// Try relative to the binary, then relative to the source file
	candidates := []string{
		"db/migrations",
		"../../db/migrations",
	}

	// Also try relative to the source file location (for `go run`)
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		candidates = append(candidates, filepath.Join(filepath.Dir(filename), "../../db/migrations"))
	}

	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			abs, _ := filepath.Abs(dir)
			return abs
		}
	}

	log.Fatal("could not find db/migrations directory")
	return ""
}

// markBaselineApplied inserts the baseline migration into goose's version table
// without running it. Use this on databases that already have the Rails schema.
func markBaselineApplied(db *sql.DB, dir string) error {
	// Ensure the goose version table exists
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS goose_db_version (
		id SERIAL PRIMARY KEY,
		version_id BIGINT NOT NULL,
		is_applied BOOLEAN NOT NULL,
		tstamp TIMESTAMP DEFAULT now()
	)`); err != nil {
		return fmt.Errorf("create goose table: %w", err)
	}

	// Check if baseline is already marked
	var count int
	db.QueryRow("SELECT COUNT(*) FROM goose_db_version WHERE version_id = 1").Scan(&count)
	if count > 0 {
		fmt.Println("baseline migration (00001) is already marked as applied")
		return nil
	}

	// Insert baseline as applied
	if _, err := db.Exec("INSERT INTO goose_db_version (version_id, is_applied) VALUES (1, true)"); err != nil {
		return fmt.Errorf("mark baseline: %w", err)
	}

	fmt.Println("marked baseline migration (00001) as applied")
	return nil
}
