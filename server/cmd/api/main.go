package main

import (
	"database/sql"
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

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	log.Println("API listening on :8080 (with DB)")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
