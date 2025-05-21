package main

import (
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"site-analyzer/internal/analyzer"
	"time"

	"golang.org/x/net/context"
)

var tmpl = template.Must(template.ParseFiles("internal/web/templates/index.html"))
var logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/analyze", analyzeHandler)
	logger.Info("Server started at http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		logger.Error("Error starting server:" + err.Error())
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.Execute(w, nil)
}

func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	url := r.FormValue("url")
	if url == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	result, err := analyzer.AnalyzePage(ctx, url)
	if err != nil {

		tmpl.Execute(w, map[string]interface{}{
			"Error": "Error analyzing page: " + err.Error(),
		})
		return
	}

	data := map[string]interface{}{
		"Error":        nil, // No error
		"Title":        result.Title,
		"HTMLVersion":  result.HTMLVersion,
		"LoginForm":    result.LoginForm,
		"Headings":     result.Headings,
		"Internal":     result.Internal,
		"External":     result.External,
		"Inaccessible": result.Inaccessible,
	}

	errorExecute := tmpl.Execute(w, data)
	if errorExecute != nil {
		http.Error(w, "Error rendering template: "+errorExecute.Error(), http.StatusInternalServerError)
	}
}
