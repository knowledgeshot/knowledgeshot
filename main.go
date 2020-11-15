package main

import (
	"crypto/tls"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/ptgms/knowledgeshot/helpers"
	"html/template"
	"net/http"
	"strconv"
)

var version = "0.1"

func homePage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/home.html")

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	files, errF := helpers.FileCount("pages/")
	println(files)
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

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homePage)
	router.HandleFunc("/page/{term}", getPage)
	server := http.Server{
		Addr:    ":80",
		Handler: router,
		TLSConfig: &tls.Config{
			NextProtos: []string{"h2", "http/1.1"},
		},
	}

	fmt.Printf("Server listening on %s", server.Addr)
	if err := /*server.ListenAndServeTLS("certs/cert.pem", "certs/privkey.pem");*/ server.ListenAndServe(); err != nil {
		fmt.Println(err)
	}
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

		files, errF := helpers.FileCount("pages/")
		println(files)
		if errF != nil {
			println("Error while trying to retrieve indexed!")
			println(errF.Error())
			return
		}

		items := struct {
			TitleHead string
			Title     string
			Img       string
			Content   string
			Sources   string
			Version   string
		}{
			TitleHead: response.Title,
			Title:     response.Title,
			Img:       response.Image,
			Content:   response.Text,
			Sources:   links,
			Version:   version,
		}
		_ = t.Execute(w, items)

		//_, _ = fmt.Fprintf(w, response.Title+"\n\n"+response.Text+"\n\nImages:\n"+response.Image+"\n\nSources:\n"+links)
	}

}

func main() {
	println("Khinsider-Ripper API has started!\nPress CTRL+C to end.")
	testLinks := []string{"test1", "test2"}
	helpers.MakePage("Test", "Test", "Test", testLinks)

	handleRequests()
}
