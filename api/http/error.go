package http

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents an error response body.
type ErrorResponse struct {
	Message string `json:"message"`
}

// Error makes ErrorResponse implement the error interface.
func (e *ErrorResponse) Error() string {
	return e.Message
}

// Common API errors.
var (
	ErrUnauthorized       = &ErrorResponse{Message: "No autorizado."}
	ErrForbidden          = &ErrorResponse{Message: "Acceso denegado."}
	ErrNotFound           = &ErrorResponse{Message: "Recurso no encontrado."}
	ErrMethodNotAllowed   = &ErrorResponse{Message: "Método no permitido."}
	ErrBadRequest         = &ErrorResponse{Message: "Solicitud inválida."}
	ErrInternalServer     = &ErrorResponse{Message: "Error interno del servidor."}
	ErrConflict           = &ErrorResponse{Message: "El recurso ya existe."}
	ErrUnprocessable      = &ErrorResponse{Message: "No se pudo procesar la solicitud."}
	ErrInvalidCredentials = &ErrorResponse{Message: "Credenciales inválidas."}
)

// sendJSON writes a JSON response with the given status code and data.
func sendJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, `{"message":"Error al codificar respuesta."}`, http.StatusInternalServerError)
		}
	}
}

// sendError writes a JSON error response with the given status code.
func sendError(w http.ResponseWriter, status int, err *ErrorResponse) {
	sendJSON(w, status, err)
}

// sendErrorMessage writes a JSON error response with a custom message.
func sendErrorMessage(w http.ResponseWriter, status int, message string) {
	sendError(w, status, &ErrorResponse{Message: message})
}
