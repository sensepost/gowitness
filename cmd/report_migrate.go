package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/glebarez/sqlite"
	"github.com/sensepost/gowitness/internal/ascii"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/models"
	"github.com/sensepost/gowitness/pkg/models/oldv2"
	"github.com/sensepost/gowitness/pkg/writers"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var migrateCmdFlags = struct {
	Source string
}{}
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate a gowitness v2 SQLite database to v3",
	Long: ascii.LogoHelp(ascii.Markdown(`
# report migrate

Migrate a gowitness v2 SQLite database to v3.

Given a source gowitness v2 SQLite database, this command will read and map data
to the appropriate v3 structure, writing results to a new database file. The new
database file will be next to the source file, titled _*-v3-migrated.sqlite3_.

Naturally, not all fields that exist in a v3 database will be in a v2 database,
and as a result will remain empty.`)),
	Example: ascii.Markdown(`
- gowitness report migrate -s ~/gowitnessv2.sqlite3`),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if migrateCmdFlags.Source == "" {
			return errors.New("a source must be specified")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		source, err := gorm.Open(sqlite.Open(migrateCmdFlags.Source), &gorm.Config{})
		if err != nil {
			log.Error("could not open source database", "err", err)
			return
		}

		if !isV2Database(source) {
			log.Error("the source database does not appear to be a gowitness v2 database. table 'urls' not found")
			return
		}

		targetFile := fmt.Sprintf("%s.v3-migrated.sqlite3",
			strings.TrimSuffix(migrateCmdFlags.Source, filepath.Ext(migrateCmdFlags.Source)))
		log.Info("writing to new SQLite database file", "target", targetFile)

		writer, err := writers.NewDbWriter(fmt.Sprintf("sqlite://%s", targetFile), false)
		if err != nil {
			log.Error("could not open new database", "err", err)
			return
		}

		// read URLs from the source database, and write to the destination
		var v2URLs []oldv2.URL
		source.Preload("Headers").
			Preload("Technologies").
			Preload("TLS.TLSCertificates.DNSNames").
			Preload("Console").
			Preload("Network").
			Find(&v2URLs)
		log.Info("total URLs to process", "total", len(v2URLs))

		for _, v2URL := range v2URLs {
			v3result := mapV2ToV3(v2URL)

			if err := writer.Write(&v3result); err != nil {
				log.Error("could not write v2 URL result to v3 database", "err", err)
			}
		}

		log.Info("database migrated")
	},
}

func init() {
	reportCmd.AddCommand(migrateCmd)

	migrateCmd.Flags().StringVarP(&migrateCmdFlags.Source, "source", "s", "", "A gowitness v2 SQLite database file to migrate to v3")
}

// isV2Database checks if the 'urls' table exists in the source database to verify it's a v2 database
func isV2Database(db *gorm.DB) bool {
	var count int64
	err := db.Table("urls").Count(&count).Error
	return err == nil
}

// mapV2ToV3 maps the v2 URL data to a fully populated models.Result (v3)
func mapV2ToV3(v2URL oldv2.URL) models.Result {
	return models.Result{
		URL:            v2URL.URL,
		FinalURL:       v2URL.FinalURL,
		ResponseCode:   v2URL.ResponseCode,
		ResponseReason: v2URL.ResponseReason,
		Protocol:       v2URL.Proto,
		ContentLength:  v2URL.ContentLength,
		Title:          v2URL.Title,
		PerceptionHash: v2URL.PerceptionHash,
		Filename:       v2URL.Filename,
		IsPDF:          v2URL.IsPDF,
		ProbedAt:       v2URL.CreatedAt,
		HTML:           v2URL.DOM,
		Screenshot:     v2URL.Screenshot,

		TLS:          convertTLS(v2URL.TLS),
		Technologies: convertTechnologies(v2URL.Technologies),
		Headers:      convertHeaders(v2URL.Headers),
		Network:      convertNetworkLogs(v2URL.Network),
		Console:      convertConsoleLogs(v2URL.Console),

		// these fields that don't exist in v2
		Failed:       false,
		FailedReason: "",
	}
}

func convertTLS(v2TLS oldv2.TLS) models.TLS {
	return models.TLS{
		Protocol: strconv.Itoa(int(v2TLS.Version)),
		Issuer:   v2TLS.ServerName,
		SanList:  convertTLSSanList(v2TLS.TLSCertificates),
	}
}

func convertTLSSanList(v2Certs []oldv2.TLSCertificate) []models.TLSSanList {
	var sanList []models.TLSSanList
	for _, cert := range v2Certs {
		for _, dnsName := range cert.DNSNames {
			sanList = append(sanList, models.TLSSanList{Value: dnsName.Name})
		}
	}
	return sanList
}

func convertTechnologies(v2Technologies []oldv2.Technologie) []models.Technology {
	var v3Technologies []models.Technology
	for _, tech := range v2Technologies {
		v3Technologies = append(v3Technologies, models.Technology{
			Value: tech.Value,
		})
	}
	return v3Technologies
}

func convertHeaders(v2Headers []oldv2.Header) []models.Header {
	var v3Headers []models.Header
	for _, header := range v2Headers {
		v3Headers = append(v3Headers, models.Header{
			Key:   header.Key,
			Value: header.Value,
		})
	}
	return v3Headers
}

func convertNetworkLogs(v2Logs []oldv2.NetworkLog) []models.NetworkLog {
	var v3Logs []models.NetworkLog
	for _, log := range v2Logs {
		v3Logs = append(v3Logs, models.NetworkLog{
			RequestType: models.RequestType(log.RequestType),
			StatusCode:  log.StatusCode,
			URL:         log.URL,
			RemoteIP:    log.IP, // IP field is now RemoteIP in v3
			Time:        log.Time,
		})
	}
	return v3Logs
}

func convertConsoleLogs(v2Logs []oldv2.ConsoleLog) []models.ConsoleLog {
	var v3Logs []models.ConsoleLog
	for _, log := range v2Logs {
		v3Logs = append(v3Logs, models.ConsoleLog{
			Type:  log.Type,
			Value: log.Value,
		})
	}
	return v3Logs
}
