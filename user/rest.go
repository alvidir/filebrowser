package user

import (
	"encoding/json"
	"net/http"

	fb "github.com/alvidir/filebrowser"
	"go.uber.org/zap"
)

type UserRestService struct {
	app       *UserApplication
	handler   *http.ServeMux
	logger    *zap.Logger
	uidHeader string
}

func NewUserRestServer(app *UserApplication, logger *zap.Logger, authHeader string) *UserRestService {
	server := &UserRestService{
		app:       app,
		handler:   http.NewServeMux(),
		logger:    logger,
		uidHeader: authHeader,
	}

	server.handler.HandleFunc("/profile", server.getProfileHandler)
	return server
}

func (server *UserRestService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server.handler.ServeHTTP(w, r)
}

func (server *UserRestService) getProfileHandler(w http.ResponseWriter, r *http.Request) {
	uid, err := fb.GetUidFromHttpRequest(r, server.uidHeader, server.logger)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	profile, err := server.app.GetProfile(r.Context(), uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(profile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(body); err != nil {
		server.logger.Error("writing http response",
			zap.Error(err))
	}
}
