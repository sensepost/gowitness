package cmd

import (
	"html/template"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/sensepost/gowitness/lib"
	"github.com/sensepost/gowitness/storage"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var (
	tmpl *template.Template
	rsDB *gorm.DB
	TagFilterMap   map[interface{}]interface{}
	RespCodeFilter map[string]bool
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

		tmpl = template.Must(template.ParseFS(Templates, "web/templates/*.html"))

		// db
		dbh, err := db.Get()
		if err != nil {
			log.Fatal().Err(err).Msg("could not get db handle")
		}
		rsDB = dbh

		// initialize the filter DB values
		initializeFilterValues(rsDB)

		log.Info().Str("path", db.Path).Msg("db path")
		log.Info().Str("path", options.ScreenshotPath).Msg("screenshot path")

		// routes
		// messing with the trailing /'s breaks routing in confusing ways :<
		http.HandleFunc("/", indexHandler)
		http.HandleFunc("/table/", tableHandler)
		http.HandleFunc("/details", detailHandler)
		http.HandleFunc("/submit", submitHandler)
		http.HandleFunc("/updaterecord", updateRecordHandler)
		http.HandleFunc("/updatetags", updateTagValues)

		// static assets & screenshots
		assetFs, err := fs.Sub(Assets, "web")
		if err != nil {
			log.Fatal().Err(err).Msg("could not fs.Sub Assets")
		}
		// assetsFs := http.FileServer(http.FS(Assets))
		http.Handle("/assets/", http.FileServer(http.FS(assetFs)))
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

// submitHandler handles url submissions
func submitHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		t := tmpl.Lookup("submit.html")
		err := t.ExecuteTemplate(w, "submit", nil)
		if err != nil {
			panic(err)
		}
	case "POST":
		// prepare target
		url, err := url.Parse(strings.TrimSpace(r.FormValue("url")))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !options.AllowInsecureURIs {
			if !strings.HasPrefix(url.Scheme, "http") {
				http.Error(w, "only http(s) urls are accepted", http.StatusNotAcceptable)
				return
			}
		}

		fn := lib.SafeFileName(url.String())
		fp := lib.ScreenshotPath(fn, url, options.ScreenshotPath)

		resp, title, technologies, err := chrm.Preflight(url)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var rid uint
		if rsDB != nil {
			if rid, err = chrm.StorePreflight(url, rsDB, resp, title, technologies, fn); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		buf, err := chrm.Screenshot(url)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := ioutil.WriteFile(fp, buf, 0644); err != nil {
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
		Preload("Technologies").
		Preload("Filter").
		Preload("Filter.Tagmaps").
		First(&url, id)

	// populate the selected tags off of the url's tagmap ids
	for k, _ := range TagFilterMap{
		TagFilterMap[k] = 0
	}
	for _, tagid := range url.Filter.Tagmaps{
		for key, _ := range TagFilterMap{
			if key.(storage.Tag).ID == tagid.TagID{
				TagFilterMap[key] = 1
				break
			}
		}
	}

	// return struct for two data types
	detailTemplateHelper := DetailTemplateHelper{
		Url:	&url,
		Tags:	&TagFilterMap,
	}

	t := tmpl.Lookup("detail.html")
	err = t.ExecuteTemplate(w, "detail", detailTemplateHelper)
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

	// filter checks to perform joins
	attemptedTagFilter := false
	attemptedFilter := false
	
	// perception hashing
	if strings.TrimSpace(r.URL.Query().Get("perception_sort")) == "true" {
		pager.OrderBy = []string{"perception_hash desc"}
	}

	// visibility filter, this statement requires that every URL already have a filter defined
	vis := strings.TrimSpace(r.URL.Query().Get("hide"))
	if  vis == "on" || vis == "true" {
		pager.FilterBy = append(pager.FilterBy, lib.Filter{
			Column: "filters.visible",
			Value:  "1",
			Oper:	"=",
		})

		attemptedFilter = true
	}

	// Notes Filtering
	onlyshownotes := strings.TrimSpace(r.URL.Query().Get("onlynotes"))
	if onlyshownotes == "on" || onlyshownotes == "true"{
		pager.FilterBy = append(pager.FilterBy, lib.Filter{
			Column: "filters.notes",
			Value:  "",
			Oper:	"<>",
		})
		attemptedFilter = true
	}

	// clear the selected tags
	for k, _ := range TagFilterMap{
		TagFilterMap[k] = 0
	}

	tags := r.URL.Query()["tag"]
	// find the matches from the DB
	var tagmatchids []uint
	for _, tag := range tags{
		for key, _ := range TagFilterMap{
			if key.(storage.Tag).Color == tag{
				tagmatchids = append(tagmatchids,uint(key.(storage.Tag).ID))
				TagFilterMap[key] = 1
			}
		}
	}

	// signal the tagmap joins
	if len(tagmatchids) > 0{
		attemptedTagFilter = true
	}

	// Setup the Filter Table JOINS to do Intersect queries
	if attemptedFilter || attemptedTagFilter{
		pager.JoinsBy = append(pager.JoinsBy, lib.Filter{
			Column: "filters",
			Value:	"filters.url_id = urls.id",
		})
		pager.FilterBy = append(pager.FilterBy, lib.Filter{
			Column: "filters.deleted_at",
			Value:  nil,
			Oper:	"IS",
		})
	}

	// Setup the Tagmap Table JOINS to do Intersect queries
	if attemptedTagFilter{
		pager.JoinsBy = append(pager.JoinsBy, lib.Filter{
			Column: "tagmaps",
			Value:	"tagmaps.filter_id = filters.id",
		})
		pager.FilterBy = append(pager.FilterBy, lib.Filter{
			Column: "tagmaps.tag_id",
			Value:  tagmatchids,
			Oper:	"IN",
		})
		pager.FilterBy = append(pager.FilterBy, lib.Filter{
			Column: "tagmaps.deleted_at",
			Value:  nil,
			Oper:	"IS",
		})
	}

	// clear the response code filters
	for k, _ := range RespCodeFilter{
		RespCodeFilter[k] = false
	}

	// HTTP Response Code Filtering
	codes := r.URL.Query()["code"]
	var thecodes []string
	if len(codes) > 0{
		for _, code := range codes {
			thecodes = append(thecodes,strings.TrimSpace(code))
			RespCodeFilter[code] = true
		}

		pager.FilterBy = append(pager.FilterBy, lib.Filter{
			Column: "urls.response_code",
			Value:  thecodes,
			Oper:	"IN",
		})
	}

	// Search Filtering
	if strings.TrimSpace(r.URL.Query().Get("search")) != "" {
		pager.FilterBy = append(pager.FilterBy, lib.Filter{
			Column: "urls.title",
			Value:  r.URL.Query().Get("search"),
			Oper:	"LIKE",
		})
	}

	// Get the page data
	var urls []storage.URL
	page, err := pager.Page(&urls)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Additional page modifications for filtering. "on" is for the default checkbox
	// value before the 'true' is added via the filter handler
	if vis == "on" || vis == "true"{
		page.ShowHidden = true
	}

	if onlyshownotes == "on" || onlyshownotes == "true" {
		page.OnlyShowNotes = true
	}
	
	// page map references
	page.FiltTagMap = &TagFilterMap
	page.FiltRespCodes = &RespCodeFilter

	t := tmpl.Lookup("gallery.html")
	err = t.ExecuteTemplate(w, "gallery", page)
	if err != nil {
		panic(err)
	}
}

// updateRecordHandler updates a database record from the form values of gallery
func updateTagValues(w http.ResponseWriter, r *http.Request){
	switch r.Method{
		case "GET":
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
		case "POST":
			r.ParseForm()
			
			// upadate the tags in the database
			for ind, tag := range r.Form["tag"]{
				tagstr := strings.TrimSpace(tag)
				if tag != ""{
					rsDB.Model(storage.Tag{}).Table("tags").Where("id = ?",ind+1).Update("name",tagstr)
				}
			}

			// update the 'global' tags
			var dbtags []storage.Tag
			rsDB.Table("tags").Find(&dbtags)

			// remake the filter map, exiting one should be sent into gc
			TagFilterMap = make(map[interface{}]interface{})
			for _, tg := range dbtags{
				TagFilterMap[tg] = 0
			}
	
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
		}
}

// updateRecordHandler updates a database record from the form values of gallery
func updateRecordHandler(w http.ResponseWriter, r *http.Request){
	switch r.Method{
		case "GET":
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
		case "POST":
			// parse the incoming form
			r.ParseForm()

			// check for a valid ID
			id, err := strconv.Atoi(r.FormValue("id"))
			if err != nil {
				http.Error(w, "ID Not sent With Request", http.StatusInternalServerError)
				return
			}

			// find or create a filter
			var filter storage.Filter
			rsDB.Where("url_id = ?",id).FirstOrCreate(&filter)
			filter.URLID = uint(id)

			// setup the return URL paramaters
			action := strings.TrimSpace(r.FormValue("action"))
			var redirecturl string

			switch action {
				case "detail":
					redirecturl = "/details?id=" + r.FormValue("id")
				case "togglevisibility", "gallery":
					// page
					currPage, limit := getPageLimitFromForm(r)

					redirecturl = "/?perception_sort=" + r.FormValue("perception_sort") + "&page=" + currPage + "&limit=" + limit

					// get codes
					for _, code := range r.Form["code"] {
						redirecturl += "&code=" + code
					}

					// get already set filter tags
					for _, tag := range r.Form["ftag"] {
						redirecturl += "&tag=" + tag
					}

					// get show only
					if r.FormValue("onlynotes") == "true" {
						redirecturl += "&onlynotes=true"
					} else {
						redirecturl += "&onlynotes=false"
					}

					// get visibility flag
					if action == "togglevisibility" || r.FormValue("hide") == "true" {
						redirecturl += "&hide=true"
					} else {
						redirecturl += "&hide=false"
					}	
			}			

			// Finish the toggle visibility action here
			if action == "togglevisibility" {
				filter.Visible = false
				rsDB.Save(&filter)
				http.Redirect(w, r, redirecturl, http.StatusMovedPermanently)
				return
			}

			// The remainder of this action assumes an update to the filter record
			filter.GenericName = r.FormValue("genericname")
			filter.Notes = r.FormValue("notes")

			// Tags from form
			var tags []storage.Tag
			rsDB.Where("color IN ?",r.Form["tag"]).Find(&tags)

			// Not sure if this is the best way to handle 'changes'. Deleting all the old records,
			// then adding new ones. Seems like a waste of a query, but not sure how to replace/add values
			// without being a whole set of multiple queries
			var tagmaps []storage.Tagmap
			rsDB.Where("filter_id = ?", filter.ID).Delete(&tagmaps)

			// add the selected tag IDs to the filter
			for _, tag := range tags{
				filter.Tagmaps = append(filter.Tagmaps,storage.Tagmap{
					TagID:	tag.ID,
				})
			}

			// Save the record
			rsDB.Save(&filter)
			
			// Return to the expected page
			http.Redirect(w, r, redirecturl, http.StatusMovedPermanently)
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

// getPageLimit gets the limit and page query string values from a request
func getPageLimitFromForm(r *http.Request) (page string, limit string) {

	page = strings.TrimSpace(r.FormValue("page"))
	limit = strings.TrimSpace(r.FormValue("limit"))

	if page == "" {
		page = "-1"
	}
	if limit == "" {
		limit = "0"
	}
	return
}

func initializeFilterValues(rsDB *gorm.DB) error {
	// setup universe of filters from the database
	var codes []int
	rsDB.Table("urls").Distinct("response_code").Order("response_code").Find(&codes)
	RespCodeFilter = make(map[string]bool)
	for _, c := range codes{
		// checking against '0' in case a preflight/chromedp error.
		if c != 0 {
			cstr := strconv.Itoa(c)
			RespCodeFilter[cstr] = false
		}
	}

	// populate the tags table if not already in the DB
	TagFilterMap = make(map[interface{}]interface{})
	tagcolors := [11]string{"azure","indigo","purple","pink","red","orange","yellow","lime","green","teal","cyan"}
	for _, color := range tagcolors{
		var tag storage.Tag
		rsDB.Table("tags").FirstOrCreate(&tag,storage.Tag{Color: color,Name: ""})
		TagFilterMap[tag] = 0
	}

	return nil
}

// Structure to help render detail page objects
type DetailTemplateHelper struct {
	Url		*storage.URL
	Tags	*map[interface{}]interface{}
}