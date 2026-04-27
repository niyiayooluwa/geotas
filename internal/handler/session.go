package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/niyiayooluwa/geotas/internal/middleware"
	"github.com/niyiayooluwa/geotas/internal/model"
	"github.com/niyiayooluwa/geotas/internal/service"
)

type SessionHandler struct {
	sessionService *service.SessionService
}

func NewSessionHandler(sessionService *service.SessionService) *SessionHandler {
	return &SessionHandler{sessionService: sessionService}
}

func (h *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var userID string = r.Context().Value(middleware.UserIDKey).(string)

	var req model.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.sessionService.CreateSession(r.Context(), userID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *SessionHandler) CloseSession(w http.ResponseWriter, r *http.Request) {
	var userID string = r.Context().Value(middleware.UserIDKey).(string)
	var sessionID string = chi.URLParam(r, "id")

	response, err := h.sessionService.CloseSession(r.Context(), userID, sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *SessionHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	var userID string = r.Context().Value(middleware.UserIDKey).(string)
	var sessionID string = chi.URLParam(r, "id")

	if err := h.sessionService.DeleteSession(r.Context(), userID, sessionID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *SessionHandler) GetSessionsByCourse(w http.ResponseWriter, r *http.Request) {
	var userID string = r.Context().Value(middleware.UserIDKey).(string)
	var courseID string = chi.URLParam(r, "courseId")

	response, err := h.sessionService.GetSessionsByCourse(r.Context(), userID, courseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}