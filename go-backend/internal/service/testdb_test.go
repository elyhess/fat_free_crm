package service

import (
	"fmt"
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testTables = []string{
	"emails", "comments", "versions", "permissions", "groups_users", "groups",
	"leads", "contacts", "accounts", "users",
}

func testDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := buildTestDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("connect to test db: %v", err)
	}

	// Truncate before test so each test starts with a clean slate
	for _, table := range testTables {
		db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
	}

	t.Cleanup(func() {
		for _, table := range testTables {
			db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
		}
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})

	return db
}

func buildTestDSN() string {
	host := envOr("DB_HOST", "localhost")
	port := envOr("DB_PORT", "5432")
	user := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	dbName := envOr("DB_TEST_DATABASE", "fat_free_crm_elyhess_test")
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
