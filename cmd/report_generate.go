package cmd

import (
	"github.com/sensepost/gowitness/storage"
	"github.com/spf13/cobra"
	"html/template"
	"os"
)

const outputTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <!--    Gallery example code https://www.imarketinx.de/artikel/responsive-image-gallery-with-css-grid.html-->
    <style>
        * { box-sizing: border-box;}

    html {
      font-size: 100%;
    }

    body {
      padding: 1rem;
      font-familiy: Verdana, Geneva, sans-serif;
    }

    h2 {
      font-size: 1.5rem;
      font-weight: bold;
      line-height: 1.5;
    }

    p {
      margin: .5rem 0;
      font-size: 1.25rem;
      line-height: 1.5;
    }

    /* First the Grid */
    .gallery-grid {
      display: -ms-grid;
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(500px, 1fr));
      grid-gap: 1.5rem;
      justify-items: center;
      margin: 0;
      padding: 0;
    }

    /* The Picture Frame */
    .gallery-frame {
      padding: .5rem;
      font-size: 1.2rem;
      text-align: center;
      background-color: #333;
      color: #d9d9d9;
    }

    /* The Images */
    .gallery-img {
      max-width: 100%;
      height: auto;
      object-fit: cover;
      transition: opacity 0.25s ease-in-out;
    }

    .gallery-img:hover {
      opacity: .7;
    }
    </style>
</head>

<body>
<div class="gallery-grid">
    {{ range .Data }}
    <figure class="gallery-frame">
        <img class="gallery-img" src="{{ $.ScreenshotPath }}/{{ .Filename }}" alt="{{ .Title }}" title="{{ .Title }}">
        <figcaption>{{ .FinalURL }}</figcaption>
    </figure>
    {{ end }}
</div>
</body>
</html>
`

// reportGenerateCmd represents the reportGenerate command
var reportGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates a static HTML report which presents entries in the gowitness database and their screenshots",
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

		parsedTemplate, err := template.New("template").Parse(outputTemplate)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to read template")
		}

		f, err := os.Create("report.html")
		if err != nil {
			log.Fatal().Err(err).Msg("could not create a new file, report.html")
		}

		templateData := make(map[string]interface{})
		templateData["Data"] = data
		templateData["ScreenshotPath"] = options.ScreenshotPath

		err = parsedTemplate.Execute(f, templateData)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to execute template")
		}

	},
}

func init() {
	reportCmd.AddCommand(reportGenerateCmd)
	reportGenerateCmd.Flags().BoolVarP(&options.PerceptionSort, "sort", "S", false, "sort by image perceptions")
}
