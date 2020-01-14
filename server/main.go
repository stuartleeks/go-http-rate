package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"golang.org/x/time/rate"
)

func main() {

	router := mux.NewRouter().StrictSlash(true)

	// router.HandleFunc("/api", index)
	router.HandleFunc("/api", limit2(index))
	router.HandleFunc("/dummy", dummy)
	router.NotFoundHandler = http.HandlerFunc(notFoundPage)
	router.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowed)

	port := ":8080"
	print("API running on http://localhost" + port + "\n")

	// log.Fatal(http.ListenAndServe(port, limit(router)))
	log.Fatal(http.ListenAndServe(port, router))
}

var limiter = rate.NewLimiter(200, 10)

func limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.RequestURI, "/dummy") || limiter.Allow() {
			next.ServeHTTP(w, r)
		} else {
			log.Print("429 response")
			t := time.Now().UTC()
			responseText := t.Format("2006-01-02T15:04:05") + "-429 Busy\n"
			http.Error(w, responseText, http.StatusTooManyRequests)
		}
	})
}

func limit2(handler http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if limiter.Allow() {
			handler.ServeHTTP(w, r)
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
	time.Sleep(100 * time.Millisecond)
}
func dummy(w http.ResponseWriter, r *http.Request) {
	t := time.Now().UTC()
	responseText := t.Format("Dummy-2006-01-02T15:04:05") + "-200 OK\n"
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
