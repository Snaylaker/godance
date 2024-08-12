package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	http.HandleFunc("POST /dances/", insertDanceHandler)
	http.HandleFunc("PUT /dances/{id}", updateDanceHandler)
	http.HandleFunc("GET /dances/{id}", getDancesHandler)
	http.HandleFunc("GET /dances/{id}/edit", danceFormHandler)
	http.HandleFunc("GET /modal", modalHandler)

	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func modalHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "modal.html", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, file_name, title, description FROM dance")

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Dance not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	var dances []Dance
	for rows.Next() {
		var dance Dance

		if err := rows.Scan(&dance.ID, &dance.FileName, &dance.Title, &dance.Description); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		dances = append(dances, dance)
	}

	err = templates.ExecuteTemplate(w, "index.html", dances)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func rowsToDances(rows *sql.Rows) []Dance {
	var dances []Dance
	for rows.Next() {
		var dance Dance

		if err := rows.Scan(&dance.ID, &dance.FileName, &dance.Title, &dance.Description); err != nil {
			fmt.Print("Error while parisng dances from databases")
		}

		dances = append(dances, dance)
	}

	return dances
}

func danceFormHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var dance Dance
	err := db.QueryRow("SELECT id, file_name, title, description FROM dance WHERE id = ?", id).Scan(&dance.ID, &dance.FileName, &dance.Title, &dance.Description)

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

func insertDanceHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // max memory 10MB
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "failed ", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Create a new file in the uploads directory
	uploadDir := "./static/"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, 0755)
	}

	filePath := filepath.Join(uploadDir, handler.Filename)
	destFile, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Error saving the file", http.StatusInternalServerError)
		return
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(file)
	if err != nil {
		http.Error(w, "Error saving the file", http.StatusInternalServerError)
		return
	}

	title := r.FormValue("titre")
	description := r.FormValue("description")

	_, err = db.Exec("INSERT INTO dance (file_name, title, description) VALUES (?, ?, ?)", handler.Filename, title, description)
	if err != nil {
		http.Error(w, "Failed to insert dance", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func getDancesHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var dance Dance
	db.QueryRow("SELECT id, file_name, title, description FROM dance WHERE id = ?", id).Scan(&dance.ID, &dance.FileName, &dance.Title, &dance.Description)

	templates.ExecuteTemplate(w, "danceCard.html", dance)
}

func updateDanceHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	title := r.FormValue("title")
	description := r.FormValue("description")

	db.Exec("UPDATE dance SET title = ?, description = ? WHERE id = ?", title, description, id)

	var dance Dance

	db.QueryRow("SELECT id, file_name, title, description FROM dance WHERE id = ?", id).Scan(&dance.ID, &dance.FileName, &dance.Title, &dance.Description)

	templates.ExecuteTemplate(w, "danceCard.html", dance)
}
