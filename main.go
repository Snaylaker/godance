package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Dance struct {
	ID          int
	FileName    string
	Title       string
	Description string
}

var (
	db        *sql.DB
	templates *template.Template
)

func init() {
	var err error
	db, err = sql.Open("sqlite3", "./dance.db")
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	templates = template.Must(template.ParseGlob("./html/template/*.html"))
	fmt.Println("Parsed Templates:")
	for _, t := range templates.Templates() {
		fmt.Println("-", t.Name())
	}
}

func main() {
	fs := http.FileServer(http.Dir("./static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("PUT /dances/1", editDanceHandler)
	http.HandleFunc("GET /dances/1", dancesHandler)
	http.HandleFunc("GET /dances/1/edit", danceHandler)

	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	var dance Dance
	err := db.QueryRow("SELECT id, file_name, title, description FROM dance").Scan(&dance.ID, &dance.FileName, &dance.Title, &dance.Description)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Dance not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	err = templates.ExecuteTemplate(w, "index.html", dance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func danceHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("dancehandler")
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/dances/")
	idStr := strings.TrimSuffix(path, "/edit")
	id, err := strconv.Atoi(idStr)

	var dance Dance
	err = db.QueryRow("SELECT id, file_name, title, description FROM dance WHERE id = ?", id).Scan(&dance.ID, &dance.FileName, &dance.Title, &dance.Description)

	if err != nil {
		http.Error(w, "Invalid dance ID", http.StatusBadRequest)
		return
	}

	err = templates.ExecuteTemplate(w, "editDance.html", dance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func dancesHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/dances/")
	id, _ := strconv.Atoi(path)

	var dance Dance
	db.QueryRow("SELECT id, file_name, title, description FROM dance WHERE id = ?", id).Scan(&dance.ID, &dance.FileName, &dance.Title, &dance.Description)
	templates.ExecuteTemplate(w, "danceCard.html", dance)
}

func editDanceHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/dances/")
	id, _ := strconv.Atoi(path)

	title := r.FormValue("title")
	description := r.FormValue("description")

	db.Exec("UPDATE dance SET title = ?, description = ? WHERE id = ?", title, description, id)
	var dance Dance
	db.QueryRow("SELECT id, file_name, title, description FROM dance WHERE id = ?", id).Scan(&dance.ID, &dance.FileName, &dance.Title, &dance.Description)

	templates.ExecuteTemplate(w, "danceCard.html", dance)
}
