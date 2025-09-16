package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
)

type createTaskReq struct {
	ListID      string `json:"list_id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type taskCreatedResp struct {
	ID       string `json:"id"`
	ListID   string `json:"list_id"`
	Title    string `json:"title"`
	Position int    `json:"position"`
	Status   string `json:"status"`
}

func createTaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := requireAuthAndCSRF(w, r) // reuse helper
	if !ok {
		return
	}

	var req createTaskReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.ListID == "" || req.Title == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}

	// Ensure the list belongs to a board in a workspace the user is a member of
	var valid bool
	if err := db.QueryRow(`
		SELECT EXISTS(
		  SELECT 1
		  FROM lists l
		  JOIN boards b ON b.id = l.board_id
		  JOIN workspace_members m ON m.workspace_id = b.workspace_id
		  WHERE l.id = $1 AND m.user_id = $2
		)
	`, req.ListID, sess.UserID).Scan(&valid); err != nil || !valid {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// Next position in the list
	var nextPos int
	_ = db.QueryRow(`SELECT COALESCE(MAX(position)+1, 0) FROM tasks WHERE list_id=$1`, req.ListID).Scan(&nextPos)

	// Insert
	var id, status string
	if err := db.QueryRow(`
		INSERT INTO tasks (list_id, title, description, position, status, created_by)
		VALUES ($1,$2,$3,$4,'todo',$5)
		RETURNING id, status
	`, req.ListID, req.Title, req.Description, nextPos, sess.UserID).Scan(&id, &status); err != nil {
		http.Error(w, "insert failed", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(taskCreatedResp{
		ID: id, ListID: req.ListID, Title: req.Title, Position: nextPos, Status: status,
	})
}
