package cmd

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/sensepost/gowitness/lib"
	"github.com/sensepost/gowitness/storage"
	"github.com/sensepost/gowitness/web"
	"github.com/shurcooL/httpfs/html/vfstemplate"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

type ApiRequest struct {
	Name string
	Url  string
}

var (
	tmpl *template.Template
	rsDB *gorm.DB
)

// reportServeCmd represents the reportServe command
var reportServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "starts a web server to view screenshot reports",
	Long: `Starts a web server to view screenshot reports.

The global database and screenshot paths should be set to the same as
what they were when a scan was run. The report server also has the ability
to screenshot ad-hoc URLs provided to the submission page.

NOTE: When changing the server address to something other than localhost, make 
sure that only authorised connections can be made to the server port. By default,
access is restricted to localhost to reduce the risk of SSRF attacks against the
host or hosting infrastructure (AWS/Azure/GCP, etc). Consider strict IP filtering
or fronting this server with an authentication aware reverse proxy.

Allowed URLs, by default, need to start with http:// or https://. If you need
this restriction lifted, add the --allow-insecure-uri / -A flag. A word of 
warning though, that also means that someone may request a URL like file:///etc/passwd.
`,
	Run: func(cmd *cobra.Command, args []string) {
		log := options.Logger

		if !strings.Contains(options.ServerAddr, "localhost") {
			log.Warn().Msg("exposing this server to other networks is dangerous! see the report serve command help for more information")
		}

		tmpl = template.Must(vfstemplate.ParseGlob(web.Assets, nil, "templates/*.html"))

		// db
		dbh, err := db.Get()
		if err != nil {
			log.Fatal().Err(err).Msg("could not gt db handle")
		}
		rsDB = dbh

		log.Info().Str("path", db.Path).Msg("db path")
		log.Info().Str("path", options.ScreenshotPath).Msg("screenshot path")

		// routes
		// messing with the trailing /'s breaks routing in confusing ways :<
		http.HandleFunc("/", indexHandler)
		http.HandleFunc("/table/", tableHandler)
		http.HandleFunc("/details", detailHandler)
		http.HandleFunc("/submit", submitHandler)
		http.HandleFunc("/api", submitHandler)

		// static
		http.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(web.Assets)))
		http.Handle("/screenshots/", http.StripPrefix("/screenshots", http.FileServer(http.Dir(options.ScreenshotPath))))

		log.Info().Str("address", options.ServerAddr).Msg("server listening")
		if err := http.ListenAndServe(options.ServerAddr, nil); err != nil {
			log.Fatal().Err(err).Msg("webserver failed")
		}
	},
}

func init() {
	reportCmd.AddCommand(reportServeCmd)

	reportServeCmd.Flags().StringVarP(&options.ServerAddr, "address", "a", "localhost:7171", "server listening address")
	reportServeCmd.Flags().BoolVarP(&options.AllowInsecureURIs, "allow-insecure-uri", "A", false, "allow uris that dont start with http(s)")
}

func takeScreenshot(rUrl string, fn string) (uint, error) {
	url, err := url.Parse(rUrl)
	if err != nil {
		return 0, err
	}

	if fn == "" {
		fn = lib.SafeFileName(url.String())
	} else if !strings.HasSuffix(fn, ".png") {
		fn = fn + ".png"
	}
	fp := lib.ScreenshotPath(fn, url, options.ScreenshotPath)

	resp, title, err := chrm.Preflight(url)
	if err != nil {
		return 0, err
	}

	var rid uint
	if rsDB != nil {
		if rid, err = chrm.StorePreflight(url, rsDB, resp, title, fn); err != nil {
			return 0, err
		}
	}

	buf, err := chrm.Screenshot(url)
	if err != nil {
		return 0, err
	}

	if err := ioutil.WriteFile(fp, buf, 0644); err != nil {
		return 0, err
	}

	return rid, nil
}

// submitHandler handles url submissions
func submitHandler(w http.ResponseWriter, r *http.Request) {
	// check content-type
	ct := r.Header.Get("Content-Type")
	isApi := ct == "application/json"

	switch r.Method {
	case "GET":
		if isApi {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"success": false, "message": "only POST requests are accepted"}`))
			return
		} else {
			t := tmpl.Lookup("submit.html")
			err := t.ExecuteTemplate(w, "submit", nil)
			if err != nil {
				panic(err)
			}
		}

	case "POST":
		var rUrl string
		var fn string
		if isApi {
			var p ApiRequest

			err := json.NewDecoder(r.Body).Decode(&p)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf(`{"success": false, "message": "%s"}`, err.Error())))
				return
			}

			rUrl = p.Url
			fn = p.Name
		} else {
			rUrl = strings.TrimSpace(r.FormValue("url"))
		}

		// prepare target
		url, err := url.Parse(rUrl)
		if err != nil {
			if isApi {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf(`{"success": false, "message": "%s"}`, err.Error())))
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		if !options.AllowInsecureURIs {
			if !strings.HasPrefix(url.Scheme, "http") {
				if isApi {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNotAcceptable)
					w.Write([]byte(`{"success": false, "message": "only http(s) urls are accepted"}`))
				} else {
					http.Error(w, "only http(s) urls are accepted", http.StatusNotAcceptable)
				}
				return
			}
		}

		if isApi {
			go takeScreenshot(rUrl, fn)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true}`))
			return
		} else {
			rid, err := takeScreenshot(rUrl, "")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if rid > 0 {
				http.Redirect(w, r, "/details?id="+strconv.Itoa(int(rid)), http.StatusMovedPermanently)
				return
			}

			http.Redirect(w, r, "/submit", http.StatusMovedPermanently)
		}
	}
}

// detailHandler gets all of the details for a particular url id
func detailHandler(w http.ResponseWriter, r *http.Request) {

	d := strings.TrimSpace(r.URL.Query().Get("id"))
	if d == "" {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}
	id, err := strconv.Atoi(d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var url storage.URL
	rsDB.
		Preload("Headers").
		Preload("TLS").
		Preload("TLS.TLSCertificates").
		Preload("TLS.TLSCertificates.DNSNames").
		First(&url, id)

	// fmt.Printf("%+v\n", url)

	t := tmpl.Lookup("detail.html")
	err = t.ExecuteTemplate(w, "detail", url)
	if err != nil {
		panic(err)
	}
}

// tableHandler handles the URL table view
func tableHandler(w http.ResponseWriter, r *http.Request) {

	var urls []storage.URL
	rsDB.Find(&urls)

	t := tmpl.Lookup("table.html")
	err := t.ExecuteTemplate(w, "table", urls)
	if err != nil {
		panic(err)
	}
}

// indexHandler handles the index page. this is the main gallery view
func indexHandler(w http.ResponseWriter, r *http.Request) {

	currPage, limit, err := getPageLimit(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pager := &lib.Pagination{
		DB:       rsDB,
		CurrPage: currPage,
		Limit:    limit,
	}

	// perception hashing
	if strings.TrimSpace(r.URL.Query().Get("perception_sort")) == "true" {
		pager.OrderBy = []string{"perception_hash desc"}
	}

	// search
	if strings.TrimSpace(r.URL.Query().Get("search")) != "" {
		pager.FilterBy = append(pager.FilterBy, lib.Filter{
			Column: "title",
			Value:  r.URL.Query().Get("search"),
		})
	}

	var urls []storage.URL
	page, err := pager.Page(&urls)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// fmt.Printf("%+v\n", currPage)

	t := tmpl.Lookup("gallery.html")
	err = t.ExecuteTemplate(w, "gallery", page)
	if err != nil {
		panic(err)
	}
}

// getPageLimit gets the limit and page query string values from a request
func getPageLimit(r *http.Request) (page int, limit int, err error) {

	pageS := strings.TrimSpace(r.URL.Query().Get("page"))
	limitS := strings.TrimSpace(r.URL.Query().Get("limit"))

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
