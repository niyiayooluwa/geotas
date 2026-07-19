package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/niyiayooluwa/geotas/internal/middleware"
	"github.com/niyiayooluwa/geotas/internal/model"
	"github.com/niyiayooluwa/geotas/internal/service"
)

type AttendanceHandler struct {
	attendanceService *service.AttendanceService
}

func NewAttendanceHandler(attendanceService *service.AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{
		attendanceService: attendanceService,
	}
}

func (h *AttendanceHandler) MarkAttendanceQR(w http.ResponseWriter, r *http.Request) {
	var userID string = r.Context().Value(middleware.UserIDKey).(string)

	var req model.MarkAttendanceQRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.attendanceService.MarkAttendanceQR(r.Context(), userID, req)
	if err != nil {
		log.Printf("[AttendanceQR Error] UserID: %s, Error: %v\n", userID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *AttendanceHandler) MarkAttendanceOTP(w http.ResponseWriter, r *http.Request) {
	var userID string = r.Context().Value(middleware.UserIDKey).(string)

	var req model.MarkAttendanceOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.attendanceService.MarkAttendanceOTP(r.Context(), userID, req)
	if err != nil {
		log.Printf("[AttendanceOTP Error] UserID: %s, Error: %v\n", userID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *AttendanceHandler) GetAttendanceBySession(w http.ResponseWriter, r *http.Request) {
	var userID string = r.Context().Value(middleware.UserIDKey).(string)
	var sessionID string = chi.URLParam(r, "id")

	response, err := h.attendanceService.GetAttendanceBySession(r.Context(), userID, sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
