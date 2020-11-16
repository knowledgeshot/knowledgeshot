package main

import (
	"crypto/tls"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/ptgms/knowledgeshot/helpers"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
)

type searchResult struct {
	SearchResult string `json:"Title"`
	LinkContent  string `json:"text"`
	AuthorName   string `json:"authorname"`
	WrittenDate  string `json:"writtendate"`
}

var version = "0.1"
var niceStrings = []string{"I hope you're having an nice day!", "We don't track you - it's your right!", "We are grateful for your visit!"}

var searchTemplate = "<div class=\"media-body pb-3 mb-0 small lh-125 border-bottom border-gray\">\n                <div class=\"d-flex justify-content-between align-items-center w-100\">\n                    <strong class=\"text-gray-dark\">{{ .SearchResult }}</strong>\n                    <a href=\"{{ .LinkContent }}\">View</a>\n                </div>\n                <span class=\"d-block\">Written by {{ .AuthorName }} on {{ .WrittenDate }}.</span>\n            </div>"

func homePage(w http.ResponseWriter, _ *http.Request) {
	t, err := template.ParseFiles("templates/home.html")

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	files, errF := helpers.FileCount("pages/")
	if errF != nil {
		println("Error while trying to retrieve indexed!")
		println(errF.Error())
		return
	}

	items := struct {
		Indexed string
	}{
		Indexed: strconv.Itoa(files),
	}
	_ = t.Execute(w, items)

	//_, _ = fmt.Fprintf(w, "Welcome to the Knowledgeshot Homepage!\n"+
	//		"")
	//fmt.Println("Endpoint Hit: " + r.RequestURI)
}

var limiter = helpers.NewIPRateLimiter(1, 5)

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
	server := http.Server{
		Addr:    ":80",
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

func searchBlank(w http.ResponseWriter, r *http.Request) {
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

func allArts(w http.ResponseWriter, r *http.Request) {
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

func searchFor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	key := vars["search"]

	searchresults := helpers.ReturnSearch(key)
	//println(strconv.Itoa(len(searchresults)))

	t, err := template.ParseFiles("templates/search.html")

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var searchParsed []searchResult

	for i := range searchresults {
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
			Content     string
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
			Content:     response.Text,
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
	println(fmt.Sprintf("Knowledgeshot %s has started!\nPress CTRL+C to end.", version))
	//helpers.MakePage("Test", "Test", "Test", []string{"test1", "test2"})
	helpers.IndexSearch()
	handleRequests()
}
