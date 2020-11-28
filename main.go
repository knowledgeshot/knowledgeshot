package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/ptgms/knowledgeshot/helpers"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os/user"
	"strconv"
	"strings"
)

// Search result struct re-defined here so we can directly access it later
type searchResult struct {
	SearchResult string `json:"Title"`
	LinkContent  string `json:"text"`
	AuthorName   string `json:"authorname"`
	WrittenDate  string `json:"writtendate"`
}

type randomPageStruct struct {
	Link string `json:"link"`
	//Title string `json:"title"`
}

// Bump this on updates please :)
var version = "0.5"

// Currently kind of unused, maybe later i will add them somewhere
var niceStrings = []string{"I hope you're having an nice day!", "We don't track you - it's your right!", "We are grateful for your visit!"}

var username string

// Home page serve
func homePage(w http.ResponseWriter, _ *http.Request) {
	t, err := template.ParseFiles("templates/home.html")

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Let us get all jsons in pages to show how many sites we hold!
	files, errF := helpers.FileCount("pages/")
	if errF != nil {
		println("Error while trying to retrieve indexed!")
		println(errF.Error())

		items := struct {
			Indexed  string
			HostName string
			Version  string
		}{
			Indexed:  "ERROR",
			HostName: username,
			Version:  version,
		}
		_ = t.Execute(w, items)
		return
	}

	items := struct {
		Indexed  string
		HostName string
		Version  string
	}{
		Indexed:  strconv.Itoa(files),
		HostName: username,
		Version:  version,
	}
	_ = t.Execute(w, items)

	//_, _ = fmt.Fprintf(w, "Welcome to the Knowledgeshot Homepage!\n"+
	//		"")
	//fmt.Println("Endpoint Hit: " + r.RequestURI)
}

// Rate limit so servers wont get overwhelmed.
var limiter = helpers.NewIPRateLimiter(1, 5)

// Hell.
func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homePage)
	router.HandleFunc("/page/{term}", getPage)
	router.HandleFunc("/random", randomPage)
	router.HandleFunc("/res/bootstrap.css", bootstrap)
	router.HandleFunc("/search/{search}", searchFor)
	router.HandleFunc("/res/bootstrap4.css", bootstrap4)
	router.HandleFunc("/all", allArts)
	router.HandleFunc("/search", searchBlank)

	// API FUNCTIONS
	router.HandleFunc("/api/search/{term}", apiSearch)
	router.HandleFunc("/api/page/{term}", apiDisplay)
	router.HandleFunc("/api/random", apiRandom)

	server := http.Server{
		Addr:    ":8081",
		Handler: limitMiddleware(router),
		TLSConfig: &tls.Config{
			NextProtos: []string{"h2", "http/1.1"},
		},
	}

	fmt.Printf("Server listening on %s", server.Addr)
	if err := /*server.ListenAndServeTLS("certs/cert.pem", "certs/privkey.pem");*/ server.ListenAndServe(); err != nil {
		fmt.Println(err)
	}
}

func limitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limiter := limiter.GetLimiter(r.RemoteAddr)
		if !limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Serve blank page with no real function other than searching, which is handled by JS. We do insert our version though.
func searchBlank(w http.ResponseWriter, _ *http.Request) {
	t, err := template.ParseFiles("templates/searchblank.html")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	items := struct {
		Version string
	}{
		Version: version,
	}
	_ = t.Execute(w, items)

}

// Display all articles in one place, baby! (same structure as search just with all articles.)
func allArts(w http.ResponseWriter, _ *http.Request) {
	t, err := template.ParseFiles("templates/allresults.html")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	allResults := helpers.ReturnAll()

	var articlesParsed []searchResult

	for i := range allResults {
		articlesParsed = append(articlesParsed, searchResult{
			SearchResult: allResults[i].Title,
			LinkContent:  allResults[i].Path,
			AuthorName:   allResults[i].Author[0],
			WrittenDate:  allResults[i].Author[2],
		})
	}

	items := struct {
		Count         string
		SearchResults []searchResult
	}{
		Count:         strconv.Itoa(len(articlesParsed)),
		SearchResults: articlesParsed,
	}

	errEx := t.Execute(w, items)
	if errEx != nil {
		println(errEx.Error())
		w.WriteHeader(http.StatusInternalServerError)
	}

}

func apiSearch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if !helpers.Validate_key(r.Header.Get("API-Key")) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	key := vars["term"] // get our search key defined in the handler.

	searchresults := helpers.ReturnSearch(key) // let us fetch the results from our helper

	_ = json.NewEncoder(w).Encode(searchresults) // return our struct array as json
}

func apiRandom(w http.ResponseWriter, r *http.Request) {
	if !helpers.Validate_key(r.Header.Get("API-Key")) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	files, err := ioutil.ReadDir("pages/")
	if err != nil {
		println("Error while trying to retrieve indexed!")
		println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "An internal server error occured! It has been logged!")
		return
	}
	_ = json.NewEncoder(w).Encode(randomPageStruct{
		Link: strings.ReplaceAll(files[rand.Intn(len(files))].Name(), ".json", ""),
		//Title: files[rand.Intn(len(files))].Name(),
	})
}

func apiDisplay(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if !helpers.Validate_key(r.Header.Get("API-Key")) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	key := vars["term"]
	response := helpers.GetPage(key)

	if response.Title == "404ERROR" {
		w.WriteHeader(http.StatusNotFound)
		//_, _ = fmt.Fprintf(w, "404 - Article not found!")
	} else if response.Title == "500ERROR" {
		w.WriteHeader(http.StatusInternalServerError)
		//_, _ = fmt.Fprintf(w, "500 - Internal Server error!\n"+response.Text)
	} else {
		_ = json.NewEncoder(w).Encode(response) // return our struct array as json
	}
}

func searchFor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	key := vars["search"] // get our search key defined in the handler.

	searchresults := helpers.ReturnSearch(key) // let us fetch the results from our helper
	//println(strconv.Itoa(len(searchresults)))

	t, err := template.ParseFiles("templates/search.html")

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var searchParsed []searchResult

	for i := range searchresults { // convert helper searchresult to main searchresult (hacky pls so fix)
		searchParsed = append(searchParsed, searchResult{
			SearchResult: searchresults[i].Title,
			LinkContent:  searchresults[i].Path,
			AuthorName:   searchresults[i].Author[0],
			WrittenDate:  searchresults[i].Author[2],
		})
	}

	items := struct {
		SearchText    string
		Count         string
		SearchResults []searchResult
	}{
		SearchText:    key,
		Count:         strconv.Itoa(len(searchresults)),
		SearchResults: searchParsed,
	}

	errEx := t.Execute(w, items)
	if errEx != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

/* BOOTSTRAP RETURNERS */
func bootstrap(w http.ResponseWriter, _ *http.Request) {
	file, err := ioutil.ReadFile("resources-html/bootstrap.min.css")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "500 internal server error\n"+err.Error())
		return
	}
	w.Header().Add("Content-Type", "text/css")
	_, _ = w.Write(file)
}

func bootstrap4(w http.ResponseWriter, _ *http.Request) {
	file, err := ioutil.ReadFile("resources-html/bootstrap4.css")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "500 internal server error\n"+err.Error())
		return
	}
	w.Header().Add("Content-Type", "text/css")
	_, _ = w.Write(file)
}

/* --- DONE --- */

// We load an random page by selecting an random file from /pages. might make this use searchindex soon.
func randomPage(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir("pages/")
	if err != nil {
		println("Error while trying to retrieve indexed!")
		println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "An internal server error occured! It has been logged!")
		return
	}
	http.Redirect(w, r, "page/"+strings.ReplaceAll(files[rand.Intn(len(files))].Name(), ".json", ""), http.StatusSeeOther)

}

func getPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	key := vars["term"]
	response := helpers.GetPage(key)

	if response.Title == "404ERROR" {
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintf(w, "404 - Article not found!")
	} else if response.Title == "500ERROR" {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "500 - Internal Server error!\n"+response.Text)
	} else {
		var links string
		for link := range response.Links {
			links += response.Links[link] + "\n"
		}

		t, err := template.ParseFiles("templates/article.html")

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		items := struct {
			TitleHead   string
			Title       string
			Img         []string
			AuthorName  string
			AuthorImage string
			AuthorLink  string
			Written     string
			Content     template.HTML
			Sources     []string
			Version     string
			NiceString  string
		}{
			TitleHead:   response.Title,
			Title:       response.Title,
			Img:         response.Image,
			AuthorName:  response.Author[0],
			AuthorImage: response.Author[1],
			Written:     response.Author[2],
			AuthorLink:  response.Author[3],
			Content:     helpers.MarkDownReady(response.Text),
			Sources:     response.Links,
			Version:     version,
			NiceString:  niceStrings[rand.Intn(len(niceStrings))],
		}

		errEx := t.Execute(w, items)
		if errEx != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

}

func main() {
	userD, err := user.Current()
	if err != nil {
		username = "NULL"
	} else {
		username = userD.Name
	}

	//println(helpers.MarkDownReady("# test\n## test2"))
	println(fmt.Sprintf("Knowledgeshot %s has started with user %s!\nPress CTRL+C to end.", version, username))
	//helpers.MakePage("MarkdownTest", "# This is a test for markdown support inside KnowledgeShot\n## It is very experimental!\n- Lists do be\n- working doe!\nAnd [links](pog.com) too, of course.\n\nCode line: `code bro`", [4]string{"ptgms Industries", "nil", "nil", "nil"} ,[]string{}, []string{})
	helpers.IndexSearch()
	handleRequests()
}
