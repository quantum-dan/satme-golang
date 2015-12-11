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
	// "html/template"
	"log"
	// "gopkg.in/mgo.v2"
	// "gopkg.in/mgo.v2/bson"
)
