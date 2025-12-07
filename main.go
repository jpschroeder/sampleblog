package main

import (
	"context"
	"database/sql"
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/jpschroeder/sampleblog/posts"
	"github.com/jpschroeder/sampleblog/queries"
	"github.com/jpschroeder/sampleblog/util"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed **/*.html
var htmlTemplates embed.FS

func routes(queries *queries.Queries, templates util.TemplateParser) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/posts", http.StatusFound)
	}))
	postsController := posts.NewPostsController(queries, templates)
	mux.HandleFunc("GET /posts", postsController.Index)
	mux.HandleFunc("GET /posts/{id}", util.WithID(postsController.Show))
	mux.HandleFunc("GET /posts/new", postsController.New)
	mux.HandleFunc("GET /posts/{id}/edit", util.WithID(postsController.Edit))
	mux.HandleFunc("POST /posts", postsController.Create)
	mux.HandleFunc("POST /posts/{id}", util.WithID(postsController.Update))
	mux.HandleFunc("POST /posts/{id}/destroy", util.WithID(postsController.Destroy))

	var handler http.Handler = mux
	// Add middleware
	return handler
}

//go:embed schema.sql
var ddl string

func schema(db *sql.DB, ctx context.Context) error {
	_, err := db.ExecContext(ctx, ddl)
	return err
}

func main() {
	dbFile := flag.String("db", "database.db",
		"the path to the sqlite database file \n"+
			"it will be created if it does not already exist\n")
	httpaddr := flag.String("httpaddr", "localhost:8080",
		"the address/port to listen on for http \n"+
			"use :<port> to listen on all addresses\n")
	flag.Parse()

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	db, err := sql.Open("sqlite3", *dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = schema(db, ctx)
	if err != nil {
		log.Fatal(err)
	}

	queries := queries.New(db)
	templates := util.NewHTMLTemplateParser(htmlTemplates)

	handler := routes(queries, templates)
	httpServer := &http.Server{Addr: *httpaddr, Handler: handler}

	go func() {
		log.Printf("listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	}()
	wg.Wait()
}
