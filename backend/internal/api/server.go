package api

import (
	"net/http"

	"tfm/backend/internal/radar"
)

// Server holds the HTTP handler dependencies.
type Server struct {
	session *radar.Session
}

// New returns the HTTP handler for the API.
func New(session *radar.Session) http.Handler {
	s := &Server{session: session}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/connect", s.handleConnect)
	mux.HandleFunc("POST /api/disconnect", s.handleDisconnect)
	mux.HandleFunc("GET /api/status", s.handleStatus)
	mux.HandleFunc("GET /api/info", s.handleInfo)
	mux.HandleFunc("GET /api/params", s.handleParams)
	mux.HandleFunc("PUT /api/params", s.handleParamsPut)
	return mux
}
