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
	var id string
	if err := db.QueryRow(`
   		INSERT INTO tasks (list_id, title, description, position, created_by)
   		VALUES ($1,$2,$3,$4,$5)
   		RETURNING id
 		`, req.ListID, req.Title, req.Description, nextPos, sess.UserID).Scan(&id); err != nil {
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
		RETURNING t.id, t.title, t.description, t.position
	`

	log.Printf("[updateTaskHandler] query OK:\n%s\nargs: %#v", query, args)

	var out taskCreatedResp
	// taskCreatedResp now includes Description (you already added it)
	if err := db.QueryRow(query, args...).Scan(&out.ID, &out.Title, &out.Description, &out.Position); err != nil {
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

// POST /api/tasks/reorder
// Body: { "task_id": "...", "to_list_id": "...", "to_index": 0 }
type reorderOrMoveReq struct {
	TaskID   string `json:"task_id"`
	ToListID string `json:"to_list_id"`
	ToIndex  int    `json:"to_index"`
}

type reorderOrMoveResp struct {
	ID       string `json:"id"`
	ListID   string `json:"list_id"`
	Position int    `json:"position"`
}

func reorderOrMoveTaskHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := requireAuthAndCSRF(w, r)
	if !ok {
		return
	}

	var req reorderOrMoveReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TaskID == "" || req.ToListID == "" {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "tx begin failed", http.StatusInternalServerError)
		return
	}
	defer func() { _ = tx.Rollback() }()

	// Current location
	var srcListID string
	var oldPos int
	if err := tx.QueryRow(`SELECT list_id, position FROM tasks WHERE id=$1`, req.TaskID).Scan(&srcListID, &oldPos); err == sql.ErrNoRows {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "lookup failed", http.StatusInternalServerError)
		return
	}

	// ACL: user must belong to src & dest workspaces
	var okSrc, okDst bool
	if err := tx.QueryRow(`
		SELECT EXISTS(
		  SELECT 1 FROM lists l
		  JOIN boards b ON b.id = l.board_id
		  JOIN workspace_members m ON m.workspace_id = b.workspace_id
		  WHERE l.id=$1 AND m.user_id=$2
		)`, srcListID, sess.UserID).Scan(&okSrc); err != nil || !okSrc {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if err := tx.QueryRow(`
		SELECT EXISTS(
		  SELECT 1 FROM lists l
		  JOIN boards b ON b.id = l.board_id
		  JOIN workspace_members m ON m.workspace_id = b.workspace_id
		  WHERE l.id=$1 AND m.user_id=$2
		)`, req.ToListID, sess.UserID).Scan(&okDst); err != nil || !okDst {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// Clamp index to valid bounds in destination
	var destCount int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM tasks WHERE list_id=$1`, req.ToListID).Scan(&destCount); err != nil {
		http.Error(w, "count dest failed", http.StatusInternalServerError)
		return
	}
	toIndex := req.ToIndex
	if toIndex < 0 {
		toIndex = 0
	}
	if srcListID == req.ToListID {
		if toIndex >= destCount {
			toIndex = max(0, destCount-1)
		}
	} else {
		// when moving across lists, inserting at the end is allowed (== destCount)
		if toIndex > destCount {
			toIndex = destCount
		}
	}

	// Same-list reorder: shift neighbors then set the task
	if srcListID == req.ToListID {
		if toIndex != oldPos {
			if toIndex < oldPos {
				// task moves up → push down the block [toIndex..oldPos-1]
				if _, err := tx.Exec(`
					UPDATE tasks SET position = position + 1
					WHERE list_id = $1 AND position >= $2 AND position < $3
				`, srcListID, toIndex, oldPos); err != nil {
					http.Error(w, "shift up block failed", http.StatusBadRequest)
					return
				}
			} else {
				// task moves down → pull up the block [oldPos+1..toIndex]
				if _, err := tx.Exec(`
					UPDATE tasks SET position = position - 1
					WHERE list_id = $1 AND position > $2 AND position <= $3
				`, srcListID, oldPos, toIndex); err != nil {
					http.Error(w, "shift down block failed", http.StatusBadRequest)
					return
				}
			}
			if _, err := tx.Exec(`UPDATE tasks SET position=$1, updated_at=NOW() WHERE id=$2`, toIndex, req.TaskID); err != nil {
				http.Error(w, "update pos failed", http.StatusBadRequest)
				return
			}
		}
	} else {
		// Cross-list: compact source, make room in dest, then move
		if _, err := tx.Exec(`
			UPDATE tasks SET position = position - 1
			WHERE list_id = $1 AND position > $2
		`, srcListID, oldPos); err != nil {
			http.Error(w, "compact source failed", http.StatusBadRequest)
			return
		}
		if _, err := tx.Exec(`
			UPDATE tasks SET position = position + 1
			WHERE list_id = $1 AND position >= $2
		`, req.ToListID, toIndex); err != nil {
			http.Error(w, "make room dest failed", http.StatusBadRequest)
			return
		}
		if _, err := tx.Exec(`
			UPDATE tasks SET list_id=$1, position=$2, updated_at=NOW() WHERE id=$3
		`, req.ToListID, toIndex, req.TaskID); err != nil {
			http.Error(w, "move failed", http.StatusBadRequest)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "commit failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(reorderOrMoveResp{
		ID: req.TaskID, ListID: req.ToListID, Position: toIndex,
	})
}

// tiny helper
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
