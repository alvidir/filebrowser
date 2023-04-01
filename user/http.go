package user

import (
	"encoding/json"
	"net/http"

	fb "github.com/alvidir/filebrowser"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type UserHttpServer struct {
	app       *UserApplication
	router    *mux.Router
	logger    *zap.Logger
	uidHeader string
}

func NewUserHttpServer(app *UserApplication, logger *zap.Logger, authHeader string) *UserHttpServer {
	server := &UserHttpServer{
		router:    mux.NewRouter(),
		logger:    logger,
		uidHeader: authHeader,
	}

	server.router.HandleFunc("/", server.getProfileHandler).Methods(http.MethodGet)
	return server
}

func (server *UserHttpServer) Handler() http.Handler {
	return server.router
}

func (server *UserHttpServer) getProfileHandler(w http.ResponseWriter, r *http.Request) {
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
