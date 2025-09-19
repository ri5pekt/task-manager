// handlers_lists.go
package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// POST /api/lists/reorder
// Body: { "board_id": "...", "list_ids": ["id1","id2",...] }
type reorderListsReq struct {
	BoardID string   `json:"board_id"`
	ListIDs []string `json:"list_ids"`
}

func reorderListsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := requireAuthAndCSRF(w, r)
	if !ok {
		return
	}

	var req reorderListsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.BoardID == "" || len(req.ListIDs) == 0 {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	// ACL: user must be a member of the workspace that owns this board
	var allowed bool
	if err := db.QueryRow(`
		SELECT EXISTS (
		  SELECT 1
		  FROM boards b
		  JOIN workspace_members m ON m.workspace_id = b.workspace_id
		  WHERE b.id = $1 AND m.user_id = $2
		)
	`, req.BoardID, sess.UserID).Scan(&allowed); err != nil || !allowed {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "tx begin failed", http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	// Ensure all provided lists belong to this board (cheap guard)
	for _, id := range req.ListIDs {
		var ok bool
		if err := tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM lists WHERE id=$1 AND board_id=$2)`, id, req.BoardID).Scan(&ok); err != nil || !ok {
			http.Error(w, "list not in board", http.StatusBadRequest)
			return
		}
	}

	// Update positions 0..n-1 in the given order
	for i, id := range req.ListIDs {
		if _, err := tx.Exec(`UPDATE lists SET position=$1 WHERE id=$2 AND board_id=$3`, i, id, req.BoardID); err != nil {
			http.Error(w, "update failed", http.StatusBadRequest)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "commit failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"ok":true}`))
}
