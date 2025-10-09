package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lavavrik/go-sm/api/kv"
	"github.com/lavavrik/go-sm/sysauth"
)

func SignIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if !sysauth.Validate(creds.Username, creds.Password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	sid, err := kv.CreateSession(creds.Username)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create session: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cookie", fmt.Sprintf("session_id=%s; HttpOnly; Path=/; Max-Age=%d", sid, int(kv.SessionTTL.Seconds())))
	json.NewEncoder(w).Encode(map[string]string{"session_id": sid})
}
