package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/glebarez/sqlite"
	"github.com/sensepost/gowitness/internal/ascii"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/models"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var mergeCmdFlags = struct {
	SourceFiles []string
	SourcePath  string
	OutputFile  string
}{}
var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge multiple SQLite databases into a single database",
	Long: ascii.LogoHelp(ascii.Markdown(`
# report merge

Merge multiple SQLite databases into a single database.

You can specify source files using --source-file (can be specified multiple
times) or a directory containing multiple SQLite databases using --source-path.
The command will scan for databases that match the required schema and merge
their data.`)),
	Example: ascii.Markdown(`
- gowitness report merge --source-path ./databases --output-file merged.sqlite3
- gowitness report merge --source-file gowitness.db --source-file db2.sqlite3 --output-file merged.sqlite3`),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(mergeCmdFlags.SourceFiles) == 0 && mergeCmdFlags.SourcePath == "" {
			return errors.New("either --source-file or --source-path must be specified")
		}
		if mergeCmdFlags.OutputFile == "" {
			return errors.New("output file not specified")
		}

		if mergeCmdFlags.SourcePath != "" {
			isDir, err := isDirectory(mergeCmdFlags.SourcePath)
			if err != nil {
				return fmt.Errorf("failed to access source path: %w", err)
			}
			if !isDir {
				return errors.New("--source-path must be a directory")
			}
		}

		// Check if source files exist
		for _, file := range mergeCmdFlags.SourceFiles {
			if _, err := os.Stat(file); os.IsNotExist(err) {
				return fmt.Errorf("source file does not exist: %s", file)
			}
			if !isSQLiteDatabase(file) {
				return fmt.Errorf("source file is not a valid SQLite database: %s", file)
			}
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var dbFiles []string

		if len(mergeCmdFlags.SourceFiles) > 0 {
			dbFiles = append(dbFiles, mergeCmdFlags.SourceFiles...)
		}

		if mergeCmdFlags.SourcePath != "" {
			filesFromDir, err := findSQLiteDatabases(mergeCmdFlags.SourcePath)
			if err != nil {
				log.Error("failed to find SQLite databases", "err", err)
				return
			}
			if len(filesFromDir) == 0 && len(dbFiles) == 0 {
				log.Error("no SQLite databases found in the specified directory or files")
				return
			}
			dbFiles = append(dbFiles, filesFromDir...)
		}

		if len(dbFiles) == 0 {
			log.Error("no SQLite databases to process")
			return
		}

		// remove duplicates from dbFiles
		dbFiles = removeDuplicateFiles(dbFiles)

		// create the output database
		destDB, err := createOutputDatabase(mergeCmdFlags.OutputFile)
		if err != nil {
			log.Error("failed to create output database", "err", err)
			return
		}

		// Iterate over each source database and copy data
		for _, dbFile := range dbFiles {
			log.Info("processing database", "database", dbFile)

			sourceDB, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
			if err != nil {
				log.Error("failed to open source database", "dbFile", dbFile, "err", err)
				continue
			}

			// Verify the schema
			hasSchema, err := hasRequiredSchema(sourceDB)
			if err != nil {
				log.Error("failed to verify schema", "dbFile", dbFile, "err", err)
				continue
			}
			if !hasSchema {
				log.Warn("database does not have the required schema", "dbFile", dbFile)
				continue
			}

			// Copy data
			if err := copyData(sourceDB, destDB); err != nil {
				log.Error("failed to copy data", "dbFile", dbFile, "err", err)
				continue
			}
		}

		log.Info("data merge completed successfully.")
	},
}

func init() {
	reportCmd.AddCommand(mergeCmd)

	mergeCmd.Flags().StringSliceVar(&mergeCmdFlags.SourceFiles, "source-file", nil, "One or more source SQLite database files")
	mergeCmd.Flags().StringVar(&mergeCmdFlags.SourcePath, "source-path", "", "The source directory containing SQLite databases")
	mergeCmd.Flags().StringVar(&mergeCmdFlags.OutputFile, "output-file", "", "The output SQLite database file")
}

func isDirectory(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func findSQLiteDatabases(dir string) ([]string, error) {
	var dbFiles []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(info.Name()))
			if ext == ".sqlite" || ext == ".sqlite3" || ext == ".db" {
				dbFiles = append(dbFiles, path)
			} else {
				if isSQLiteDatabase(path) {
					dbFiles = append(dbFiles, path)
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return dbFiles, nil
}

func isSQLiteDatabase(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	header := make([]byte, 16)
	if _, err := file.Read(header); err != nil {
		return false
	}
	return string(header[:15]) == "SQLite format 3"
}

func hasRequiredSchema(db *gorm.DB) (bool, error) {
	// Check if the 'results' table exists
	var count int64
	err := db.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?", "results").Scan(&count).Error
	if err != nil {
		return false, err
	}
	if count == 0 {
		return false, nil
	}
	return true, nil
}

func createOutputDatabase(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	// Migrate the schema
	if err := db.AutoMigrate(
		&models.Result{},
		&models.TLS{},
		&models.TLSSanList{},
		&models.Technology{},
		&models.Header{},
		&models.NetworkLog{},
		&models.ConsoleLog{},
		&models.Cookie{},
	); err != nil {
		return nil, err
	}

	return db, nil
}

func copyData(source *gorm.DB, dest *gorm.DB) error {
	batchSize := 10
	var results []models.Result
	if err := source.Model(&models.Result{}).Preload(clause.Associations).Preload("TLS.SanList").
		FindInBatches(&results, batchSize, func(tx *gorm.DB, batch int) error {
			// Begin a transaction in the destination database
			return dest.Transaction(func(destTx *gorm.DB) error {
				for _, result := range results {
					// Reset IDs
					result.ID = 0
					// Remove associations
					headers := result.Headers
					result.Headers = nil
					networkLogs := result.Network
					result.Network = nil
					consoleLogs := result.Console
					result.Console = nil
					cookies := result.Cookies
					result.Cookies = nil
					technologies := result.Technologies
					result.Technologies = nil
					tlsData := result.TLS
					result.TLS = models.TLS{}

					// Insert Result
					if err := destTx.Create(&result).Error; err != nil {
						return fmt.Errorf("failed to insert Result: %w", err)
					}
					newResultID := result.ID

					// Insert TLS Data
					if tlsData.Protocol != "" || tlsData.Issuer != "" {
						tlsData.ID = 0
						tlsData.ResultID = newResultID
						sanList := tlsData.SanList
						tlsData.SanList = nil

						if err := destTx.Create(&tlsData).Error; err != nil {
							return fmt.Errorf("failed to insert TLS data: %w", err)
						}
						newTLSID := tlsData.ID

						// Insert SanList
						for i := range sanList {
							sanList[i].ID = 0
							sanList[i].TLSID = newTLSID
						}
						if len(sanList) > 0 {
							if err := destTx.Create(&sanList).Error; err != nil {
								return fmt.Errorf("failed to insert TLS SanList: %w", err)
							}
						}
					}

					// Insert Headers
					for i := range headers {
						headers[i].ID = 0
						headers[i].ResultID = newResultID
					}
					if len(headers) > 0 {
						if err := destTx.Create(&headers).Error; err != nil {
							return fmt.Errorf("failed to insert Headers: %w", err)
						}
					}

					// Insert Network Logs
					for i := range networkLogs {
						networkLogs[i].ID = 0
						networkLogs[i].ResultID = newResultID
					}
					if len(networkLogs) > 0 {
						if err := destTx.Create(&networkLogs).Error; err != nil {
							return fmt.Errorf("failed to insert Network Logs: %w", err)
						}
					}

					// Insert Console Logs
					for i := range consoleLogs {
						consoleLogs[i].ID = 0
						consoleLogs[i].ResultID = newResultID
					}
					if len(consoleLogs) > 0 {
						if err := destTx.Create(&consoleLogs).Error; err != nil {
							return fmt.Errorf("failed to insert Console Logs: %w", err)
						}
					}

					// Insert Cookies
					for i := range cookies {
						cookies[i].ID = 0
						cookies[i].ResultID = newResultID
					}
					if len(cookies) > 0 {
						if err := destTx.Create(&cookies).Error; err != nil {
							return fmt.Errorf("failed to insert Cookies: %w", err)
						}
					}

					// Insert Technologies
					for i := range technologies {
						technologies[i].ID = 0
						technologies[i].ResultID = newResultID
					}
					if len(technologies) > 0 {
						if err := destTx.Create(&technologies).Error; err != nil {
							return fmt.Errorf("failed to insert Technologies: %w", err)
						}
					}
				}
				return nil
			})

		}).Error; err != nil {
		return err
	}

	return nil
}

func removeDuplicateFiles(files []string) []string {
	fileMap := make(map[string]struct{})
	var uniqueFiles []string

	for _, file := range files {
		absPath, err := filepath.Abs(file)
		if err != nil {
			absPath = file
		}
		if _, exists := fileMap[absPath]; !exists {
			fileMap[absPath] = struct{}{}
			uniqueFiles = append(uniqueFiles, file)
		}
	}
	return uniqueFiles
}
