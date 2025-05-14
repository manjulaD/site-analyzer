package main

import (
	"html/template"
	"log"
	"net/http"
	"site-analyzer/internal/analyzer"
)

var tmpl = template.Must(template.ParseFiles("internal/web/templates/index.html"))

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/analyze", analyzeHandler)
	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.Execute(w, nil)
}

func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	url := r.FormValue("url")
	if url == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	result, err := analyzer.AnalyzePage(url)
	if err != nil {
		http.Error(w, "Error analyzing page: "+err.Error(), http.StatusInternalServerError)
		return
	}

	error := tmpl.Execute(w, result)
	if error != nil {
		return
	}
}
