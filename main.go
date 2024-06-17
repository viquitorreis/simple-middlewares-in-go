package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

type Middleware func(http.Handler) http.Handler

const port = ":4000"

func hwHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world!")
}

func main() {
	mux := http.NewServeMux()

	mainHandler := MiddlewareChain(http.HandlerFunc(hwHandler), RecoverMiddleware, LoggingMiddleware, StructureMiddleware)
	mux.Handle("/", mainHandler)

	println("Starting server on port", port)

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatal(err)
	}
}

func MiddlewareChain(h http.Handler, middlewares ...Middleware) http.Handler {
	for _, m := range middlewares {
		h = m(h)
	}
	return h
}

func StructureMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("StructureMiddleware: before")
		next.ServeHTTP(w, r) // call the next middleware in the chain
		slog.Info("StructureMiddleware: after")
	})
}

func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Info("Panic recovered %s\n%s", err, debug.Stack())
				http.Error(w, "Something went wrong during the request", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		log.Println("Req:", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)

		endTime := time.Since(startTime)
		slog.Info("request logging middleware", slog.String("method", r.Method), slog.String("path", r.URL.Path), slog.Duration("duration", endTime))
	})
}
