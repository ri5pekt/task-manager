package main

import (
	"database/sql"
	"net/http"
)

func registerRoutes(db *sql.DB) {
	http.HandleFunc("/api/boards", func(w http.ResponseWriter, r *http.Request) {
		boardsHandler(w, r, db)
	})
	http.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) {
		registerHandler(w, r, db)
	})
	http.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		loginHandler(w, r, db)
	})
	http.HandleFunc("/api/me", func(w http.ResponseWriter, r *http.Request) {
		meHandler(w, r, db)
	})
	http.HandleFunc("/api/comments", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listCommentsHandler(w, r, db)
		case http.MethodPost:
			createCommentHandler(w, r, db)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			createTaskHandler(w, r, db)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/api/logout", logoutHandler)
	http.HandleFunc("/api/uploads", uploadHandler)
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir))))
}
