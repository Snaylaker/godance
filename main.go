package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type Dance struct {
	ID          int
	FileName    string
	Title       string
	Description string
}

var db *sql.DB

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.ParseFiles("./html/template/index.html")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db, err = sql.Open("sqlite3", "./dance.db")

	var dance Dance
	err = db.QueryRow("SELECT id, file_name, title, description FROM dance ").Scan(&dance.ID, &dance.FileName, &dance.Title, &dance.Description)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Dance not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	fmt.Print(dance.Description)
	err = tpl.Execute(w, dance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {

	fs := http.FileServer(http.Dir("./static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", indexHandler)
	http.ListenAndServe(":8082", nil)
}
