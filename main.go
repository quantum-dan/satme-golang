package main

/* This is the Golang version of the SATme backend.
	It is currently designed for a synchronous front-end.  However, gorilla/websocket does facilitate websocket usage.
	It uses gorilla/sessions for sessions, gorilla/schema for forms, gorilla/mux for routing, and html/template for templates.
	It uses a MongoDB backend via mgo.v2.
	This is built in Golang for the combination of rapid development, ease & simplicity of use (vs Yesod), and high concurrent performance
with Goroutines (green threads).  This may or may not be the production version.
*/

import (
	// "github.com/gorilla/sessions"
	"github.com/gorilla/mux"
	// "github.com/gorilla/schema"
	"net/http"
	"html/template"
	"log"
	// "gopkg.in/mgo.v2"
	// "gopkg.in/mgo.v2/bson"
	"fmt"
)

var PORT int = 8080

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", index)
	r.HandleFunc("/static/{file}", serve_static)
	http.Handle("/", r)
	logstr := fmt.Sprintf("Listening on port %d", PORT)
	log.Println(logstr)
	portstr := fmt.Sprintf(":%d", PORT)
	err := http.ListenAndServe(portstr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func serve_static(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/" + mux.Vars(r)["file"])
}

func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		fmt.Fprintf(w, "Error 404: file not found.  Template could not be parsed.")
		log.Println("Error: template index.html not found")
	} else {
		err = t.Execute(w, "world")
		if err != nil {
			fmt.Fprintf(w, "Error 500: Internal server error.")
			log.Println("Error: template index.html (func index) failed to execute.")
		}
	}
}

