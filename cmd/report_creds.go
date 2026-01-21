package cmd

import (
	"fmt"

	"github.com/sensepost/gowitness/internal/ascii"
	"github.com/sensepost/gowitness/pkg/creds"
	"github.com/sensepost/gowitness/pkg/database"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/models"
	"github.com/spf13/cobra"
	"gorm.io/gorm/clause"
)

var credsCmdFlags = struct {
	DbURI    string
	JsonFile string
}{}
var credsCmd = &cobra.Command{
	Use:   "creds",
	Short: "List sites that may have default credentials",
	Long: ascii.LogoHelp(ascii.Markdown(`
# report creds

List sites that may have default credentials.`)),
	Example: ascii.Markdown(`
- gowitness report creds
`),
	Run: func(cmd *cobra.Command, args []string) {
		log.Warn("this command is a *work in progress*.")
		log.Warn("this command is a *work in progress*.")

		var results = []*models.Result{}

		conn, err := database.Connection(credsCmdFlags.DbURI, true, false)
		if err != nil {
			log.Error("could not connect to database", "err", err)
			return
		}

		if err := conn.Model(&models.Result{}).Preload(clause.Associations).
			Find(&results).Error; err != nil {
			log.Error("could not get list", "err", err)
			return
		}

		matchCreds(results)
	},
}

func init() {
	reportCmd.AddCommand(credsCmd)

	credsCmd.Flags().StringVar(&credsCmdFlags.DbURI, "db-uri", "sqlite://gowitness.sqlite3", "The location of a gowitness database")
}

func matchCreds(results []*models.Result) {
	for _, result := range results {
		log.Debug("processing result", "url", result.URL, "tile", result.Title)

		credentials := creds.Find(result.HTML)
		if len(credentials) == 0 {
			continue
		}

		fmt.Printf("%s (%s)\n", result.URL, result.Title)

		for _, c := range credentials {
			for _, candidate := range c.Credentials {
				fmt.Printf(" - %s = %s\n", c.Name, candidate)
			}
		}
	}
}
