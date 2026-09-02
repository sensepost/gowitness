package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/sensepost/gowitness/pkg/models"
	"gorm.io/gorm"
)

func TestSQLiteMigrationWithForeignKeys(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "gowitness.sqlite3")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	if err := db.AutoMigrate(
		&models.Result{},
		&models.TLS{},
		&models.TLSSanList{},
	); err != nil {
		t.Fatalf("failed to create initial schema: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql db: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("failed to close initial db: %v", err)
	}

	uri := "sqlite://" + filepath.ToSlash(dbPath)

	c, err := Connection(uri, false, false)
	if err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	sqlDB, err = c.DB()
	if err != nil {
		t.Fatalf("failed to get migrated sql db: %v", err)
	}
	defer sqlDB.Close()

	var violations []struct {
		Table  string
		RowID  int
		Parent string
		FKID   int
	}

	if err := c.Raw("PRAGMA foreign_key_check").Scan(&violations).Error; err != nil {
		t.Fatalf("foreign key check failed: %v", err)
	}

	if len(violations) != 0 {
		t.Fatalf("expected no foreign key violations, got %d", len(violations))
	}

	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("database disappeared after migration: %v", err)
	}
}
