package cmd

import (
	"fmt"
	"html/template"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sensepost/gowitness/lib"
	"github.com/sensepost/gowitness/storage"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var (
	rsDB *gorm.DB
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts a webserver that serves the report interface, api and screenshot tool",
	Long: `Starts a webserver that serves the report interface, api and screenshot tool.

The global database and screenshot paths should be set to the same as
what they were when a scan was run. The report server also has the ability
to screenshot ad-hoc URLs provided to the submission page.

The API is usable to take screenshots and reflect them back amongst other useful things.
Most of the Gowitness core is exposed via the API.

NOTE: When changing the server address to something other than localhost, make 
sure that only authorised connections can be made to the server port. By default,
access is restricted to localhost to reduce the risk of SSRF attacks against the
host or hosting infrastructure (AWS/Azure/GCP, etc). Consider strict IP filtering
or fronting this server with an authentication aware reverse proxy.

Allowed URLs, by default, need to start with http:// or https://. If you need
this restriction lifted, add the --allow-insecure-uri / -A flag. A word of 
warning though, that also means that someone may request a URL like file:///etc/passwd.

Assuming the server is hosted on localhost, an HTTP GET request to
take a screenshot of google.com would be:
	http://localhost:7171/?url=https://www.google.com`,
	Example: `$ gowitness server
$ gowitness server --addr 0.0.0.0:8080`,
	Run: func(cmd *cobra.Command, args []string) {
		log := options.Logger

		if !strings.Contains(options.ServerAddr, "localhost") {
			log.Warn().Msg("exposing this server to other networks is dangerous! see the server command help for more information")
		}

		// db
		dbh, err := db.Get()
		if err != nil {
			log.Fatal().Err(err).Msg("could not gt db handle")
		}
		rsDB = dbh

		log.Info().Str("path", db.Path).Msg("db path")
		log.Info().Str("path", options.ScreenshotPath).Msg("screenshot path")

		if options.Debug {
			gin.SetMode(gin.DebugMode)
		} else {
			gin.SetMode(gin.ReleaseMode)
		}

		r := gin.Default()

		tmpl := template.Must(template.New("").ParseFS(Embedded, "web/templates/*.html"))
		r.SetHTMLTemplate(tmpl)

		// routes
		r.GET("/", dashboardHandler)
		r.GET("/gallery", galleryHandler)
		r.GET("/table", tableHandler)
		r.GET("/details/:id", detailHandler)
		r.GET("/submit", getSubmitHandler)
		r.POST("/submit", submitHandler)

		// static assets & screenshots
		assetFs, err := fs.Sub(Embedded, "web/assets")
		if err != nil {
			log.Fatal().Err(err).Msg("could not fs.Sub Assets")
		}

		// assets & screenshots
		r.StaticFS("/assets/", http.FS(assetFs))
		r.StaticFS("/screenshots", http.Dir(options.ScreenshotPath))

		// API
		api := r.Group("/api")
		{
			api.GET("/", handler)
		}

		log.Info().Str("address", options.ServerAddr).Msg("server listening")
		if err := r.Run(options.ServerAddr); err != nil {
			log.Fatal().Err(err).Msg("webserver failed")
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringVarP(&options.ServerAddr, "address", "a", "localhost:7171", "server listening address")
	serverCmd.Flags().BoolVarP(&options.AllowInsecureURIs, "allow-insecure-uri", "A", false, "allow uris that dont start with http(s)")
}

// dashboardHandler handles dashboard requests
func dashboardHandler(c *gin.Context) {

	// get the sqlite db size
	var size int64
	rsDB.Raw("SELECT page_count * page_size as size FROM pragma_page_count(), pragma_page_size();").Take(&size)

	// count some statistics

	var urlCount int64
	rsDB.Model(&storage.URL{}).Count(&urlCount)

	var certCount int64
	rsDB.Model(&storage.TLS{}).Count(&certCount)

	var certDNSNameCount int64
	rsDB.Model(&storage.TLSCertificateDNSName{}).Count(&certDNSNameCount)

	var headerCount int64
	rsDB.Model(&storage.Header{}).Count(&headerCount)

	var techCount int64
	rsDB.Model(&storage.Technologie{}).Distinct().Count(&techCount)

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"DBSzie":       fmt.Sprintf("%.2f", float64(size)/1e6),
		"URLCount":     urlCount,
		"CertCount":    certCount,
		"DNSNameCount": certDNSNameCount,
		"HeaderCount":  headerCount,
		"TechCount":    techCount,
	})
}

// getSubmitHandler handles generating the view to submit urls
func getSubmitHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "submit.html", nil)
}

// submitHandler handles url submissions
func submitHandler(c *gin.Context) {

	// prepare target
	url, err := url.Parse(strings.TrimSpace(c.PostForm("url")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	if !options.AllowInsecureURIs {
		if !strings.HasPrefix(url.Scheme, "http") {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "only http(s) urls are accepted",
			})
			return
		}
	}

	fn := lib.SafeFileName(url.String())
	fp := lib.ScreenshotPath(fn, url, options.ScreenshotPath)

	preflight, err := chrm.Preflight(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	var rid uint
	if rsDB != nil {
		if rid, err = chrm.StorePreflight(rsDB, preflight, fn); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
	}

	buf, err := chrm.Screenshot(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	if err := ioutil.WriteFile(fp, buf, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	if rid > 0 {
		c.Redirect(http.StatusMovedPermanently, "/details/"+strconv.Itoa(int(rid)))
		return
	}

	c.Redirect(http.StatusMovedPermanently, "/submit")
}

// detailHandler gets all of the details for a particular url id
func detailHandler(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	var url storage.URL
	rsDB.
		Preload("Headers").
		Preload("TLS").
		Preload("TLS.TLSCertificates").
		Preload("TLS.TLSCertificates.DNSNames").
		Preload("Technologies").
		First(&url, id)

	c.HTML(http.StatusOK, "detail.html", gin.H{
		"Data": url,
	})
}

// tableHandler handles the URL table view
func tableHandler(c *gin.Context) {

	var urls []storage.URL
	rsDB.Find(&urls)

	c.HTML(http.StatusOK, "table.html", gin.H{
		"Data": urls,
	})
}

// galleryHandler handles the index page. this is the main gallery view
func galleryHandler(c *gin.Context) {

	currPage, limit, err := getPageLimit(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	pager := &lib.Pagination{
		DB:       rsDB,
		CurrPage: currPage,
		Limit:    limit,
	}

	// perception hashing
	if strings.TrimSpace(c.Query("perception_sort")) == "true" {
		pager.OrderBy = []string{"perception_hash desc"}
	}

	// search
	if strings.TrimSpace(c.Query("search")) != "" {
		pager.FilterBy = append(pager.FilterBy, lib.Filter{
			Column: "title",
			Value:  c.Query("search"),
		})
	}

	var urls []storage.URL
	page, err := pager.Page(&urls)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "gallery.html", gin.H{
		"Data": page,
	})
}

// getPageLimit gets the limit and page query string values from a request
func getPageLimit(c *gin.Context) (page int, limit int, err error) {

	pageS := strings.TrimSpace(c.Query("page"))
	limitS := strings.TrimSpace(c.Query("limit"))

	if pageS == "" {
		pageS = "-1"
	}
	if limitS == "" {
		limitS = "0"
	}

	page, err = strconv.Atoi(pageS)
	if err != nil {
		return
	}
	limit, err = strconv.Atoi(limitS)
	if err != nil {
		return
	}

	return
}

// handler is the HTTP handler for the web service this command exposes
func handler(c *gin.Context) {
	rawURL := strings.TrimSpace(c.Query("url"))
	if rawURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "url parameter missing. eg ?url=https://google.com",
		})
		return
	}

	url, err := url.Parse(rawURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	if !options.AllowInsecureURIs {
		if !strings.HasPrefix(url.Scheme, "http") {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "only http(s) urls are accepted",
			})
			return
		}
	}

	buf, err := chrm.Screenshot(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.Data(http.StatusOK, "image/png", buf)
}
