package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
)

const addr = "0.0.0.0"

func main() {
	http.HandleFunc("/", func(wr http.ResponseWriter, r *http.Request) {
		t, e := template.ParseFiles("kreativio/index.html")
		if e != nil {
			log.Fatal(e.Error())
		}
		t.Execute(wr, nil)
	})
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("kreativio"))))
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("kreativio/assets"))))
	log.Println("Starting on", addr, os.Getenv("FRONTEND_PORT"))
	log.Println(http.ListenAndServe(strings.Join([]string{addr, os.Getenv("FRONTEND_PORT")}, ":"), nil))
}
