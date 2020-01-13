package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"golang.org/x/time/rate"
)

func main() {

	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", index)
	router.HandleFunc("/", index)
	router.NotFoundHandler = http.HandlerFunc(notFoundPage)
	router.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowed)

	port := ":8080"
	print("API running on http://localhost" + port + "\n")

	log.Fatal(http.ListenAndServe(port, limit(router)))
}

var limiter = rate.NewLimiter(10, 1)

func limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if limiter.Allow() {
			next.ServeHTTP(w, r)
		} else {
			log.Print("429 response")
			t := time.Now().UTC()
			responseText := t.Format("2006-01-02T15:04:05") + "-429 Busy\n"
			http.Error(w, responseText, http.StatusTooManyRequests)
		}
	})
}

func index(w http.ResponseWriter, r *http.Request) {
	t := time.Now().UTC()
	log.Printf("Request starting: %s\n", t.Format("2006-01-02T15:04:05.999999"))

	responseText := t.Format("2006-01-02T15:04:05") + "-200 OK\n"
	w.Write([]byte(responseText))
}
func notFoundPage(w http.ResponseWriter, r *http.Request) {
	log.Printf("*** Not Found: %s\n", r.URL)
	http.Error(w, "404 page not found", http.StatusNotFound)
}

func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	log.Printf("*** Method Not Allowed: %s - %s\n", r.Method, r.URL)
	http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
}
