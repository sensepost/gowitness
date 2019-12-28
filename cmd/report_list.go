package cmd

import (
	"encoding/json"
	"image/png"
	"os"
	"strconv"
	"sort"

	"github.com/corona10/goimagehash"
	"github.com/olekukonko/tablewriter"
	"github.com/sensepost/gowitness/storage"
	gwtmpl "github.com/sensepost/gowitness/template"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tidwall/buntdb"
	"github.com/sensepost/gowitness/utils"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List entries in the gowitness database",
	Long:  `List entries in the gowitness database`,
	Run: func(cmd *cobra.Command, args []string) {
		validateGenerateFlags()

		// Populate a variable with the data the template will
		// want to parse
		var screenshotEntries []storage.HTTResponse
		var hash uint64 = 0

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"URL", "HTTP Code", "Title", "Hash"})
		table.SetColWidth(colWidth)

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

				// filters - ignore errors
				if ignoreFailed && (400 <= data.ResponseCode && data.ResponseCode < 600) {
					return true
				}

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
					if screenshotEntries[i].Hash == screenshotEntries[j].Hash {
						return screenshotEntries[i].ResponseCode < screenshotEntries[j].ResponseCode
					}
					return screenshotEntries[i].Hash < screenshotEntries[j].Hash
				})
			}

			// Sort by Status Codes
			if statusCodeSort {
				sort.Slice(screenshotEntries, func(i, j int) bool {
					return screenshotEntries[i].ResponseCode < screenshotEntries[j].ResponseCode
				})
			}

			// Sort by Title
			if titleSort {
				sort.Slice(screenshotEntries, func(i, j int) bool {
					return screenshotEntries[i].Title < screenshotEntries[j].Title
				})
			}

			return nil
		})

		if err != nil {
			log.Fatal(err)
		}
		// Sort the screenshots
		for _, element := range screenshotEntries {
			table.Append([]string{
						element.FinalURL, strconv.Itoa(element.ResponseCode), element.Title, strconv.FormatUint(element.Hash,10),
			})
		}
		table.Render()
	},
}

func init() {
	reportCmd.AddCommand(listCmd)
	listCmd.Flags().IntVarP(&colWidth, "column-width", "w", 120, "The column width use to print")
	listCmd.Flags().BoolVarP(&perceptionSort, "sort-perception", "P", false, "Sort screenshots with perception hashing")
	listCmd.Flags().BoolVarP(&statusCodeSort, "sort-status-code", "S", false, "Sort screenshots by HTTP status codes")
	listCmd.Flags().BoolVarP(&titleSort, "sort-title", "L", false, "Sort screenshots by parsed <title> tags")
}


