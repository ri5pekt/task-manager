// handlers_tasks.go
package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type createTaskReq struct {
	ListID      string `json:"list_id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type taskCreatedResp struct {
	ID          string `json:"id"`
	ListID      string `json:"list_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Position    int    `json:"position"`
	Status      string `json:"status"`
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

type updateTaskReq struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
}

func updateTaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	sess, ok := requireAuthAndCSRF(w, r)
	if !ok {
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	var req updateTaskReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	// Build dynamic SET clause with correct $1, $2, ...
	sets := []string{}
	args := []any{}

	if req.Title != nil {
		sets = append(sets, "title=$"+strconv.Itoa(len(args)+1))
		args = append(args, strings.TrimSpace(*req.Title))
	}
	if req.Description != nil {
		sets = append(sets, "description=$"+strconv.Itoa(len(args)+1))
		args = append(args, *req.Description)
	}
	if len(sets) == 0 {
		http.Error(w, "nothing to update", http.StatusBadRequest)
		return
	}

	// WHERE placeholders come after SET args
	//  ... id is next
	args = append(args, id)
	idPos := len(args) // position of id we just appended

	//  ... user id is after that
	args = append(args, sess.UserID)
	userPos := len(args)

	query := `
		UPDATE tasks t
		SET ` + strings.Join(sets, ", ") + `, updated_at=NOW()
		WHERE t.id=$` + strconv.Itoa(idPos) + `
		AND EXISTS (
			SELECT 1
			FROM lists l
			JOIN boards b ON b.id = l.board_id
			JOIN workspace_members m ON m.workspace_id = b.workspace_id
			WHERE l.id = t.list_id AND m.user_id = $` + strconv.Itoa(userPos) + `
		)
		RETURNING t.id, t.title, t.description, t.status, t.position
	`

	log.Printf("[updateTaskHandler] query OK:\n%s\nargs: %#v", query, args)

	var out taskCreatedResp
	// taskCreatedResp now includes Description (you already added it)
	if err := db.QueryRow(query, args...).Scan(&out.ID, &out.Title, &out.Description, &out.Status, &out.Position); err != nil {
		http.Error(w, "update failed", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	sess, ok := requireAuthAndCSRF(w, r)
	if !ok {
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	res, err := db.Exec(`
        DELETE FROM tasks t
        WHERE t.id=$1
        AND EXISTS (
            SELECT 1
            FROM lists l
            JOIN boards b ON b.id=l.board_id
            JOIN workspace_members m ON m.workspace_id=b.workspace_id
            WHERE l.id=t.list_id AND m.user_id=$2
        )
    `, id, sess.UserID)
	if err != nil {
		http.Error(w, "delete failed", http.StatusBadRequest)
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		http.Error(w, "not found or forbidden", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
