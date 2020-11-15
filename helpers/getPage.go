package helpers

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
)

type pageData struct {
	Title string   `json:"Title"`
	Text  string   `json:"text"`
	Image string   `json:"image"`
	Links []string `json:"links"`
}

func GetPage(term string) pageData {
	path := "pages/" + term + ".json"
	println(path)
	if FileExists(path) {
		jsonFile, err := os.Open(path)

		if err != nil {
			return pageData{
				Title: "500ERROR",
				Text:  err.Error(),
				Image: "",
				Links: nil,
			}
		}

		defer jsonFile.Close()

		byteValue, _ := ioutil.ReadAll(jsonFile)

		var data pageData

		_ = json.Unmarshal(byteValue, &data)

		return data

	} else {
		println("Tried accessing a page that doesn't exist")
		return pageData{
			Title: "404ERROR",
			Text:  "",
			Image: "",
			Links: nil,
		}
	}
}

func MakePage(title string, text string, image string, links []string) {
	page := pageData{
		Title: title,
		Text:  text,
		Image: image,
		Links: links,
	}
	file, _ := json.MarshalIndent(page, "", " ")

	_ = ioutil.WriteFile("pages/"+url.QueryEscape(title)+".json", file, 0644)
}
