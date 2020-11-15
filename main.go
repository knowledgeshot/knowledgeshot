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

var version = "0.1"
var niceStrings = []string{"I hope you're having an nice day!", "We don't track you - it's your right!", "We are grateful for your visit!"}

func homePage(w http.ResponseWriter, r *http.Request) {
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

func randomPage(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir("pages/")
	if err != nil {
		println("Error while trying to retrieve indexed!")
		println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "An internal server error occured! It has been logged!")
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
		_, _ = fmt.Fprintf(w, "500 - Server error!\n"+response.Text)
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

		imageToDraw := response.Image

		if imageToDraw == "nil" || imageToDraw == "" {
			imageToDraw = "https://i.ibb.co/cgRJ97N/unknown.png"
		}

		items := struct {
			TitleHead  string
			Title      string
			Img        string
			Content    string
			Sources    []string
			Version    string
			NiceString string
		}{
			TitleHead:  response.Title,
			Title:      response.Title,
			Img:        imageToDraw,
			Content:    response.Text,
			Sources:    response.Links,
			Version:    version,
			NiceString: niceStrings[rand.Intn(len(niceStrings))],
		}
		_ = t.Execute(w, items)

		//_, _ = fmt.Fprintf(w, response.Title+"\n\n"+response.Text+"\n\nImages:\n"+response.Image+"\n\nSources:\n"+links)
	}

}

func main() {
	println(fmt.Sprintf("Knowledgeshot %s has started!\nPress CTRL+C to end.", version))
	//testLinks := []string{"test1", "test2"}
	//helpers.MakePage("Test", "Test", "Test", testLinks)

	handleRequests()
}
