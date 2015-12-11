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
	"github.com/gorilla/schema"
	"net/http"
	"html/template"
	"log"
	// "gopkg.in/mgo.v2"
	// "gopkg.in/mgo.v2/bson"
	"fmt"
)

type Person struct {
	Id int
	Name string
}

type NameForm struct {
	Get bool `schema:"-"`
	Name string `schema:"name"`
}

func main() {
	var PORT int = 8080
	r := mux.NewRouter()
	r.HandleFunc("/", index)
	r.HandleFunc("/static/{file}", serve_static)
	r.HandleFunc("/tmpl", tmpl_demo)
	r.HandleFunc("/form_demo", get_form_demo)
	r.HandleFunc("/form", form_demo)
	http.Handle("/", r)
	logstr := fmt.Sprintf("Listening on port %d", PORT)
	log.Println(logstr)
	portstr := fmt.Sprintf(":%d", PORT)
	err := http.ListenAndServe(portstr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func form_demo(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Fprintf(w, "Error 500: internal server error.  Failed to parse form.")
		log.Println("Error parsing form in form_demo")
	} else {
		decoder := schema.NewDecoder()
		results := new(NameForm)
		err = decoder.Decode(results, r.PostForm)
		if err != nil {
			fmt.Fprintf(w, "Error 500: internal server error.  Failed to evaluate form.")
			log.Println("Error reading form in form_demo")
		} else {
			t, _ := template.ParseFiles("templates/form.html")
			results.Get = false
			err = t.Execute(w, results)
			if err != nil {
				fmt.Fprintf(w, "Error 500: internal server error.  Failed to execute template.")
				log.Println("Failed to execute template in form_demo")
			}
		}
	}
}

func get_form_demo(w http.ResponseWriter, r *http.Request) {
	val := NameForm{Get: true, Name: ""}
	t, _ := template.ParseFiles("templates/form.html")
	err := t.Execute(w, val)
	if err != nil {
		fmt.Fprintf(w, "Error 500: internal server error.  Failed to execute template.")
		log.Println("Failed to execute template in get_form_demo")
	}
} 

func serve_static(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/" + mux.Vars(r)["file"])
}

func tmpl_demo(w http.ResponseWriter, r *http.Request) {
	people := []*Person{ &Person {Id: 0, Name: "Dan"},
		&Person {Id: 1, Name: "Josh"},
		}
	t, _ := template.ParseFiles("templates/demo.html")
	err := t.Execute(w, people)
	if err != nil {
		fmt.Fprintf(w, "Error 500: internal server error")
		log.Println("Error: Failed to execute template demo.html in function tmpl_demo")
	}
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

