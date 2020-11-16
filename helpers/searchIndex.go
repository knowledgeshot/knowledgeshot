package helpers

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
)

type searchResult struct {
	Title  string    `json:"Title"`
	Path   string    `json:"text"`
	Author [4]string `json:"author"`
}

func IndexSearch() {
	var searchResults []searchResult

	files, err := ioutil.ReadDir("pages/")
	if err != nil {
		println("ERROR ON SEARCH INDEX:")
		println(err.Error())
		return
	}

	for fileList := range files {
		path := strings.ReplaceAll(files[fileList].Name(), ".json", "")
		response := GetPage(path)
		searchResults = append(searchResults, searchResult{
			Title:  response.Title,
			Path:   path,
			Author: response.Author,
		})
	}

	file, _ := json.MarshalIndent(searchResults, "", " ")

	_ = ioutil.WriteFile("resources-html/searchIndex-en.json", file, 0644)

}

func ReturnAll() []searchResult {
	var searchResults []searchResult
	jsonFile, err := os.Open("resources-html/searchIndex-en.json")

	if err != nil {
		searchResults = append(searchResults, searchResult{
			Title:  "500ERROR",
			Path:   "",
			Author: [4]string{},
		})
		return searchResults
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var data []searchResult

	_ = json.Unmarshal(byteValue, &data)

	return data
}

func ReturnSearch(term string) []searchResult {
	var searchResults []searchResult
	jsonFile, err := os.Open("resources-html/searchIndex-en.json")

	if err != nil {
		searchResults = append(searchResults, searchResult{
			Title:  "500ERROR",
			Path:   "",
			Author: [4]string{},
		})
		return searchResults
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var data []searchResult

	_ = json.Unmarshal(byteValue, &data)

	for i := range data {
		if strings.Contains(strings.ToLower(data[i].Title), strings.ToLower(term)) {
			searchResults = append(searchResults, searchResult{
				Title:  data[i].Title,
				Path:   data[i].Path,
				Author: data[i].Author,
			})
		}
	}

	return searchResults
}
