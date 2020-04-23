package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/markbates/pkger"
)

type router struct {
	devMode bool
}

func (a *router) handler() http.Handler {
	web := web{devMode: a.devMode}

	var dir http.FileSystem = pkger.Dir("/public")
	if a.devMode {
		d, _ := os.Getwd()
		log.Println("Pwd:", d)
		dir = http.Dir("./public")
	}
	public := http.FileServer(dir)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", web.home)
	r.Get("/date/{date}/", web.date)
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		public.ServeHTTP(w, r)
	})
	return r
}
