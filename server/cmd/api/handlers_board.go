package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// ---- DTOs for board payload ----

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
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Status       string   `json:"status"`
	Position     int      `json:"position"`
	Assignees    []string `json:"assignees"`
	CommentCount int      `json:"comment_count"`
}

// ---- Handler ----

func boardsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// 🔐 now requires auth (so we can scope by workspace membership)
	sess, ok := getSessionFromRequest(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	qid := r.URL.Query().Get("id")
	inc := r.URL.Query().Get("include")

	var wantLists, wantTasks bool
	switch inc {
	case "", "lists,tasks":
		wantLists, wantTasks = true, true
	case "lists":
		wantLists, wantTasks = true, false
	default:
		wantLists, wantTasks = false, false
	}

	// 1) board (scoped to user's workspace membership)
	var boardID, boardName string
	var err error
	if qid != "" {
		err = db.QueryRow(`
            SELECT b.id, b.name
            FROM boards b
            JOIN workspace_members m ON m.workspace_id = b.workspace_id
            WHERE m.user_id = $1 AND b.id = $2
        `, sess.UserID, qid).Scan(&boardID, &boardName)
		if err == sql.ErrNoRows {
			http.Error(w, "board not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "board query failed", http.StatusInternalServerError)
			return
		}
	} else {
		err = db.QueryRow(`
            SELECT b.id, b.name
            FROM boards b
            JOIN workspace_members m ON m.workspace_id = b.workspace_id
            WHERE m.user_id = $1
            ORDER BY b.created_at ASC
            LIMIT 1
        `, sess.UserID).Scan(&boardID, &boardName)
		if err == sql.ErrNoRows {
			http.Error(w, "no board found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "board query failed", http.StatusInternalServerError)
			return
		}
	}

	// 2) lists
	lists := make([]ListDTO, 0)
	if wantLists {
		rows, err := db.Query(`SELECT id, name, position FROM lists WHERE board_id=$1 ORDER BY position ASC`, boardID)
		if err != nil {
			http.Error(w, "lists query failed", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var l ListDTO
			if err := rows.Scan(&l.ID, &l.Name, &l.Position); err == nil {
				l.Tasks = make([]TaskDTO, 0) // non-nil slice
				lists = append(lists, l)
			}
		}
	}

	// 3) tasks per list
	if wantTasks {
		for i := range lists {
			trows, err := db.Query(`SELECT id, title, description, status, position FROM tasks WHERE list_id=$1 ORDER BY position ASC`, lists[i].ID)
			if err != nil {
				http.Error(w, "tasks query failed", http.StatusInternalServerError)
				return
			}
			tasks := make([]TaskDTO, 0)
			for trows.Next() {
				var t TaskDTO
				if err := trows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Position); err == nil {
					// assignees
					arows, _ := db.Query(`SELECT user_id FROM task_assignees WHERE task_id=$1`, t.ID)
					aids := make([]string, 0)
					for arows.Next() {
						var uid string
						if err := arows.Scan(&uid); err == nil {
							aids = append(aids, uid)
						}
					}
					arows.Close()
					t.Assignees = aids

					// comment count
					_ = db.QueryRow(`SELECT COUNT(*) FROM comments WHERE task_id=$1`, t.ID).Scan(&t.CommentCount)

					tasks = append(tasks, t)
				}
			}
			trows.Close()
			lists[i].Tasks = tasks
		}
	}

	// 4) respond
	payload := BoardDTO{ID: boardID, Name: boardName, Lists: lists}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(payload)
}
