package app

import (
	"masspay/internal/handler"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func NewRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"POST", "GET"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Use(removeTrailingSlashMiddleware)

	r.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/masspay", http.StatusFound)
	}))

	r.Route("/", func(r chi.Router) {
		r.Post("/masspay/api/upload", handler.FileUploadHandler)
		r.Get("/masspay", serveFile("./web/index.html"))
		serveAllStaticFiles(r, "/masspay/static", "./web/static")
	})

	return r
}

func serveFile(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}
}

func serveAllStaticFiles(r chi.Router, route string, dir string) {
	fs := http.StripPrefix(route, http.FileServer(http.Dir(dir)))
	r.Get(route+"/*", func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})
}

func removeTrailingSlashMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") {
			http.Redirect(w, r, strings.TrimSuffix(r.URL.Path, "/"), http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	})
}
