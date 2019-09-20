package cmd

import (
	"encoding/json"
	"image/png"
	"os"
	"strconv"

	"github.com/corona10/goimagehash"
	"github.com/olekukonko/tablewriter"
	"github.com/sensepost/gowitness/storage"
	gwtmpl "github.com/sensepost/gowitness/template"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tidwall/buntdb"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List entries in the gowitness database",
	Long:  `List entries in the gowitness database`,
	Run: func(cmd *cobra.Command, args []string) {
		err := db.Db.View(func(tx *buntdb.Tx) error {

			var hash uint64 = 0

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"URL", "HTTP Code", "Title"})
			table.SetColWidth(colWidth)

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

				table.Append([]string{
					data.FinalURL, strconv.Itoa(data.ResponseCode), data.Title,
				})

				return true
			})

			table.Render()

			return nil
		})

		if err != nil {
			log.Fatal(err)
		}

	},
}

func init() {
	reportCmd.AddCommand(listCmd)
	listCmd.Flags().IntVarP(&colWidth, "column-width", "w", 120, "The column width use to print")
}
