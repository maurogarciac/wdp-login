package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	d "frontend/internal/domain"
	auth "frontend/internal/middleware"
	"frontend/internal/services"
	"frontend/internal/templates"

	"go.uber.org/zap"
)

type LoginHandler struct {
	ctx context.Context
	lg  *zap.SugaredLogger
	b   *services.BackendService
}

func NewLoginHandler(context context.Context, logger *zap.SugaredLogger, backend *services.BackendService) *LoginHandler {
	return &LoginHandler{
		ctx: context,
		lg:  logger,
		b:   backend,
	}
}

func (h LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	c := templates.LogIn()

	switch r.Method {

	case http.MethodGet:

		if auth.GetAccessToken(r) != "" {
			http.Redirect(w, r, "/home/", http.StatusMovedPermanently)
		}

		pageRender("login", c, false, h.lg, w, r)

	case http.MethodPost:

		err := r.ParseForm()
		if err != nil {
			h.lg.Error(err)
		}

		// Maybe hash+salt this later if I have time

		login_creds := d.User{
			Username: r.FormValue("username"),
			Password: r.FormValue("password"),
		}

		h.lg.Info("User candidate: ", login_creds)

		res, err := h.b.PostLogin(h.ctx, login_creds)
		if err != nil {
			h.lg.Error("Could not authenticate user")
		}

		auth.SetTokenCookie(
			auth.AccessTokenCookieName,
			res.AccessToken,
			time.Now().Add(15*time.Minute), w)

		auth.SetTokenCookie(
			auth.RefreshTokenCookieName,
			res.RefreshToken,
			time.Now().Add(24*time.Hour), w)

		http.Redirect(w, r, "/home", http.StatusMovedPermanently)
	default:
		fmt.Fprintf(w, "only get and post methods are supported")
		return
	}
}
