package web

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sensepost/gowitness/pkg/database"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/web/api"
)

// Server is a web server
type Server struct {
	Port           int
	DbUri          string
	ScreenshotPath string
}

// NewServer returns a new server intance
func NewServer(port int, dburi string, screenshotpath string) *Server {
	return &Server{
		Port:           port,
		DbUri:          dburi,
		ScreenshotPath: screenshotpath,
	}
}

// isJSON sets the Content-Type header to application/json
func isJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// Run a server
func (s *Server) Run() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.CleanPath)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	// get a db handle
	conn, err := database.Connection(s.DbUri, false)
	if err != nil {
		log.Fatal("could not connect to the database", "err", err)
		return
	}
	apih := api.NewApiHandler(conn)

	r.Route("/api", func(r chi.Router) {
		r.Use(isJSON)
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins: []string{"*"}, // TODO: flag this
		}))

		r.Get("/ping", apih.PingHandler)
		r.Get("/statistics", apih.StatisticsHandler)
		r.Get("/gallery", apih.GalleryHandler)
		r.Get("/list", apih.ListHandler)
		r.Get("/detail/{id}", apih.GalleryDetailHandler)
		r.Get("/wappalyzer", apih.WappalyzerHandler)
	})

	r.Mount("/screenshots",
		http.StripPrefix(
			"/screenshots/", http.FileServer(http.Dir(s.ScreenshotPath)),
		),
	)
	r.Handle("/*", SpaHandler())

	log.Info("starting web server", "port", s.Port)
	http.ListenAndServe(":"+strconv.Itoa(s.Port), r)
}
