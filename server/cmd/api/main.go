package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN not set")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal("cannot open db:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("cannot ping db:", err)
	}
	log.Println("DB OK")

	var version string
	if err := db.QueryRow("select version()").Scan(&version); err != nil {
		log.Fatal("db query failed:", err)
	}
	log.Println("DB version:", version)

	// ✅ all routes must be registered inside main()
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	http.HandleFunc("/api/boards", boardsHandler)

	log.Println("API listening on :8080 (with DB)")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// ===== Handlers and DTOs below main() =====
type BoardDTO struct {
	ID    string    `json:"id"`
	Name  string    `json:"name"`
	Lists []ListDTO `json:"lists"`
}

type ListDTO struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Position int       `json:"position"`
	Tasks    []TaskDTO `json:"tasks"`
}

type TaskDTO struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Position int    `json:"position"`
}

func boardsHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	payload := BoardDTO{
		ID:   id,
		Name: "Demo Board",
		Lists: []ListDTO{
			{ID: "l1", Name: "To Do", Position: 0, Tasks: []TaskDTO{
				{ID: "t1", Title: "Wire API → DB", Status: "in_progress", Position: 0},
				{ID: "t2", Title: "Add migrations", Status: "todo", Position: 1},
			}},
			{ID: "l2", Name: "In Progress", Position: 1, Tasks: []TaskDTO{}},
			{ID: "l3", Name: "Done", Position: 2, Tasks: []TaskDTO{}},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}
