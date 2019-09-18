package cmd

import (
	"bytes"
	"encoding/json"
	"image/png"
	"io/ioutil"
	"os"
	"sort"
	"text/template"

	"github.com/corona10/goimagehash"
	gwtmpl "github.com/sensepost/gowitness/template"
	log "github.com/sirupsen/logrus"

	"github.com/sensepost/gowitness/storage"
	"github.com/sensepost/gowitness/utils"
	"github.com/spf13/cobra"
	"github.com/tidwall/buntdb"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate an HTML report from a database file",
	Long: `
Generate an HTML report of the screenshot information found in a gowitness.db file

For example:

$ gowitness generate`,
	Run: func(cmd *cobra.Command, args []string) {
		validateGenerateFlags()

		// Populate a variable with the data the template will
		// want to parse
		var screenshotEntries []storage.HTTResponse
		var hash uint64 = 0
		err := db.Db.View(func(tx *buntdb.Tx) error {

			tx.Ascend("", func(key, value string) bool {

				data := storage.HTTResponse{}
				if err := json.Unmarshal([]byte(value), &data); err != nil {
					log.Fatal(err)
				}

				// check if the screenshot path exists. if not, slide in
				// a placeholder image
				if _, err := os.Stat(data.ScreenshotFile); os.IsNotExist(err) {

					log.WithField("screenshot-file", data.ScreenshotFile).
						Debug("Adding placeholder for missing screenshot")
					data.ScreenshotFile = gwtmpl.PlaceHolderImage
				}

				// calculate image hash
				if perceptionSort {
					file, _ := os.Open(data.ScreenshotFile)
					defer file.Close()

					img, err := png.Decode(file)
					if err == nil {
						computation, _ := goimagehash.PerceptionHash(img)
						hash = computation.GetHash()
					}
				}
				data.Hash = hash

				log.WithField("url", data.FinalURL).Debug("Generating screenshot entry")

				// filters – http status codes
				if len(filterStatusCodes) > 0 {

					if utils.SliceContainsInt(filterStatusCodes, data.ResponseCode) {
						screenshotEntries = append(screenshotEntries, data)
					}

				} else {
					screenshotEntries = append(screenshotEntries, data)
				}

				return true
			})

			// Sort by Image Perception Hashes
			if perceptionSort {
				sort.Slice(screenshotEntries, func(i, j int) bool {
					return screenshotEntries[i].Hash < screenshotEntries[j].Hash
				})
			}

			// Sort by Status Codes
			if statusCodeSort {
				sort.Slice(screenshotEntries, func(i, j int) bool {
					return screenshotEntries[i].ResponseCode < screenshotEntries[j].ResponseCode
				})
			}

			return nil
		})

		if err != nil {
			log.Fatal(err)
		}

		if len(screenshotEntries) <= 0 {
			log.WithField("count", len(screenshotEntries)).Error("No screenshot entries exist to create a report")
			return
		}

		// Prepare and render the template
		tmpl, err := template.New("go-witness-report").Parse(gwtmpl.HTMLContent)
		if err != nil {
			log.WithField("err", err).Fatal("Failed to parse template")
		}

		type TemplateData struct {
			ScreenShots []storage.HTTResponse
		}
		templateData := TemplateData{ScreenShots: screenshotEntries}

		var doc bytes.Buffer
		tmpl.Execute(&doc, templateData)

		if err := ioutil.WriteFile(reportFileName, []byte(doc.String()), 0644); err != nil {
			log.WithField("err", err).Fatal("Failed to write report html")
		}

		log.WithField("report-file", reportFileName).Info("Report generated")
	},
}

func init() {
	reportCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVarP(&reportFileName, "name", "n", "report.html", "Destination report filename")
	generateCmd.Flags().BoolVarP(&perceptionSort, "sort-perception", "P", false, "Sort screenshots with perception hashing")
	generateCmd.Flags().BoolVarP(&statusCodeSort, "sort-status-code", "S", false, "Sort screenshots by HTTP status codes")
	generateCmd.Flags().IntSliceVarP(&filterStatusCodes, "filter-code", "C", []int{}, "The HTTP status code to filter on (Can specify more than one --filter-status-codes)")
}

func validateGenerateFlags() {
	if perceptionSort && statusCodeSort {
		log.Fatal("Only one sort option is allowed")
	}
}
