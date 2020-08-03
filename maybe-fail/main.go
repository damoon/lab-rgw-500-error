package main

import (
	"log"
	"math/rand"
	"net/http"
)

func hello(w http.ResponseWriter, req *http.Request) {
	fail := (rand.Int31() % 2) == 0

	if fail {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("fail"))
		log.Println("fail")
		return
	}

	w.Write([]byte("ok"))
	log.Println("ok")
}

func main() {
	http.HandleFunc("/", hello)

	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("http server: %v", err)
	}
}
