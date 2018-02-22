package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/goincremental/negroni-sessions"
	"github.com/goincremental/negroni-sessions/cookiestore"
	gmux "github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gorp.v2"
)

type User struct {
	Username string `db:"username"`
	Secret   []byte `db:"secret"`
}

type SearchResult struct {
	Name  string
	Price string
}

type Favourite struct {
	Id   int64  `db:Id`
	Name string `db:Name`
}

var db *sql.DB
var dbmap *gorp.DbMap

func main() {
	mux := gmux.NewRouter().StrictSlash(true)

	initDb()

	// router setting
	mux.HandleFunc("/", HomePageHandler).Methods("GET")
	mux.HandleFunc("/login/", LoginPageHandler).Methods("GET")
	mux.HandleFunc("/logout/", LogoutHandler).Methods("GET")
	mux.HandleFunc("/product/", ProductsHandler).Methods("GET")
	mux.HandleFunc("/about/", AboutHandler).Methods("GET")
	mux.HandleFunc("/contact/", ContactHandler).Methods("GET")
	mux.HandleFunc("/list/", ListHandler).Methods("PUT")

	mux.HandleFunc("/product/", SearchPageHandler).Methods("POST")
	//mux.HandleFunc("/order/", OrderHandler).Methods("POST")

	// static file
	cssPath := http.FileServer(http.Dir("./static/css"))
	imgPath := http.FileServer(http.Dir("./static/img"))
	rjsPath := http.FileServer(http.Dir("./static/rjs"))
	mux.PathPrefix("/css/").Handler(http.StripPrefix("/css/", cssPath))
	mux.PathPrefix("/img/").Handler(http.StripPrefix("/img/", imgPath))
	mux.PathPrefix("/rjs/").Handler(http.StripPrefix("/rjs/", rjsPath))

	n := negroni.Classic()
	n.Use(sessions.Sessions("wildview-session", cookiestore.New([]byte("my-secret-wildview"))))
	n.Use(negroni.HandlerFunc(verifyUser))
	n.UseHandler(mux)
	n.Run(":80")
}

func checkErr(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}

func getStringFromSession(r *http.Request, key string) string {
	var strVal string
	if val := sessions.GetSession(r).Get(key); val != nil {
		strVal = val.(string)
	}
	return strVal
}

func initDb() {
	var user, password, server, database string

	user = "root"
	password = "iloveyou"
	server = "tcp(127.0.0.1)"
	database = "wildviewdb"

	db, err := sql.Open("mysql", user+":"+password+"@"+server+"/"+database)
	checkErr(err, "sql.Open failed")

	dbmap = &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
	dbmap.AddTableWithName(Favourite{}, "favourites").SetKeys(true, "Id")
	dbmap.AddTableWithName(User{}, "users").SetKeys(false, "username")
	err = dbmap.CreateTablesIfNotExists()
	checkErr(err, "Create tables failed")
}

type ContentReturn struct {
	Error string
}

type Page struct {
	Favourites []Favourite
	User       string
	Content    ContentReturn
}

// Handlers begin here
func HomePageHandler(w http.ResponseWriter, r *http.Request) {
	p := Page{Favourites: []Favourite{}, User: getStringFromSession(r, "User"), Content: ContentReturn{}}
	if _, err := dbmap.Select(&p.Favourites, "select * from favourites"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl, err := template.New("").ParseFiles("templates/header.html",
		"templates/footer.html",
		"templates/home.html",
		"templates/base.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if err := tmpl.ExecuteTemplate(w, "base", p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type LoginPage struct {
	User    string
	Content ContentReturn
}

func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	p := LoginPage{}
	if r.FormValue("register") != "" {
		secret, _ := bcrypt.GenerateFromPassword([]byte(r.FormValue("password")), bcrypt.DefaultCost)
		user := User{r.FormValue("username"), secret}
		if err := dbmap.Insert(&user); err != nil {
			p.Content.Error = "Username is already taken, pick another one!"
		} else {
			sessions.GetSession(r).Set("User", user.Username)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
	} else if r.FormValue("login") != "" {
		user, err := dbmap.Get(User{}, r.FormValue("username"))
		if err != nil {
			p.Content.Error = "Either " + r.FormValue("username") + " or your password is invalid!"
		} else if user == nil {
			p.Content.Error = "Either " + r.FormValue("username") + " or your password is invalid!"
		} else {
			u := user.(*User)
			if err = bcrypt.CompareHashAndPassword(u.Secret, []byte(r.FormValue("password"))); err != nil {
				p.Content.Error = "Either " + r.FormValue("username") + " or your password is invalid!"
			} else {
				sessions.GetSession(r).Set("User", u.Username)
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
		}
	}
	tmpl, err := template.New("").ParseFiles("templates/header.html",
		"templates/footer.html",
		"templates/login.html",
		"templates/base.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if err := tmpl.ExecuteTemplate(w, "base", p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	sessions.GetSession(r).Set("User", nil)
	http.Redirect(w, r, "/", http.StatusFound)
}

func ProductsHandler(w http.ResponseWriter, r *http.Request) {
	p := Page{Favourites: []Favourite{}, User: getStringFromSession(r, "User"), Content: ContentReturn{}}
	if _, err := dbmap.Select(&p.Favourites, "select * from favourites"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl, err := template.New("").ParseFiles("templates/header.html",
		"templates/footer.html",
		"templates/home.html",
		"templates/base.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if err := tmpl.ExecuteTemplate(w, "base", p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	p := Page{Favourites: []Favourite{}, User: getStringFromSession(r, "User"), Content: ContentReturn{}}
	if _, err := dbmap.Select(&p.Favourites, "select * from favourites"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl, err := template.New("").ParseFiles("templates/header.html",
		"templates/footer.html",
		"templates/about.html",
		"templates/base.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if err := tmpl.ExecuteTemplate(w, "base", p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ContactHandler(w http.ResponseWriter, r *http.Request) {
	p := Page{Favourites: []Favourite{}, User: getStringFromSession(r, "User"), Content: ContentReturn{}}
	if _, err := dbmap.Select(&p.Favourites, "select * from favourites"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl, err := template.New("").ParseFiles("templates/header.html",
		"templates/footer.html",
		"templates/contact.html",
		"templates/base.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if err := tmpl.ExecuteTemplate(w, "base", p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ListHandler(w http.ResponseWriter, r *http.Request) {
	//initDb()

	inv1 := &Favourite{0, "He"}
	inv2 := &Favourite{0, "sdfasdfa"}

	err := dbmap.Insert(inv1, inv2)

	fmt.Printf("inv1.Id=%d  inv2.Id=%d\n", inv1.Id, inv2.Id)
	/*stmt, err := db.Prepare("INSERT INTO likes (Name) values(?);")
	checkErr(err, "Prepare insertion failed")

	_, err = stmt.Exec("hello")
	checkErr(err, "Execute insertion failed")
	*/

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func SearchPageHandler(w http.ResponseWriter, r *http.Request) {
	results := []SearchResult{
		SearchResult{"Hello", "3.5"},
		SearchResult{"gg", "4.5"},
	}
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(results); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

/*func OrderHandler(w http.ResponseWriter, r *http.Request) {

}*/

// Middleware Functions begin here
func verifyUser(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.URL.Path != "/login/" {
		next(w, r)
		return
	}
	if username := getStringFromSession(r, "User"); username != "" {
		if user, _ := dbmap.Get(User{}, username); user != nil {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		next(w, r)
		return
	}
	next(w, r)
	return
}
