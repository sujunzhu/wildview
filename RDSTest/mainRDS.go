package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"net/http"

	"encoding/json"
	"github.com/codegangsta/negroni"
)

type Page struct {
	Name string
}

type SearchResult struct {
	Name string
}

var db *sql.DB

func verifyDatabase(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if err := db.Ping(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func main() {
	templates := template.Must(template.ParseFiles("templates/index.html"))

	mux := http.NewServeMux()

	var user, password, server, database string

	user = "sujunzhu"
	password = "iloveyourassbaby"
	server = "tcp(wildviewdb.cozovlbefpqs.us-west-2.rds.amazonaws.com:3344)"
	database = "wildviewdb"

	_, err := sql.Open("mysql", user+":"+password+"@"+server+"/"+database)

	if err != nil {

	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		p := Page{Name: "Sujun"}

		if name := r.FormValue("name"); name != "" {
			p.Name = name
		}

		if err := templates.ExecuteTemplate(w, "index.html", p); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		results := []SearchResult{
			SearchResult{"Hello"},
			SearchResult{"gg"},
		}

		encoder := json.NewEncoder(w)
		if err := encoder.Encode(results); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	})

	n := negroni.Classic()
	//n.Use(negroni.HandlerFunc(verifyDatabase))
	n.UseHandler(mux)
	n.Run(":80")
}
