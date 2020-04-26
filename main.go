package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
)

// Allows user to specify IP address and port number to listen on
var addr = flag.String("addr", ":8080", "http service address")

// An HTML template that can be returned by a request
var homeTemplate = template.Must(template.New("").Parse("<h1>SUP BRAH</h1>"))

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)

	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func serveSup(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)

	_ = homeTemplate.Execute(w, "")
}

func main() {
	flag.Parse()

	hub := newHub()
	go hub.run()

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/sup", serveSup)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
