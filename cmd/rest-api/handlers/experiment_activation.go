package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

// CreateExperimentActivationHandler returns an http handler to create an activation for an experiment
func CreateExperimentActivationHandler(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "project_id")
		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(fmt.Sprintf("activation for project %s", projectID)))
	}
}
