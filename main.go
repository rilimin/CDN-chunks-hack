package main

import (
	"fmt"
	"net/http"
	"text/template"

	"chunks/downloader"
	"chunks/uploader"
	"chunks/utils"
)

func main() {
	cfg := utils.ReadCfg()

	http.HandleFunc("/", home)

	http.HandleFunc("/download/{id1}/{id2}", downloader.DownloadHandler)
	http.HandleFunc("/video/{id1}/{id2}", downloader.RangeVideo)
	http.HandleFunc("/file", uploader.UploadHandler)

	fs := http.FileServer(http.Dir("./public/css"))
	http.Handle("/css/", http.StripPrefix("/css/", fs))

	fmt.Println("Server started on port " + cfg.Port)

	http.ListenAndServe(":"+cfg.Port, nil)
}

func home(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("public/home.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
