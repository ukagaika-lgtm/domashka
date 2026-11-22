package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

var (
	sessions   = make(map[string]int64)
	sessionMux sync.RWMutex
)

func createSession(userID int64) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	sessionID := hex.EncodeToString(bytes)

	sessionMux.Lock()
	sessions[sessionID] = userID
	sessionMux.Unlock()

	return sessionID, nil
}

func getUserID(r *http.Request) (int64, bool) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return 0, false
	}

	sessionMux.RLock()
	userID, ok := sessions[cookie.Value]
	sessionMux.RUnlock()

	return userID, ok
}

func deleteSession(sessionID string) {
	sessionMux.Lock()
	delete(sessions, sessionID)
	sessionMux.Unlock()
}

func setSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})
}
