package session

import (
	"encoding/json"
	"errors"
	"net/http"
)

type SessionHandler struct {
	manager SessionManager
}

func NewSessionHandler(manager SessionManager) *SessionHandler {
	return &SessionHandler{manager: manager}
}

func (h *SessionHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	var request struct {
		WorkingDir string `json:"working_dir"`
	}

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if request.WorkingDir == "" {
		http.Error(w, "working_dir is required", http.StatusBadRequest)
		return
	}

	created, err := h.manager.Create(
		r.Context(),
		request.WorkingDir,
	)
	if err != nil {
		if errors.Is(err, ErrInvalidWorkingDir) {
			http.Error(w, "invalid working directory", http.StatusBadRequest)
			return
		}

		http.Error(w, "could not create session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created)
}

func (h *SessionHandler) Get(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := r.PathValue("id")

	found, err := h.manager.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}

		http.Error(w, "could not get session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(found)
}

func (h *SessionHandler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := r.PathValue("id")

	err := h.manager.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}

		http.Error(w, "could not delete session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
