package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

func routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/posts", http.StatusFound)
	}))
	mux.Handle("GET /posts", postsIndexHandler())
	mux.Handle("GET /posts/{id}", postsShowHandler())
	mux.Handle("GET /posts/new", postsNewHandler())
	mux.Handle("GET /posts/{id}/edit", postsEditHandler())
	mux.Handle("POST /posts", postsCreateHandler())
	mux.Handle("POST /posts/{id}", postsUpdateHandler())
	mux.Handle("POST /posts/{id}/destroy", postsDestroyHandler())

	var handler http.Handler = mux
	// Add middleware
	return handler
}

func schema(db *sql.DB, ctx context.Context) error {
	_, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS posts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	content TEXT NOT NULL
)`)
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

	db, err := initDB(*dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = schema(db, ctx)
	if err != nil {
		log.Fatal(err)
	}

	handler := routes()
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
