package main

import (
	"net/http"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "ok",
		"env":     app.config.env,
		"version": VERSION,
	}

	if err := writeJSON(w, http.StatusOK, data); err != nil {
		_ = writeJSONError(w, http.StatusInternalServerError, err.Error())
	}
}
