package response

import (
	"encoding/json"
	"net/http"
)

type envelope map[string]any

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func OK(w http.ResponseWriter, data any) { JSON(w, http.StatusOK, data) }

func Created(w http.ResponseWriter, data any) { JSON(w, http.StatusCreated, data) }

func NoContent(w http.ResponseWriter) { w.WriteHeader(http.StatusNoContent) }

func Error(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, envelope{"error": msg})
}

func BadRequest(w http.ResponseWriter, msg string)  { Error(w, http.StatusBadRequest, msg) }
func Unauthorized(w http.ResponseWriter, msg string) { Error(w, http.StatusUnauthorized, msg) }
func Forbidden(w http.ResponseWriter, msg string)   { Error(w, http.StatusForbidden, msg) }
func NotFound(w http.ResponseWriter, msg string)    { Error(w, http.StatusNotFound, msg) }
func Internal(w http.ResponseWriter, msg string)    { Error(w, http.StatusInternalServerError, msg) }
