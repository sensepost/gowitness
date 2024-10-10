package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sensepost/gowitness/internal/ascii"
	"github.com/sensepost/gowitness/internal/islazy"
	"github.com/sensepost/gowitness/pkg/database"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/models"
	"github.com/sensepost/gowitness/pkg/writers"
	"github.com/spf13/cobra"
	"gorm.io/gorm/clause"
)

var conversionCmdExtensions = []string{".sqlite3", ".jsonl"}
var convertCmdFlags = struct {
	fromFile string
	toFile   string

	fromExt string
	toExt   string
}{}
var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert between SQLite and JSON Lines file formats",
	Long: ascii.LogoHelp(ascii.Markdown(`
# report convert

Convert between SQLite and JSON Lines file formats.

A --from-file and --to-file must be specified. The extension used for the
specified filenames will be used to determine the conversion direction and
target.`)),
	Example: ascii.Markdown(`
- gowitness report convert --from-file gowitness.sqlite3 --to-file data.jsonl
- gowitness report convert --from-file gowitness.jsonl --to-file db.sqlite3`),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if convertCmdFlags.fromFile == "" {
			return errors.New("from file not set")
		}
		if convertCmdFlags.toFile == "" {
			return errors.New("to file not set")
		}

		convertCmdFlags.fromExt = strings.ToLower(filepath.Ext(convertCmdFlags.fromFile))
		convertCmdFlags.toExt = strings.ToLower(filepath.Ext(convertCmdFlags.toFile))

		if convertCmdFlags.fromExt == "" || convertCmdFlags.toExt == "" {
			return errors.New("source and destination files must have extensions")
		}

		if convertCmdFlags.fromExt == convertCmdFlags.toExt {
			return errors.New("👀 source and destination file types must be different")
		}

		if convertCmdFlags.fromFile == convertCmdFlags.toFile {
			return errors.New("source and destination files cannot be the same")
		}

		if !islazy.SliceHasStr(conversionCmdExtensions, convertCmdFlags.fromExt) {
			return errors.New("unsupported from file type")
		}
		if !islazy.SliceHasStr(conversionCmdExtensions, convertCmdFlags.toExt) {
			return errors.New("unsupported to file type")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var writer writers.Writer
		var err error
		if convertCmdFlags.toExt == ".sqlite3" {
			writer, err = writers.NewDbWriter(fmt.Sprintf("sqlite://%s", convertCmdFlags.toFile), false)
			if err != nil {
				log.Error("could not get a database writer up", "err", err)
				return
			}
			if err := convertFromJsonlTo(convertCmdFlags.fromFile, writer); err != nil {
				log.Error("failed to convert to SQLite", "err", err)
				return
			}
		} else if convertCmdFlags.toExt == ".jsonl" {
			toFile, err := islazy.CreateFileWithDir(convertCmdFlags.toFile)
			if err != nil {
				log.Error("could not create target file", "err", err)
				return
			}
			writer, err = writers.NewJsonWriter(toFile)
			if err != nil {
				log.Error("could not get a JSON writer up", "err", err)
				return
			}
			if err := convertFromDbTo(convertCmdFlags.fromFile, writer); err != nil {
				log.Error("failed to convert to JSON Lines", "err", err)
				return
			}
		}
	},
}

func init() {
	reportCmd.AddCommand(convertCmd)

	convertCmd.Flags().StringVar(&convertCmdFlags.fromFile, "from-file", "", "The file to convert from")
	convertCmd.Flags().StringVar(&convertCmdFlags.toFile, "to-file", "", "The file to convert to. Use .sqlite3 for conversion to SQLite, and .jsonl for conversion to JSON Lines")
}

func convertFromDbTo(from string, writer writers.Writer) error {
	var results = []*models.Result{}
	conn, err := database.Connection(fmt.Sprintf("sqlite://%s", from), true, false)
	if err != nil {
		return err
	}

	if err := conn.Model(&models.Result{}).Preload(clause.Associations).Find(&results).Error; err != nil {
		return err
	}

	for _, result := range results {
		if err := writer.Write(result); err != nil {
			return err
		}
	}

	log.Info("converted from a database", "rows", len(results))
	return nil
}
func convertFromJsonlTo(from string, writer writers.Writer) error {
	file, err := os.Open(from)
	if err != nil {
		return err
	}
	defer file.Close()

	var c = 0

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				if len(line) == 0 {
					break // End of file
				}
				// Handle the last line without '\n'
			} else {
				return err
			}
		}

		var result models.Result
		if err := json.Unmarshal(line, &result); err != nil {
			log.Error("could not unmarshal JSON line", "err", err)
			continue
		}

		if err := writer.Write(&result); err != nil {
			return err
		}
		c++

		if err == io.EOF {
			break
		}
	}

	log.Info("converted from a JSON Lines file", "rows", c)
	return nil
}
