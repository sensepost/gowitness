package cmd

import (
	"archive/zip"
	"bytes"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sensepost/gowitness/storage"
	"github.com/spf13/cobra"
)

// reportExportCmd represents the reportExport command
var reportExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export an HTML report of all screenshots",
	Long: `Export an HTML report of all screenshots.

The exported report will be a completely offline viewable (as in, there is
no need for the embedded report server). The file name that needs to be
specified will be the target for the final ZIP file that will be created.
`,
	Run: func(cmd *cobra.Command, args []string) {
		log := options.Logger

		if options.File == "" {
			log.Fatal().Msg("a target file must be specified with --file / -f")
		}

		tmpl := template.Must(template.ParseFS(Embedded, "web/static-templates/*.html"))
		t := tmpl.Lookup("gallery.html")

		// db
		dbh, err := db.Get()
		if err != nil {
			log.Fatal().Err(err).Msg("could not get db handle")
		}

		var urls []storage.URL
		dbh.Preload("Technologies").Find(&urls)

		if len(urls) == 0 {
			log.Error().Msg("there are no urls in the database")
			return
		}

		// create a temp working directory
		tempDir, err := os.MkdirTemp("", "")
		if err != nil {
			log.Fatal().Err(err).Msg("could not create temp working directory")
		}
		defer os.RemoveAll(tempDir)
		log.Debug().Str("tempdir", tempDir).Msg("created temp working directory")

		// create the screenshots directory
		if err := os.MkdirAll(tempDir+"/screenshots", 0755); err != nil {
			log.Fatal().Err(err).Msg("could not create screenshots directory")
		}
		screenshotDir := tempDir + "/screenshots"

		log.Info().Msg("preparing screenshots and other assets")

		// copy screenshot files to the new path
		if err := filepath.Walk(options.ScreenshotPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			dest := screenshotDir + "/" + info.Name()
			log.Debug().Str("destination", dest).Str("source", path).Msg("copying screenshot file to temp directory")

			if err := copyFile(path, dest); err != nil {
				return err
			}

			return nil

		}); err != nil {
			log.Fatal().Str("screenshot-path", options.ScreenshotPath).Err(err).Msg("could not walk screenshot path and copy screenshots")
		}

		// copy template css & js to new path
		js, err := Embedded.ReadFile("web/assets/js/tabler.min.js")
		if err != nil {
			log.Fatal().Err(err).Msg("could not open embedded javascript")
		}
		if err := copyByte(js, tempDir+"/"+"tabler.min.js"); err != nil {
			log.Fatal().Err(err).Msg("failed to write tabler.min.js to file")
		}

		css, err := Embedded.ReadFile("web/assets/css/tabler.min.css")
		if err != nil {
			log.Fatal().Err(err).Msg("could not open embedded css")
		}
		if err := copyByte(css, tempDir+"/"+"tabler.min.css"); err != nil {
			log.Fatal().Err(err).Msg("failed to write tabler.min.css to file")
		}

		// write and generate the template
		f, err := os.Create(tempDir + "/index.html")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create index.html")
		}

		if err := t.Execute(f, urls); err != nil {
			log.Fatal().Err(err).Msg("failed to generate template")
		}

		// create the destination archive
		target, err := os.Create(options.File)
		if err != nil {
			log.Fatal().Err(err).Str("target-file", options.File).Msg("failed to open target file")
		}
		defer target.Close()

		log.Info().Str("target-file", options.File).Msg("writing zip archive")

		// create the zip file...
		w := zip.NewWriter(target)
		defer w.Close()
		// .. and add the files to it
		if err := filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			// create the file in the archive
			relPath := "gowitness/" + strings.TrimPrefix(path, tempDir+"/")
			df, err := w.Create(relPath)
			if err != nil {
				return err
			}

			_, err = io.Copy(df, f)
			if err != nil {
				return err
			}

			return nil

		}); err != nil {
			log.Fatal().Str("target-file", options.File).Err(err).Msg("failed to create report archive")
		}

		log.Info().Msg("done")
	},
}

func init() {
	reportCmd.AddCommand(reportExportCmd)

	reportExportCmd.Flags().StringVarP(&options.File, "file", "f", "", "the target file to save the zipped report to")

	cobra.MarkFlagRequired(reportExportCmd.Flags(), "file")
}

// copyFile well, copies a file
func copyFile(from string, to string) error {

	s, err := os.Open(from)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.Create(to)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, s)
	if err != nil {
		return err
	}

	return nil
}

// copyByte copies bytes to a file
func copyByte(b []byte, to string) error {
	d, err := os.Create(to)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, bytes.NewReader(b))
	if err != nil {
		return err
	}

	return nil
}
