package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

func main() {
	http.DefaultClient.Timeout = 180 * time.Second

	if len(os.Args) != 3 {
		log.Fatalf("target argument missing\nusage: %s [address] [target]\nexample: %s :8080 http://server/\n", os.Args[0], os.Args[0])
	}

	u := mustParse(os.Args[2])
	h := newHandler(u)

	mux := http.NewServeMux()
	mux.HandleFunc("/", h)

	s := http.Server{
		Addr:         os.Args[1],
		Handler:      h,
		ReadTimeout:  180 * time.Second,
		WriteTimeout: 180 * time.Second,
	}

	log.Printf("redirect to %s", os.Args[2])
	log.Printf("open port %s", os.Args[1])

	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("http server: %v", err)
	}
}

func newHandler(u *url.URL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		failures := 0
		done := false
		p := httputil.NewSingleHostReverseProxy(u)

		p.ModifyResponse = func(res *http.Response) error {
			if shouldRetry(res.StatusCode) {
				return fmt.Errorf("request failed: error code %d", res.StatusCode)
			}
			return nil
		}

		p.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			u := r.URL.String()
			log.Printf("proxy %s: %v", u, err)
			failures++
			done = false
		}

		body, err := ioutil.TempFile("", "req")
		if err != nil {
			log.Printf("create temp file: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer os.Remove(body.Name())

		_, err = io.Copy(body, r.Body)
		if err != nil {
			log.Printf("copy body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		r.Body.Close()

		err = body.Close()
		if err != nil {
			log.Printf("close copy of body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for !done {
			log.Printf("%s %s", r.Method, r.URL.String())

			delay := 1 * time.Second

			select {
			case <-ctx.Done():
				w.WriteHeader(http.StatusInternalServerError)
				return
			case <-time.After(delay):
			}

			done = true

			f, err := os.Open(body.Name())
			if err != nil {
				log.Printf("open copy of body: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer f.Close()

			r.Body = f

			p.ServeHTTP(w, r)
		}
	}
}

func shouldRetry(i int) bool {
	if i < 500 || i > 599 {
		return false
	}
	if i == http.StatusNotImplemented {
		return false
	}

	return true
}

func mustParse(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(fmt.Sprintf("parse url %s: %v", s, err))
	}

	return u
}
