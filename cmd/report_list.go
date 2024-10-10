package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/x/term"
	"github.com/sensepost/gowitness/internal/ascii"
	"github.com/sensepost/gowitness/pkg/database"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/models"
	"github.com/spf13/cobra"
	"gorm.io/gorm/clause"
)

var listCmdFlags = struct {
	DbURI    string
	JsonFile string
}{}
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List a summary of results from a data source",
	Long: ascii.LogoHelp(ascii.Markdown(`
# report list

List a summary of results from a data source, like an SQLite database or a JSON
lines file.`)),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if listCmdFlags.DbURI == "" && listCmdFlags.JsonFile == "" {
			return errors.New("no data source defined")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var results = []*models.Result{}

		// if we have a json path, use that
		if listCmdFlags.JsonFile != "" {
			file, err := os.Open(listCmdFlags.JsonFile)
			if err != nil {
				log.Error("could not open JSON Lines file", "err", err)
				return
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
						log.Error("error reading JSON Lines file", "err", err)
						return
					}
				}

				var result models.Result
				if err := json.Unmarshal(line, &result); err != nil {
					log.Error("could not unmarshal JSON line", "err", err)
					continue
				}
				results = append(results, &result)

				if err == io.EOF {
					break
				}
			}

			renderTable(results)
			return
		}

		// db-uri is the default
		conn, err := database.Connection(listCmdFlags.DbURI, true, false)
		if err != nil {
			log.Error("could not connect to database", "err", err)
			return
		}

		if err := conn.Model(&models.Result{}).Preload(clause.Associations).Find(&results).Error; err != nil {
			log.Error("could not get list", "err", err)
			return
		}

		renderTable(results)
	},
}

func init() {
	reportCmd.AddCommand(listCmd)

	listCmd.Flags().StringVar(&listCmdFlags.DbURI, "db-uri", "sqlite://gowitness.sqlite3", "The location of a gowitness database")
	listCmd.Flags().StringVar(&listCmdFlags.JsonFile, "json-file", "", "The location of a JSON Lines results file (e.g., ./gowitness.jsonl). This flag takes precedence over --db-uri")
}

func renderTable(results []*models.Result) {
	PaddedStyle := lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1)
	HeaderStyle := PaddedStyle.Bold(true).Underline(true)
	RowStyle := PaddedStyle

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
		Headers(
			"When", "Failed", "Code", "Input URL", "Title", "~Size",
			"Net", "Con", "Header", "Cookie",
		).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == 0:
				return HeaderStyle
			default:
				return RowStyle
			}
		})

	for _, result := range results {
		t.Row(
			result.ProbedAt.Format("Jan 2 15:04:05"),
			failedStyle(result.Failed),
			statusCode(result.ResponseCode),
			urlStyle(result.URL),
			titleStyle(result.Title),
			fmt.Sprintf("%dkb", result.ContentLength/1024),
			fmt.Sprintf("%d", len(result.Network)),
			fmt.Sprintf("%d", len(result.Console)),
			fmt.Sprintf("%d", len(result.Headers)),
			fmt.Sprintf("%d", len(result.Cookies)),
		)
	}

	w, _, _ := term.GetSize(os.Stdout.Fd())
	fmt.Println(lipgloss.NewStyle().MaxWidth(w).Render(t.String()))
}

func statusCode(code int) string {
	var style lipgloss.Style

	switch {
	case code >= 200 && code < 300:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("42")) // Green
	case code >= 300 && code < 400:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("220")) // Yellow
	case code >= 400 && code < 500:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("208")) // Orange
	case code >= 500:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // Red
	default:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("7")) // Light gray
	}

	return style.Render(fmt.Sprintf("%d", code))
}

func truncate(s string, maxLength int) string {
	if len(s) > maxLength {
		return s[:maxLength] + "..."
	}
	return s
}

func failedStyle(s bool) string {
	var color lipgloss.Color
	var value string

	if s {
		color = lipgloss.Color("196")
		value = "true"
	} else {
		color = lipgloss.Color("42")
		value = "false"
	}

	return lipgloss.NewStyle().Foreground(color).Render(value)
}

func urlStyle(url string) string {
	return lipgloss.NewStyle().Bold(true).Render(url)
}

func titleStyle(title string) string {
	return lipgloss.NewStyle().Render(truncate(title, 30))
}
