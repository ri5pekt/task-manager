package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/argon2"
)

// ---- password hashing ----

func hashPassword(pw string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	const (
		timeCost    = 3
		memoryCost  = 64 * 1024
		parallelism = 1
		keyLen      = 32
	)
	dk := argon2.IDKey([]byte(pw), salt, timeCost, memoryCost, uint8(parallelism), keyLen)
	return "$argon2id$v=19$m=65536,t=3,p=1$" +
		base64.RawStdEncoding.EncodeToString(salt) + "$" +
		base64.RawStdEncoding.EncodeToString(dk), nil
}

func verifyPassword(pw, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}
	salt, err1 := base64.RawStdEncoding.DecodeString(parts[4])
	want, err2 := base64.RawStdEncoding.DecodeString(parts[5])
	if err1 != nil || err2 != nil {
		return false
	}
	dk := argon2.IDKey([]byte(pw), salt, 3, 64*1024, 1, uint32(len(want)))
	if len(dk) != len(want) {
		return false
	}
	var v byte
	for i := range dk {
		v |= dk[i] ^ want[i]
	}
	return v == 0
}

// ---- session store (dev-only) ----

type Session struct {
	UserID  string
	CSRF    string
	Expires time.Time
}

var (
	sessions   = make(map[string]Session)
	sessionsMu sync.Mutex
)

func randToken(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b) // URL-safe
}

// ---- /api/register ----

type registerReq struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}
type registerResp struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func registerHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" || req.Name == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}
	h, err := hashPassword(req.Password)
	if err != nil {
		http.Error(w, "hashing failed", http.StatusInternalServerError)
		return
	}
	var id string
	err = db.QueryRow(
		`INSERT INTO users (email, password_hash, name) VALUES ($1,$2,$3) RETURNING id`,
		req.Email, h, req.Name,
	).Scan(&id)
	if err != nil {
		http.Error(w, "could not create user", http.StatusConflict)
		return
	}

	// 👇 best-effort provisioning (workspace + default board + lists)
	if err := provisionPersonalWorkspace(db, id, req.Name); err != nil {
		// Non-fatal for registration; just log it. User can still log in.
		log.Println("provisioning failed:", err)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(registerResp{ID: id, Email: req.Email, Name: req.Name})
}

func provisionPersonalWorkspace(db *sql.DB, userID, userName string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 1) workspace (slug NULL for now; keep it simple)
	var wsID string
	if err = tx.QueryRow(
		`INSERT INTO workspaces (name, slug) VALUES ($1, NULL) RETURNING id`,
		userName+"'s Workspace",
	).Scan(&wsID); err != nil {
		return err
	}

	// 2) membership (owner)
	if _, err = tx.Exec(
		`INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1,$2,'owner')
         ON CONFLICT DO NOTHING`,
		wsID, userID,
	); err != nil {
		return err
	}

	// 3) default board
	var boardID string
	if err = tx.QueryRow(
		`INSERT INTO boards (name, owner_id, workspace_id) VALUES ($1,$2,$3) RETURNING id`,
		"My Board", userID, wsID,
	).Scan(&boardID); err != nil {
		return err
	}

	// 4) three starter lists
	if _, err = tx.Exec(
		`INSERT INTO lists (board_id, name, position)
         VALUES ($1,'To Do',0), ($1,'In Progress',1), ($1,'Done',2)`,
		boardID,
	); err != nil {
		return err
	}

	// 4.1) three example tasks in "To Do"
	var todoListID string
	if err = tx.QueryRow(
		`SELECT id FROM lists WHERE board_id=$1 AND name='To Do' LIMIT 1`,
		boardID,
	).Scan(&todoListID); err != nil {
		return err
	}
	if _, err = tx.Exec(
		`INSERT INTO tasks (list_id, title, description, position, status, created_by)
         VALUES
         ($1, 'Welcome to your board', 'Drag cards between lists as work progresses.', 0, 'todo', $2),
         ($1, 'Create your first task', 'Click + to add tasks. Assign teammates later.', 1, 'todo', $2),
         ($1, 'Invite a teammate', 'Collaborate by inviting others to your workspace.', 2, 'todo', $2)`,
		todoListID, userID,
	); err != nil {
		return err
	}

	return tx.Commit()
}

// ---- /api/login ----

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type loginResp struct {
	UserID string `json:"user_id"`
}

func loginHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}
	var userID, pwHash string
	err := db.QueryRow(`SELECT id, password_hash FROM users WHERE email=$1`, req.Email).Scan(&userID, &pwHash)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "login failed", http.StatusInternalServerError)
		return
	}
	if !verifyPassword(req.Password, pwHash) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	sid := randToken(24)
	csrf := randToken(24)
	sessionsMu.Lock()
	sessions[sid] = Session{UserID: userID, CSRF: csrf, Expires: time.Now().Add(24 * time.Hour)}
	sessionsMu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "sid",
		Value:    sid,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf",
		Value:    csrf,
		Path:     "/",
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	})
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(loginResp{UserID: userID})
}

// ---- /api/me ----

type meResp struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func getSessionFromRequest(r *http.Request) (Session, bool) {
	c, err := r.Cookie("sid")
	if err != nil || c.Value == "" {
		return Session{}, false
	}
	sessionsMu.Lock()
	s, ok := sessions[c.Value]
	sessionsMu.Unlock()
	if !ok || time.Now().After(s.Expires) {
		return Session{}, false
	}
	return s, true
}

func meHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := getSessionFromRequest(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var email, name string
	if err := db.QueryRow(`SELECT email, name FROM users WHERE id=$1`, sess.UserID).Scan(&email, &name); err != nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(meResp{ID: sess.UserID, Email: email, Name: name})
}

// ---- CSRF helper for POST/PUT/PATCH/DELETE ----

func requireAuthAndCSRF(w http.ResponseWriter, r *http.Request) (Session, bool) {
	s, ok := getSessionFromRequest(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return Session{}, false
	}
	token := r.Header.Get("X-CSRF-Token")
	if token == "" || token != s.CSRF {
		http.Error(w, "forbidden", http.StatusForbidden)
		return Session{}, false
	}
	return s, true
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Best-effort: remove session by cookie value
	if sid, ok := readSIDCookie(r); ok {
		deleteSessionByID(sid)
	}

	// Expire cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "sid",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf",
		Value:    "",
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

// readSIDCookie extracts the session id from the "sid" cookie.
func readSIDCookie(r *http.Request) (string, bool) {
	c, err := r.Cookie("sid")
	if err != nil || c.Value == "" {
		return "", false
	}
	return c.Value, true
}

// deleteSessionByID removes a session from the in-memory store.
// Relies on the existing `sessions` map and `sessionsMu` mutex defined above.
func deleteSessionByID(sid string) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	delete(sessions, sid)
}
