package handler

import (
	"encoding/json"
	"net/http"

	"github.com/niyiayooluwa/geotas/internal/db"
	"github.com/niyiayooluwa/geotas/internal/middleware"
)



func MeHandler(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// get user_id from context (populated by AuthMiddleware)
		var userID string = r.Context().Value(middleware.UserIDKey).(string)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"user_id": userID,
		})
	}
}
