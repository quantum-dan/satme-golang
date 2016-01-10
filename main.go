package main

/* This is the Golang version of the SATme backend.
	It is currently designed for a synchronous front-end.  However, gorilla/websocket does facilitate websocket usage.
	It uses gorilla/sessions for sessions, gorilla/schema for forms, gorilla/mux for routing, and html/template for templates.
	It uses a MongoDB backend via mgo.v2.
	This is built in Golang for the combination of rapid development, ease & simplicity of use (vs Yesod), and high concurrent performance
with Goroutines (green threads).  This may or may not be the production version.
	The general functions and structs are located in the functions package.  Note that it must be moved to its own directory under $GOPATH/src before compiling.
*/

/* General Notes:
* MongoDB from here will NOT work properly with manually-input values.  All data must be inserted via mgo.
* The commented-out code is from testing the functionality and is no longer of use.  I've left it in in case someone needs a reference for how the concepts used here work.
 */

import (
	"github.com/gorilla/mux"      // Routing
	"github.com/gorilla/schema"   // Reads POST requests into structs
	"github.com/gorilla/sessions" // Used for cookies
	"html/template"               // HTML templates
	"log"                         // Logging to the terminal (for debugging)
	"net/http"                    // Basic HTTP library
	// "gopkg.in/mgo.v2"		// MongoDB driver
	// "gopkg.in/mgo.v2/bson"		// Used to convert to BSON for Mongo
	"fmt" // fmt is more or less equivalent to stdio in other languages
	// "golang.org/x/crypto/bcrypt"	// Secure password hashing, more secure for passwords than SHA3
	"time"
	// "errors"
	"functions"
	// "encoding/hex"
	"os"
)

/* START VARIABLE DECLARATIONS */

var decoder = schema.NewDecoder()                                                           // Decoder struct for form results
var store = sessions.NewCookieStore([]byte("non-production-a"), []byte("non-production-e")) // Session store with encryption and authentication keys
var dbstr = "localhost:27017"                                                               // MongoDB host

/* END VARIABLE DECLARATIONS */

/* START MAIN FUNCTION */

func main() {
	var PORT int = 8080 // So it's not hard-coded
	r := mux.NewRouter()
	r.HandleFunc("/", index)
	r.HandleFunc("/static/{file}", serve_static)
	r.HandleFunc("/login_post", post_login)
	r.HandleFunc("/login_get", get_login)
	r.HandleFunc("/create_acct_get", create_account_get)
	r.HandleFunc("/create_acct", create_account_post)
	r.HandleFunc("/quizzes", get_all_quizzes)
	r.HandleFunc("/quiz/{id}", display_quiz)
	r.HandleFunc("/grade/{id}", grade_quiz)
	r.HandleFunc("/score", view_score)
	r.HandleFunc("/admin", admin_panel)
	r.HandleFunc("/create_quiz", create_quiz)
	r.HandleFunc("/addq/{id}", addq_menu)
	r.HandleFunc("/add_question/{id}", add_question)
	http.Handle("/", r)
	logstr := fmt.Sprintf("Listening on port %d", PORT)
	log.Println(logstr)
	portstr := fmt.Sprintf(":%d", PORT)
	err := http.ListenAndServe(portstr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

/* END MAIN FUNCTION */

/* LOGGING FUNCTION */

func flog(content string) {
	// Write to log file
	file, err := os.OpenFile("./server.log", os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Println(err)
	} else {
		_, err = file.Write([]byte(content))
		if err != nil {
			log.Println(err)
		}
	}
	err = file.Close()
	if err != nil {
		log.Println(err)
	}
}

/* START ROUTING FUNCTIONS */

func serve_static(w http.ResponseWriter, r *http.Request) {
	// Static file server
	http.ServeFile(w, r, "static/"+mux.Vars(r)["file"])
}

func index(w http.ResponseWriter, r *http.Request) {
	// Index function
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		fmt.Fprintf(w, "Error 404: file not found.  Template could not be parsed.")
		log.Println("Error: template index.html not found")
		flog("index(): error parsing template")
	} else {
		err = t.Execute(w, "world")
		if err != nil {
			fmt.Fprintf(w, "Error 500: Internal server error.")
			log.Println("Error: template index.html (func index) failed to execute.")
			flog("index(): error executing template")
		}
	}
}

func create_quiz(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "login")
	if err != nil {
		http.Error(w, "failed to retrieve session", 500)
		flog("create_quiz: failed to retrieve session")
	} else {
		role, ok := session.Values["role"].(string)
		if !ok || (role != "su" && role != "admin") {
			http.Error(w, "failed to verify admin privileges.  are you logged in?", 500)
		} else {
			err = r.ParseForm()
			if err != nil {
				http.Error(w, "failed to parse form", 500)
				flog("create_quiz: failed to parse form")
			} else {
				quiz := new(functions.Quiz)
				err = decoder.Decode(quiz, r.PostForm)
				if err != nil {
					http.Error(w, "failed to read form", 500)
					log.Println(err)
					flog("create_quiz: failed to read form")
				} else {
					quiz.Questions = []functions.Question{}
					err = functions.InsertQuiz(functions.DbQuiz{quiz.Title, quiz.Questions})
					if err != nil {
						http.Error(w, "failed to insert quiz", 500)
						flog("create_quiz: failed to insert quiz")
					} else {
						fmt.Fprintf(w, "Successfully created quiz")
					}
				}
			}
		}
	}
}

func addq_menu(w http.ResponseWriter, r *http.Request) {
	// Menu to add questions to a specific quiz.  It's a workaround for some bugs--not ideal, but hopefully it works.
	session, err := store.Get(r, "login")
	if err != nil {
		http.Error(w, "failed to retrieve session", 500)
		flog("addq_menu: failed to retrieve session")
	} else {
		role, ok := session.Values["role"].(string)
		if !ok || (role != "su" && role != "admin") {
			http.Error(w, "failed to verify admin privileges.  are you logged in?", 500)
		} else {
			t, _ := template.ParseFiles("templates/addq.html")
			id, ok := mux.Vars(r)["id"]
			if !ok {
				http.Error(w, "failed to retrieve GET parameter", 500)
			} else {
				quiz, err := functions.RetrieveQuiz(id)
				if err != nil {
					http.Error(w, "failed to read quiz", 500)
					flog("addq_menu: failed to read quiz")
					log.Println(err)
				} else {
					err = t.Execute(w, quiz)
					if err != nil {
						http.Error(w, "failed to execute template", 500)
						flog("addq_menu: failed to execute template")
					}
				}
			}
		}
	}
}

func add_question(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "login")
	if err != nil {
		http.Error(w, "failed to retrieve session", 500)
		flog("add_question: failed to retrieve session")
	} else {
		role, ok := session.Values["role"].(string)
		if !ok || (role != "su" && role != "admin") {
			http.Error(w, "failed to verify admin privileges.  are you logged in?", 500)
		} else {
			err = r.ParseForm()
			if err != nil {
				http.Error(w, "failed to parse form", 500)
				flog("add_question: failed to parse form")
			} else {
				question := new(functions.Question)
				err = decoder.Decode(question, r.PostForm)
				if err != nil {
					http.Error(w, "failed to read form", 500)
					flog("add_question: failed to read form")
					log.Println(err)
				} else {
					id, ok := mux.Vars(r)["id"]
					if !ok {
						http.Error(w, "Invalid GET parameters", 500)
					} else {
						// tmp, _ := hex.DecodeString(id)
						// id = string(tmp)
						quiz, err := functions.RetrieveQuiz(id)
						if err != nil {
							http.Error(w, "failed to retrieve quiz", 500)
							flog("add_question: failed to retrieve quiz")
							log.Println(err)
						} else {
							quiz.Questions = append(quiz.Questions, *question)
							err = functions.UpdateQuiz(quiz)
							if err != nil {
								http.Error(w, "failed to update quiz", 500)
								flog("add_question: failed to update quiz")
								log.Println(err)
							} else {
								http.Redirect(w, r, "/addq/"+id, 302)
							}
						}
					}
				}
			}
		}
	}
}

func admin_panel(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "login")
	if err != nil {
		http.Error(w, "failed to retrieve session", 500)
		flog("admin_panel: failed to retrieve session")
	} else {
		role, ok := session.Values["role"].(string)
		if !ok || (role != "su" && role != "admin") {
			http.Error(w, "failed to verify admin privileges.  are you logged in?", 500)
		} else {
			quizzes, err := functions.RetrieveQuizzes("")
			if err != nil {
				http.Error(w, "failed to retrieve quizzes", 500)
				flog("admin_panel: failed to retrieve quizzes")
				log.Println(err)
			} else {
				t, _ := template.ParseFiles("templates/admin.html")
				newQuizzes := []functions.TmplQuiz{}
				for i := 0; i < len(quizzes); i++ {
					newQuizzes = append(newQuizzes, quizzes[i].GetTmplQuiz())
				}
				err := t.Execute(w, newQuizzes)
				if err != nil {
					http.Error(w, "failed to execute template", 500)
					flog("admin_panel: failed to execute template")
				}
			}
		}
	}
}

func grade_quiz(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "failed to parse form", 500)
		flog("grade_quiz: failed to parse form")
	} else {
		vars := mux.Vars(r)
		id, ok := vars["id"]
		if !ok {
			http.Error(w, "missing GET parameters", 404)
		} else {
			quiz := new(functions.Quiz)
			err = decoder.Decode(quiz, r.PostForm)
			if err != nil {
				http.Error(w, "failed to read form", 500)
				flog("grade_quiz: failed to read form")
				log.Println(err)
			} else {
				quiz.Id = id
				grade, err := quiz.Grade()
				if err != nil {
					http.Error(w, "failed to grade quiz", 500)
					flog("grade_quiz: failed to grade quiz")
				} else {
					session, err := store.Get(r, "login")
					if err == nil {
						login, ok := session.Values["username"]
						if ok {
							username, ok := login.(string)
							if ok {
								functions.UpdateScoreUsername(username, grade)
							}
						}
					}
					fmt.Fprintf(w, "Your grade is: %f%%", grade)
				}
			}
		}
	}
}

func view_score(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "login")
	if err != nil {
		http.Error(w, "failed to retrieve session", 500)
		flog("view_score: failed to retrieve session")
	} else {
		login, ok := session.Values["username"]
		if !ok {
			http.Error(w, "You are not logged in", 500)
		} else {
			username, ok := login.(string)
			if !ok {
				http.Error(w, "internal server error", 500)
			} else {
				user, err := functions.GetUser(username)
				if err != nil {
					http.Error(w, "failed to retrieve user data", 500)
					flog("view_score: failed to retrieve user data")
				} else {
					fmt.Fprintf(w, "Your highest score is %f%%", user.MaxScore)
				}
			}
		}
	}
}

func get_all_quizzes(w http.ResponseWriter, r *http.Request) {
	quizzes, err := functions.RetrieveQuizzes("")
	if err != nil {
		http.Error(w, "failed to retrieve quizzes", 500)
		flog("get_all_quizzes: failed to retrieve quizzes")
	} else {
		t, _ := template.ParseFiles("templates/all_quizzes.html")
		err = t.Execute(w, quizzes)
		if err != nil {
			http.Error(w, "failed to execute template", 500)
			flog("get_all_quizzes: failed to execute template")
		}
	}
}

func display_quiz(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	q_id, ok := vars["id"]
	if !ok {
		http.Error(w, "error: page not found--quiz page requires id parameter", 404)
	} else {
		quiz, err := functions.RetrieveQuiz(q_id)
		if err != nil {
			http.Error(w, "failed to retrieve quiz", 500)
			log.Println(err)
			flog("display_quiz: failed to retrieve quiz")
		} else {
			t, err := template.ParseFiles("templates/quiz.html")
			if err != nil {
				log.Println(err)
			} else {
				err = t.Execute(w, quiz.GetTmplQuiz())
				if err != nil {
					http.Error(w, "failed to execute template", 500)
					flog("display_quiz: failed to execute template")
					log.Println(err)
				}
			}
		}
	}
}

func create_account_get(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/acct_created.html")
	err = t.Execute(w, functions.SuccessLogin{false, "", "", false})
	if err != nil {
		http.Error(w, "failed to execute template", 500)
		flog("create_account_get: failed to execute template")
	}
}

func create_account_post(w http.ResponseWriter, r *http.Request) {
	// Creates an account from a post request.  Password is hashed with bcrypt.
	// If the request originator is logged in as an admin, role is set to the form value.  Otherwise, role is set to user.
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "failed to parse form", 500)
		flog("create_account_post: failed to parse form")
	} else {
		result := new(functions.User)
		err = decoder.Decode(result, r.PostForm)
		if err != nil {
			http.Error(w, "failed to read form", 500)
			flog("create_account_post: failed to read form")
		} else {
			session, err := store.Get(r, "login")
			if err != nil {
				http.Error(w, "internal server error", 500)
			} else {
				if role, ok := session.Values["role"]; !ok || role.(string) != "admin" {
					// Not logged in as admin: sets role to user
					result.Role = "user"
				}
				t, _ := template.ParseFiles("templates/acct_created.html")
				err = functions.CreateAccount(*result)
				if err != nil && err.Error() == "user already exists" {
					err = t.Execute(w, functions.SuccessLogin{false, result.Username, "", true})
					if err != nil {
						http.Error(w, "failed to execute template", 500)
						flog("create_account_post: failed to execute template 1")
					}
				} else if err != nil {
					http.Error(w, "internal server error", 500)
					flog("create_account_post: create account failed")
					log.Println(err)
				} else {
					err = t.Execute(w, functions.SuccessLogin{true, result.Username, result.Role, true})
					if err != nil {
						http.Error(w, "failed to execute template", 500)
						flog("create_account_post: failed to execute template 2")
					}
				}
			}
		}
	}
}

func post_login(w http.ResponseWriter, r *http.Request) {
	// Handles login requests.  Currently set up for plaintext passwords.
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "failed to parse form", 500)
		flog("post_login: failed to parse form")
	} else {
		result := new(functions.User)
		err = decoder.Decode(result, r.PostForm)
		if err != nil {
			http.Error(w, "failed to read form", 500)
			flog("post_login: failed to read form")
		} else {
			account, err := functions.CheckLogin(*result)
			if err != nil && err.Error() == "login failed" {
				time.Sleep(3 * time.Second)
				fmt.Fprintf(w, "invalid username or password")
			} else if err != nil {
				http.Error(w, "internal server error", 500)
				flog("post_login: login failure")
			} else {
				session, err := store.Get(r, "login")
				if err != nil {
					http.Error(w, "failed to retrieve session", 500)
					flog("post_login: failed to retrieve session")
				} else {
					session.Values["username"] = account.Username
					session.Values["role"] = account.Role
					session.Save(r, w)
					http.Redirect(w, r, "/login_get", 302)
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
		flog("get_login: failed to retrieve session")
	}
	username, ok := session.Values["username"]
	role, role_ok := session.Values["role"]
	if ok && role_ok {
		fmt.Fprintf(w, "You are logged in as "+username.(string)+", and your role is "+role.(string)) // Unsafe if username, role are not strings
	} else {
		fmt.Fprintf(w, "You are not logged in.")
	}
}

/* END ROUTING FUNCTIONS */
