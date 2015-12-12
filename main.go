package main

/* This is the Golang version of the SATme backend.
	It is currently designed for a synchronous front-end.  However, gorilla/websocket does facilitate websocket usage.
	It uses gorilla/sessions for sessions, gorilla/schema for forms, gorilla/mux for routing, and html/template for templates.
	It uses a MongoDB backend via mgo.v2.
	This is built in Golang for the combination of rapid development, ease & simplicity of use (vs Yesod), and high concurrent performance
with Goroutines (green threads).  This may or may not be the production version.
*/

/* General Notes:
* MongoDB from here will NOT work properly with manually-input values.  All data must be inserted via mgo.
*/

import (
	"github.com/gorilla/sessions"	// Used for cookies
	"github.com/gorilla/mux"	// Routing
	"github.com/gorilla/schema"	// Reads POST requests into structs
	"net/http"			// Basic HTTP library
	"html/template"			// HTML templates
	"log"				// Logging to the terminal (for debugging)
	"gopkg.in/mgo.v2"		// MongoDB driver
	"gopkg.in/mgo.v2/bson"		// Used to convert to BSON for Mongo
	"fmt"				// fmt is more or less equivalent to stdio in other languages
	"golang.org/x/crypto/bcrypt"	// Secure password hashing, more secure for passwords than SHA3
	"time"
)


var decoder = schema.NewDecoder()								// Decoder struct for form results
var store = sessions.NewCookieStore([]byte("non-production-a"), []byte("non-production-e"))	// Session store with encryption and authentication keys
var dbstr = "localhost:27017"									// MongoDB host

type Person struct { // For demo functions
	Id int
	Name string
}

type NameForm struct { // For demo functions
	Get bool `schema:"-"`
	Name string `schema:"name"`
}

type User struct { // For logging in.
	Username string `schema:"username"`
	Password string `schema:"password"`
	Role string `schema:"role"`
	Id int `schema:"id"`  // Not generally used beyond the database but will be useful for deleting users.
}

type DbUser struct { // Uses []byte password
	Username string
	Password []byte
	Role string
	Id int
}

type SuccessLogin struct {
	Success bool
	Username string
	Role string
	Execute bool
}

func main() {
	var PORT int = 8080 // So it's not hard-coded
	r := mux.NewRouter()
	r.HandleFunc("/", index)
	r.HandleFunc("/static/{file}", serve_static)
	r.HandleFunc("/tmpl", tmpl_demo)
	r.HandleFunc("/form_demo", get_form_demo)
	r.HandleFunc("/form", form_demo)
	r.HandleFunc("/session", session_demo_get)
	r.HandleFunc("/session_post", session_demo)
	r.HandleFunc("/login_post", post_login)
	r.HandleFunc("/login_get", get_login)
	r.HandleFunc("/create_acct_get", create_account_get)
	r.HandleFunc("/create_acct", create_account_post)
	http.Handle("/", r)
	logstr := fmt.Sprintf("Listening on port %d", PORT)
	log.Println(logstr)
	portstr := fmt.Sprintf(":%d", PORT)
	err := http.ListenAndServe(portstr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func create_account_get(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/acct_created.html")
	err = t.Execute(w, SuccessLogin{ false, "", "", false})
	if err != nil {
		http.Error(w, "failed to execute template", 500)
	}
}

func create_account_post(w http.ResponseWriter, r *http.Request) {
	// Creates an account from a post request.  Password is hashed with bcrypt.
	// If the request originator is logged in as an admin, role is set to the form value.  Otherwise, role is set to user.
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "failed to parse form", 500)
	} else {
		result := new(User)
		err = decoder.Decode(result, r.PostForm)
		if err != nil {
			http.Error(w, "failed to read form", 500)
		} else {
			session, err := store.Get(r, "login")
			if err != nil {
				http.Error(w, "internal server error", 500)
			} else {
				if role, ok := session.Values["role"]; !ok || role.(string) != "admin" {
					// Not logged in as admin: sets role to user
					result.Role = "user"
				}
				db, err := mgo.Dial(dbstr)
				defer db.Close()
				if err != nil {
					http.Error(w, "failed to access database", 500)
				} else {
					c := db.DB("server").C("users")
					dbresult := new(User)
					err = c.Find(bson.M{"username": result.Username}).One(&dbresult)
					if err != nil && err.Error() != "not found" {
						http.Error(w, "failed to read database", 500)
					} else if err == nil {
						// Result is found: account with that username already exists
						returning := SuccessLogin { false, result.Username, "", true }
						t, _ := template.ParseFiles("templates/acct_created.html")
						err = t.Execute(w, returning)
						if err != nil {
							http.Error(w, "failed to execute template", 500)
						}
					} else {
						password_bytestr, _ := bcrypt.GenerateFromPassword([]byte(result.Password), 10) // Note here: do NOT use a larger cost, it will (noticeably on the other end) take a very long time
						result.Password = string(password_bytestr)
						err = c.Insert(&result)
						if err != nil {
							http.Error(w, "failed to create account", 500)
						} else {
							t, _ := template.ParseFiles("templates/acct_created.html")
							err = t.Execute(w, SuccessLogin { true, result.Username, result.Role, true })
							if err != nil {
								http.Error(w, "failed to execute template", 500)
							}
						}
					}
				}
			}
		}
	}
}

func post_login(w http.ResponseWriter, r *http.Request) {
	// Handles login requests.  Currently set up for plaintext passwords.
	// Does not yet use sessions.  That's the next thing to be added.
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "failed to parse form", 500)
	} else {
		result := new(User)
		err = decoder.Decode(result, r.PostForm)
		if err != nil {
			http.Error(w, "failed to read form", 500)
		} else {
			dbsession, err := mgo.Dial(dbstr)
			defer dbsession.Close()
			if err != nil {
				http.Error(w, "failed to connect to database", 500)
			} else {
				c := dbsession.DB("server").C("users")
				dbresult := new(DbUser)
				err = c.Find(bson.M{ "username": result.Username }).One(&dbresult)
				if err != nil  && err.Error() != "not found" {
					http.Error(w, "failed to read database", 500)
				} else {
					if dbresult.Username == result.Username && bcrypt.CompareHashAndPassword(dbresult.Password, []byte(result.Password)) == nil {
						session, err := store.Get(r, "login")
						if err != nil {
							http.Error(w, "Failed to retrieve session", 500)
						}
						session.Values["username"] = dbresult.Username
						session.Values["role"] = dbresult.Role
						session.Save(r, w)
						http.Redirect(w, r, "/login_get", 302)
					} else {
						time.Sleep(time.Second) // Prevents brute-force attacks.  Between this and bcrypt, we should be able to allow weak passwords with relative safety.
						fmt.Fprintf(w, "Failed to log in")
					}
				}
			}
		}
	}
}

func get_login(w http.ResponseWriter, r *http.Request) {
	// Retrieves login session
	session, err := store.Get(r, "login")
	if err != nil {
		http.Error(w, "Failed to retrieve session", 500)
	}
	username, ok := session.Values["username"]
	role, role_ok := session.Values["role"]
	if ok && role_ok {
		fmt.Fprintf(w, "You are logged in as " + username.(string) + ", and your role is " + role.(string)) // Unsafe if username, role are not strings
	} else {
		fmt.Fprintf(w, "You are not logged in.")
	}
}

func dbtest() {
	// Database test function (not in use)
	dbsession, err := mgo.Dial("localhost:27017")
	if err != nil {
		log.Fatal("Database failed to connect")
	}
	defer dbsession.Close()
	c := dbsession.DB("test").C("people")
	err = c.Insert(&Person {0, "Daniel Philippus"},
		&Person {1, "Nour Haridy"})
	if err != nil {
		log.Fatal(err)
	}
	result := new(Person)
	err = c.Find(bson.M{"name": "Daniel Philippus"}).One(&result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Name: ", result.Name)
	err = c.Find(bson.M{"id": 1}).One(&result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Name: ", result.Name)
}

func session_demo(w http.ResponseWriter, r *http.Request) {
	// Session demo for a fake login
	type Login struct {
		Username string `schema:"username"`
		Password string `schema:"password"`
	}
	err := r.ParseForm()
	if err != nil {
		fmt.Fprintf(w, "Error 500: internal server error.  Failed to parse form.")
		log.Println("Error parsing form in session_demo")
	} else {
		results := new(Login)
		err = decoder.Decode(results, r.PostForm)
		if err != nil {
			fmt.Fprintf(w, "Error 500: internal server error.  Failed to evaluate form.")
			log.Println("Error evaluating form in session_demo")
		} else {
			if results.Username == "dan" && results.Password == "password" {
				session, err := store.Get(r, "login")
				if err != nil {
					http.Error(w, err.Error(), 500)
					log.Println("Error retrieving session in session_demo")
				} else {
					session.Values["user"] = "dan"
					session.Save(r, w)
					t, _ := template.ParseFiles("templates/login_demo.html")
					t.Execute(w, "You are successfully logged in as dan")
				}
			} else {
				t, _ := template.ParseFiles("templates/login_demo.html")
				t.Execute(w, "You failed to log in")
			}
		}
	}
}

func session_demo_get(w http.ResponseWriter, r *http.Request) {
	// Displays session results
	session, err := store.Get(r, "login")
	if err != nil {
		http.Error(w, err.Error(), 500)
	} else {
		if name, ok := session.Values["user"]; ok {
			t, _ := template.ParseFiles("templates/login_demo.html")
			t.Execute(w, "You are logged in as " + name.(string))
		} else {
			t, _ := template.ParseFiles("templates/login_demo.html")
			t.Execute(w, "You are not logged in.")
		}
	}
}

func form_demo(w http.ResponseWriter, r *http.Request) {
	// Demos form handling with schema
	err := r.ParseForm()
	if err != nil {
		fmt.Fprintf(w, "Error 500: internal server error.  Failed to parse form.")
		log.Println("Error parsing form in form_demo")
	} else {
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
	// Form for form test
	val := NameForm{Get: true, Name: ""}
	t, _ := template.ParseFiles("templates/form.html")
	err := t.Execute(w, val)
	if err != nil {
		fmt.Fprintf(w, "Error 500: internal server error.  Failed to execute template.")
		log.Println("Failed to execute template in get_form_demo")
	}
}

func serve_static(w http.ResponseWriter, r *http.Request) {
	// Static file server
	http.ServeFile(w, r, "static/" + mux.Vars(r)["file"])
}

func tmpl_demo(w http.ResponseWriter, r *http.Request) {
	// Demos HTML templates
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
	// Index function
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

