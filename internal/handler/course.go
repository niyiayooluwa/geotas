package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/niyiayooluwa/geotas/internal/middleware"
	"github.com/niyiayooluwa/geotas/internal/model"
	"github.com/niyiayooluwa/geotas/internal/service"
)

type CourseHandler struct {
	courseService *service.CourseService
}

func NewCourseHandler(courseService *service.CourseService) *CourseHandler {
	return &CourseHandler{courseService: courseService}
}

func (h *CourseHandler) CreateCourse(w http.ResponseWriter, r *http.Request) {
	// get user_id from context — set by auth middleware
	var userID string = r.Context().Value(middleware.UserIDKey).(string)

	// decode request body
	var req model.CreateCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// call service
	response, err := h.courseService.CreateCourse(r.Context(), userID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *CourseHandler) JoinCourse(w http.ResponseWriter, r *http.Request) {
	var userID string = r.Context().Value(middleware.UserIDKey).(string)

	var req model.JoinCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.courseService.JoinCourse(r.Context(), userID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *CourseHandler) GetMyCourses(w http.ResponseWriter, r *http.Request) {
	var userID string = r.Context().Value(middleware.UserIDKey).(string)

	response, err := h.courseService.GetCoursesByOwner(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *CourseHandler) DeleteCourse(w http.ResponseWriter, r *http.Request) {
	var userID string = r.Context().Value(middleware.UserIDKey).(string)
	var courseID string = chi.URLParam(r, "id")

	if err := h.courseService.DeleteCourse(r.Context(), userID, courseID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}