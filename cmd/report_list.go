package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/sensepost/gowitness/storage"
	"github.com/spf13/cobra"
)

// reportListCmd represents the reportList command
var reportListCmd = &cobra.Command{
	Use:   "list",
	Short: "List entries in the gowitness database in various formats",
	Long: `List entries in the gowitness database in various formats.

Export formats include CSV (via the --csv / -c flag) as well as JSON (via
the --json / -j flag).

When using the JSON format you could chain the results with an invocation of
[1]jq to further filter the data. For example, to get only URL's for which
the status code was an HTTP 200:

gowitness report list -j | jq -r ". | select(.response_code==200) | .final_url"

[1] https://stedolan.github.io/jq/`,
	Example: `$ gowitness report list
$ gowitness report list --json
$ gowitness report list --csv --sort`,
	Run: func(cmd *cobra.Command, args []string) {
		log := options.Logger

		db, err := db.Get()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get a db handle")
		}

		rows, err := db.Scopes(storage.OrderPerception(options.PerceptionSort)).
			Model(&storage.URL{}).Rows()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get rows")
		}
		defer rows.Close()

		var data []storage.URL
		for rows.Next() {
			url := &storage.URL{}
			db.ScanRows(rows, url)
			data = append(data, *url)
		}

		if options.ReportJSON {
			if err := outputJSON(&data); err != nil {
				log.Fatal().Err(err).Msg("failed to output json")
			}
			return
		}

		if options.ReportCSV {
			outputCSV(&data)
			return
		}

		outputTable(&data)
	},
}

func init() {
	reportCmd.AddCommand(reportListCmd)

	reportListCmd.Flags().BoolVarP(&options.ReportJSON, "json", "j", false, "output json")
	reportListCmd.Flags().BoolVarP(&options.ReportCSV, "csv", "c", false, "output csv")
	reportListCmd.Flags().BoolVarP(&options.PerceptionSort, "sort", "S", false, "sort by image perceptions")
}

// outputJSON prints the report in JSON format
func outputJSON(d *[]storage.URL) error {

	for _, l := range *d {
		bytes, err := json.Marshal(l)
		if err != nil {
			return err
		}
		fmt.Print(string(bytes))
	}
	return nil
}

// outputCSV prints the report in CSV format
func outputCSV(d *[]storage.URL) {

	wr := csv.NewWriter(os.Stdout)
	for _, l := range *d {
		wr.Write(l.MarshallCSV())
	}
	wr.Flush()
}

// outputTable prints the output to stdout in table format
func outputTable(d *[]storage.URL) {

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"final url", "status", "title"})
	for _, l := range *d {
		table.Append([]string{l.FinalURL, strconv.Itoa(l.ResponseCode), l.Title})
	}
	table.Render()
}
