package cmd

import (
	"os"
	"path/filepath"

	"github.com/h2non/filetype"
	"github.com/sensepost/gowitness/storage"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// mergeCmd represents the merge command
var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge gowitness sqlite databases",
	Long: `Merge gotiwness sqlite databases.

Provided a directory or multiple -i / --input flags, this command will read each
database it can find, populating a fresh, merged sqlite database.
When providing a directory with --input-path, that directory is recursively
walked with each file checked to see if it is a sqlite database. If so, it will
form part of the merge process. You can mix both -i and --input-path flags.

Duplicates are not ignored from source databases. Instead, they get a fresh primary
key in the new destination database.`,
	Example: `$ gowitness merge -i gowitness-1.sqlite3 -i gowitness-2.sqlite3
$ gowitness merge -i gowitness-1.sqlite3 -i gowitness-2.sqlite3 --input-path dbs/
$ gowitness merge --input-path dbs/
$ gowitness merge --input-path dbs/ -o merged.sqlite.3
$ gowitness merge -i gowitness.sqlite3 --input-path dbs/ --output merged.sqlite3`,
	Run: func(cmd *cobra.Command, args []string) {
		log := options.Logger

		if err := readDirDbs(); err != nil {
			log.Fatal().Err(err).Msg("failed to read source path")
		}

		if len(options.MergeDBs) <= 0 {
			log.Fatal().Msg("we have no databases to merge. specify some!")
		} else {
			log.Info().Int("database-count", len(options.MergeDBs)).Msg("number of dbs to process")
		}

		if len(options.MergeDBs) == 1 {
			log.Warn().Msg(`merging just one database does not make sense. make a copy instead?`)
			return
		}

		// get a handle for the fresh, merged db
		dstDB := storage.NewDb()
		dstDB.Location = options.MergeOutputDB
		dstDBConn, err := dstDB.Get()

		if err != nil {
			log.Fatal().Err(err).Str("destination", options.MergeOutputDB).
				Msg("could not open destination db")
		}

		log.Info().Str("db-path", options.MergeOutputDB).Msg("writing results to a new database")

		for _, file := range options.MergeDBs {
			log.Info().Str("file", file).Msg("processing source database")
			if err = mergeFromPath(file, dstDBConn); err != nil {
				log.Error().Err(err).Str("source-db", file).Msg("failed to merge database")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(mergeCmd)

	mergeCmd.Flags().StringVarP(&options.MergeOutputDB, "output", "o", "gowitness-merged.sqlite3", "output database file name")
	mergeCmd.Flags().StringVar(&options.MergeSourcePath, "input-path", "", "a path containing sqlite databases to merge")
	mergeCmd.Flags().StringSliceVarP(&options.MergeDBs, "input", "i", []string{}, "input database file location (supports multiple)")
}

// readDirDbs reads a directory, scanning for sqlite databases
func readDirDbs() error {

	if options.MergeSourcePath == "" {
		return nil
	}

	if err := filepath.Walk(options.MergeSourcePath, func(path string, _ os.FileInfo, err error) error {

		// todo: add option to do non-recursive walking

		// check that the file at least _looks_ like a sqlite db
		file, _ := os.Open(path)
		defer file.Close()
		head := make([]byte, 261)
		file.Read(head)

		kind, _ := filetype.Match(head)
		if kind.MIME.Value != "application/vnd.sqlite3" {
			return nil
		}

		options.MergeDBs = append(options.MergeDBs, path)

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// mergeFromPath will read a sqlite db specified by a path, populate
// the results into a db handle
func mergeFromPath(source string, dst *gorm.DB) error {

	log := options.Logger

	srcDB := storage.NewDb()
	srcDB.Location = source
	srcDB.SkipMigration = true

	db, err := srcDB.Get()

	if err != nil {
		return err
	}

	// read results from the current source database in chunks of 500
	// records, and populate each into the dst database handle. the
	// primary key is unset with result.ID = 0 so that a new key can
	// be populated in the dst database.
	var results []*storage.URL
	result := db.Model(&storage.URL{}).Preload(clause.Associations).
		FindInBatches(&results, 500, func(tx *gorm.DB, batch int) error {
			log.Debug().Int("batch-number", batch).Msg("working with batch")

			for _, result := range results {
				result.ID = 0 // unset primarykey
			}

			dst.Create(&results)

			return nil
		})

	if result.Error != nil {
		return result.Error
	}

	log.Info().Int64("processed-rows", result.RowsAffected).Str("source-db", source).
		Msg("done processing db")

	return nil
}
