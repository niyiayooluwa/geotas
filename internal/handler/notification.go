package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/niyiayooluwa/geotas/internal/middleware"
	"github.com/niyiayooluwa/geotas/internal/model"
	"github.com/niyiayooluwa/geotas/internal/service"
)

type NotificationHandler struct {
	notifService *service.NotificationService
}

func NewNotificationHandler(notifService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifService: notifService}
}

func (h *NotificationHandler) GetUnseenNotifications(w http.ResponseWriter, r *http.Request) {
	var userID string = r.Context().Value(middleware.UserIDKey).(string)
	since := r.URL.Query().Get("since")

	response, err := h.notifService.GetUnseenNotificationsByUser(r.Context(), userID, since)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if response == nil {
		response = []model.NotificationResponse{}
	}
	json.NewEncoder(w).Encode(response)
}

func (h *NotificationHandler) MarkSeen(w http.ResponseWriter, r *http.Request) {
	var userID string = r.Context().Value(middleware.UserIDKey).(string)
	var notifID string = chi.URLParam(r, "id")

	err := h.notifService.MarkNotificationAsSeen(r.Context(), userID, notifID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
