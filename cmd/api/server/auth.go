package server

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/vesselchuckk/go-social/internal/mails"
	"github.com/vesselchuckk/go-social/internal/store"
	"net/http"
	"time"
)

var (
	exp     = time.Hour * 24 * 3
	jwt_exp = time.Hour * 24 * 3
)

type UserWithToken struct {
	*store.User
	Token string `json:"token"`
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required,max=96"`
	Email    string `json:"email" validate:"required,max=96"`
	Password string `json:"password" validate:"required,min=8,max=16"`
}

type CreateTokenReq struct {
	Email    string `json:"email" validate:"required,email,max=96"`
	Password string `json:"password" validate:"required,min=8,max=16"`
}

func (s *Server) registerHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := ReadJSON(w, r, &req); err != nil {
		s.badRequest(w, r, err)
		return
	}

	if err := Validate.Struct(req); err != nil {
		s.badRequest(w, r, err)
		return
	}

	user := &store.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Role: store.Role{
			Name: "user",
		},
	}

	ctx := r.Context()

	plainToken := uuid.New().String()

	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	err := s.Store.Users.CreateAndInvite(ctx, user, hashToken, exp)
	if err != nil {
		switch err {
		case store.ErrDuplicateEmail:
			s.badRequest(w, r, err)
		case store.ErrDuplicateUsername:
			s.badRequest(w, r, err)
		default:
			s.internalServerError(w, r, err)
		}
	}

	userWithToken := UserWithToken{
		User:  user,
		Token: plainToken,
	}
	activationURL := fmt.Sprintf("%s/confirm/%s", s.Config.FrontendURL, plainToken)

	isProdEnv := s.Config.ENV == "production"
	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: activationURL,
	}

	status, err := s.Mailer.Send(mails.ActivationTemplate, user.Username, user.Email, vars, !isProdEnv)
	if err != nil {
		s.Logger.Errorw("error sending activation email", "email", err)

		//rollback
		if err := s.Store.Users.Delete(ctx, user.ID); err != nil {
			s.Logger.Errorw("error deleting user", "error", err)
		}

		s.internalServerError(w, r, err)
		return
	}

	s.Logger.Infow("Email sent with status code: %v", status)

	if err := s.jsonResponse(w, http.StatusCreated, userWithToken); err != nil {
		s.internalServerError(w, r, err)
	}
}

func (s *Server) activateUser(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	err := s.Store.Users.Activate(r.Context(), token)
	if err != nil {
		s.internalServerError(w, r, err)
	}

	if err := s.jsonResponse(w, http.StatusNoContent, ""); err != nil {
		s.internalServerError(w, r, err)
	}
}

func (s *Server) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateTokenReq
	if err := ReadJSON(w, r, &req); err != nil {
		s.badRequest(w, r, err)
		return
	}

	if err := Validate.Struct(req); err != nil {
		s.badRequest(w, r, err)
		return
	}

	user, err := s.Store.Users.GetByEmail(r.Context(), req.Email)
	if err != nil {
		s.unauthorizedError(w, r, err)
	}

	claims := jwt.MapClaims{
		"sub": user.ID.String(),
		"exp": time.Now().Add(jwt_exp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": s.Config.JWTiss,
		"aud": s.Config.JWTiss,
	}
	token, err := s.JWTAuth.GenerateToken(claims)
	if err != nil {
		s.internalServerError(w, r, err)
	}

	if err := s.jsonResponse(w, http.StatusCreated, token); err != nil {
		s.internalServerError(w, r, err)
	}
}
