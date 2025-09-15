package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// ---- DTOs ----
type createCommentReq struct {
	TaskID string `json:"task_id"`
	Body   string `json:"body"`
}
type commentResp struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	AuthorID  string    `json:"author_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}
type commentItem struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	AuthorID  string    `json:"author_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

// ---- POST /api/comments (auth + CSRF) ----
func createCommentHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := requireAuthAndCSRF(w, r)
	if !ok {
		return
	}

	var req createCommentReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TaskID == "" || strings.TrimSpace(req.Body) == "" {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	var id string
	var created time.Time
	if err := db.QueryRow(
		`INSERT INTO comments (task_id, author_id, body) VALUES ($1,$2,$3) RETURNING id, created_at`,
		req.TaskID, sess.UserID, req.Body,
	).Scan(&id, &created); err != nil {
		http.Error(w, "insert failed", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(commentResp{
		ID: id, TaskID: req.TaskID, AuthorID: sess.UserID, Body: req.Body, CreatedAt: created,
	})
}

// ---- GET /api/comments?task_id=... (public) ----
func listCommentsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	taskID := r.URL.Query().Get("task_id")
	if taskID == "" {
		http.Error(w, "missing task_id", http.StatusBadRequest)
		return
	}
	rows, err := db.Query(`
		SELECT id, task_id, author_id, body, created_at
		FROM comments
		WHERE task_id = $1
		ORDER BY created_at ASC
	`, taskID)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := make([]commentItem, 0)
	for rows.Next() {
		var c commentItem
		if err := rows.Scan(&c.ID, &c.TaskID, &c.AuthorID, &c.Body, &c.CreatedAt); err == nil {
			items = append(items, c)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(items)
}
