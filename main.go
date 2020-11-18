package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

const (
	mySigningKey = "WOW,MuchShibe,ToDogge"
)

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	key   = []byte("super-secret-key")
	store = sessions.NewCookieStore(key)
)

func ExampleNew(mySigningKey []byte) (string, error) {
	// Create the token
	token := jwt.New(jwt.SigningMethodHS256)
	// Set some claims
	claims := make(jwt.MapClaims)
	claims["foo"] = "bar"
	claims["exp"] = time.Now().Add(time.Hour * 2).Unix()
	token.Claims = claims
	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString(mySigningKey)
	return tokenString, err
}

var addr = flag.String("addr", ":8080", "http service address")

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

func serveChat(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)

	session, _ := store.Get(r, "cookie-name")
	exp := session.Values["exp"].(int64)
	now := time.Now().Unix()

	if now >= exp {
		fmt.Println("IT HAS EXPIRED")
		http.Redirect(w, r, "/", http.StatusFound)
	}

	if r.URL.Path != "/chat" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "chat.html")
}

func serveLogin(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	token := r.FormValue("token")
	//fmt.Println(token)

	if token == "1234" {
		//createdToken, _ := ExampleNew([]byte(mySigningKey))
		//fmt.Println(createdToken)

		session, _ := store.Get(r, "cookie-name")
		session.Values["authenticated"] = "true"
		session.Values["token"] = token
		session.Values["exp"] = time.Now().Add(time.Minute * 2).Unix()

		session.Save(r, w)

		http.Redirect(w, r, "/chat", http.StatusFound)
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func main() {
	f, err := os.OpenFile("testlogfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println("Server Started")

	flag.Parse()
	hub := newHub()
	go hub.run()
	log.Println("Hub Running")

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", serveHome)
	router.HandleFunc("/login", serveLogin).Methods("POST")
	router.HandleFunc("/chat", serveChat)
	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	log.Fatal(http.ListenAndServe(":8080", router))
}
