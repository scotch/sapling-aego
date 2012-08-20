package server

import (
	"code.google.com/p/gorilla/pat"
	"github.com/scotch/hal/auth"
	"github.com/scotch/hal/auth/appengine_openid"
	"github.com/scotch/hal/auth/password"
	"github.com/scotch/hal/context"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
)

var App = map[string]string{
	"Title":       "Sapling AEGo",
	"Description": "When you need more then a seed",
	"Author":      "Scotch Media",
}

var Env = map[string]interface{}{
	"Production": false,
}

const (
	API_URL = "/-/api/v1"
)

func init() {

	r := pat.New()

	// Auth

	// Default config; shown here for demonstration.
	auth.BaseURL = "/-/auth/"
	auth.LoginURL = "/login"
	auth.LogoutURL = "/-/auth/logout"
	auth.SuccessURL = "/"

	// Register the providers
	auth.Register("appengine_openid", appengine_openid.New())
	auth.Register("password", password.New())

	// API

	// Root
	r.Get("/", index)

	http.Handle("/", r)
}

var (
	indexTmpl = loadTmpl("template/index.html")
)

func templateData() map[string]interface{} {
	m := make(map[string]interface{})
	m["App"] = App
	m["Env"] = Env
	return m
}

func loadTmpl(t string) *template.Template {
	b, err := ioutil.ReadFile(t)
	check(err)
	s := string(b)
	tmpl, err := template.New("index").Delims("{{{", "}}}").Parse(s)
	check(err)
	return tmpl
}

func index(w http.ResponseWriter, r *http.Request) {
	// This must be added in order to Initialize the Application
	c := context.NewContext(r)
	// Handle hash fragments.
	if frag := r.URL.Query().Get("_escaped_fragment_"); frag != "" {
		c.Infof(`Search Engine Request: _escaped_fragment_: %v`, frag)
		fragHandler(w, r)
		return
	}
	// Handle not found.
	if r.URL.Path != "/" {
		p := r.URL.Path
		if strings.HasPrefix(p, "/-/") || strings.HasPrefix(p, "/_ah/") {
			http.Error(w, "404 Not found", http.StatusNotFound)
			return
		}
		// Redirect to hash so that angular routing can handle the request.
		hbp := "/#" + p
		http.Redirect(w, r, hbp, http.StatusFound)
		return
	}
	m := templateData()
	if err := indexTmpl.Execute(w, m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// TODO implement this.
func fragHandler(w http.ResponseWriter, r *http.Request) {
	_ = context.NewContext(r)
	_ = r.URL.Query().Get("_escaped_fragment_")
	http.Error(w, "404 Not found", http.StatusNotFound)
}

//check aborts the current execution if err is non-nil.
func check(err error) {
	if err != nil {
		panic(err)
	}
}
