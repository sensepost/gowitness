package cmd

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"os"
	"path/filepath"

	"github.com/sensepost/gowitness/internal/ascii"
	"github.com/sensepost/gowitness/internal/islazy"
	"github.com/sensepost/gowitness/pkg/database"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/models"
	"github.com/sensepost/gowitness/web/templates"
	"github.com/spf13/cobra"
	"gorm.io/gorm/clause"
)

var generateCmdFlags = struct {
	ReportFile     string
	ScreenshotPath string
	DbURI          string
	JsonFile       string

	// temp working dir
	TempDir string
}{}
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a static HTML report",
	Long: ascii.LogoHelp(ascii.Markdown(`
# report generate

Generate a static HTML report.

This command accepts a screenshot path as well as a data source. The screenshot
path should the the directory that contains screenshots (as named by gowitness).
The data source can be a JSON Lines file, or a database URI (ie.
sqlite://yourdatabase.sqlite3).

The output file is a zip archive with an index.html file containing the report.
`)),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		if generateCmdFlags.DbURI == "" && generateCmdFlags.JsonFile == "" {
			return errors.New("no data source defined")
		}

		generateCmdFlags.TempDir, err = os.MkdirTemp("", "gowitness3-report-*")
		if err != nil {
			return err
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var results = []models.Result{}

		// if we have a json path, use that
		if generateCmdFlags.JsonFile != "" {
			file, err := os.Open(generateCmdFlags.JsonFile)
			if err != nil {
				log.Fatal("could not open JSON Lines file", "err", err)
			}
			defer file.Close()

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
						log.Fatal("error reading JSON Lines file", "err", err)
					}
				}

				var result models.Result
				if err := json.Unmarshal(line, &result); err != nil {
					log.Error("could not unmarshal JSON line", "err", err)
					continue
				}
				results = append(results, result)

				if err == io.EOF {
					break
				}
			}

			if err := generateHTML(results); err != nil {
				log.Fatal("an error occurred generating the html report", "err", err)
			}

			return
		}

		// but, db-uri is the default
		conn, err := database.Connection(generateCmdFlags.DbURI, true, false)
		if err != nil {
			log.Fatal("could not connect to database", "err", err)
		}

		if err := conn.Model(&models.Result{}).Preload(clause.Associations).Find(&results).Error; err != nil {
			log.Fatal("could not get list", "err", err)
		}

		if err := generateHTML(results); err != nil {
			log.Fatal("an error occurred generating the html report", "err", err)
		}
	},
}

func init() {
	reportCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVar(&generateCmdFlags.ScreenshotPath, "screenshot-path", "./screenshots", "The path where screenshots are stored")
	generateCmd.Flags().StringVar(&generateCmdFlags.DbURI, "db-uri", "sqlite://gowitness.sqlite3", "The location of a gowitness database")
	generateCmd.Flags().StringVar(&generateCmdFlags.JsonFile, "json-file", "", "The location of a JSON Lines results file (e.g., ./gowitness.jsonl). This flag takes precedence over --db-uri")
	generateCmd.Flags().StringVar(&generateCmdFlags.ReportFile, "zip-name", "gowitness-report.zip", "The name and location of the final report ZIP file that will be generated")
}

// statusClass returns a CSS class based on the HTTP status code
func statusClass(responseCode int) string {
	switch {
	case responseCode >= 200 && responseCode < 300:
		return "status-2xx"
	case responseCode >= 300 && responseCode < 400:
		return "status-3xx"
	case responseCode >= 400 && responseCode < 500:
		return "status-4xx"
	case responseCode >= 500:
		return "status-5xx"
	default:
		return ""
	}
}

// generateHTML generates an HTML report from results
func generateHTML(results []models.Result) error {
	log.Info("generating HTML report for results", "count", len(results))

	tmplContent, err := templates.ReportTemplate.ReadFile("static-report.tmpl")
	if err != nil {
		return err
	}

	funcMap := template.FuncMap{
		"statusClass": statusClass,
	}

	tmpl, err := template.New("report").Funcs(funcMap).Parse(string(tmplContent))
	if err != nil {
		return err
	}

	htmlOutputPath := filepath.Join(generateCmdFlags.TempDir, "index.html")
	file, err := os.Create(htmlOutputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = tmpl.Execute(file, map[string]interface{}{
		"Results": results,
	})
	if err != nil {
		return err
	}

	cssOutputPath := copyCSS()

	// archive the results
	tempZipPath := filepath.Join(generateCmdFlags.TempDir, "report.zip")
	err = createZipFile(tempZipPath, []string{htmlOutputPath, cssOutputPath}, generateCmdFlags.ScreenshotPath)
	if err != nil {
		return err
	}

	// move the archive to where the user wanted it (by name)
	finalZipPath := generateCmdFlags.ReportFile
	err = islazy.MoveFile(tempZipPath, finalZipPath)
	if err != nil {
		return err
	}

	log.Info("report zip file generated successfully", "path", finalZipPath)

	// clean up the temporary directory
	err = os.RemoveAll(generateCmdFlags.TempDir)
	if err != nil {
		return err
	}

	return nil
}

// copyCSS extracts the embedded pico.min.css file and writes it to the temp dir
func copyCSS() string {
	cssContent, err := templates.ReportTemplate.ReadFile("pico.min.css")
	if err != nil {
		log.Fatal("failed to read CSS file", "err", err)
	}

	cssOutputPath := filepath.Join(generateCmdFlags.TempDir, "pico.min.css")
	err = os.WriteFile(cssOutputPath, cssContent, 0644)
	if err != nil {
		log.Fatal("failed to write CSS file", "err", err)
	}

	return cssOutputPath
}

// createZipFile creates the report zip archive
func createZipFile(outputZip string, filesToInclude []string, screenshotsDir string) error {
	zipFile, err := os.Create(outputZip)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, filePath := range filesToInclude {
		if err := addFileToZip(zipWriter, filePath); err != nil {
			return err
		}
	}

	if err := addScreenshotsToZip(zipWriter, screenshotsDir); err != nil {
		return err
	}

	return nil
}

// addScreenshotsToZip adds all files in the user specified screenshot directory into
// the ./screenshots/ directory in the zip file.
func addScreenshotsToZip(zipWriter *zip.Writer, screenshotsDir string) error {
	err := filepath.Walk(screenshotsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// screenshots always live in ./screenshots for the report
		relativePath := filepath.Join("screenshots", filepath.Base(path))

		zipFileWriter, err := zipWriter.Create(relativePath)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(zipFileWriter, file)
		if err != nil {
			return err
		}

		log.Debug("Added screenshot to ZIP", "file", path)
		return nil
	})

	return err
}

// addFileToZip adds a file to a zip Writer
func addFileToZip(zipWriter *zip.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	zipFileWriter, err := zipWriter.Create(filepath.Base(filePath))
	if err != nil {
		return err
	}

	_, err = io.Copy(zipFileWriter, file)
	if err != nil {
		return err
	}

	log.Debug("Added file to ZIP", "file", filePath)
	return nil
}
