package server

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/vesselchuckk/go-social/internal/store"
	"net/http"
	"strconv"
)

type postKey string
type userKey string

const postCtx postKey = "post"
const userCtx userKey = "user"

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

func WriteJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

func (s *Server) jsonResponse(w http.ResponseWriter, status int, data any) error {
	type envelope struct {
		Data any `json:"data"`
	}

	return WriteJSON(w, status, &envelope{Data: data})
}

func ReadJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1_048_578 // 1 mb
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)

	return decoder.Decode(data)
}

func WriteJSONError(w http.ResponseWriter, status int, message string) error {
	type envelope struct {
		Error string `json:"error"`
	}

	return WriteJSON(w, status, &envelope{Error: message})
}

func (s *Server) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	s.Logger.Errorw("internal server error", r.Method, "path", r.URL.Path, "error", err.Error())

	WriteJSONError(w, http.StatusInternalServerError, "error occurred on the server side")
}

func (s *Server) forbiddenResponse(w http.ResponseWriter, r *http.Request) {
	s.Logger.Warnw("forbidden", r.Method, "path", r.URL.Path, "error")

	WriteJSONError(w, http.StatusForbidden, "access restricted")
}

func (s *Server) badRequest(w http.ResponseWriter, r *http.Request, err error) {
	s.Logger.Warnf("bad request error", r.Method, "path", r.URL.Path, "error", err.Error())

	WriteJSONError(w, http.StatusBadRequest, err.Error())
}

func (s *Server) notFoundError(w http.ResponseWriter, r *http.Request, err error) {
	s.Logger.Warnf("not found error", r.Method, "path", r.URL.Path, "error", err.Error())

	WriteJSONError(w, http.StatusBadRequest, "resource not found")
}

func (s *Server) unauthorizedBasicError(w http.ResponseWriter, r *http.Request, err error) {
	s.Logger.Warnf("unauthorized", r.Method, "path", r.URL.Path, "error", err.Error())

	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	WriteJSONError(w, http.StatusUnauthorized, "resource not found")
}

func (s *Server) unauthorizedError(w http.ResponseWriter, r *http.Request, err error) {
	s.Logger.Warnf("unauthorized", r.Method, "path", r.URL.Path, "error", err.Error())

	WriteJSONError(w, http.StatusUnauthorized, "unauthorized")
}

func (s *Server) postContextFetch(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "postID")
		id, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			s.internalServerError(w, r, err)
			return
		}

		ctx := r.Context()

		post, err := s.Store.Posts.GetByID(ctx, id)
		if err != nil {
			s.notFoundError(w, r, err)
			return
		}

		ctx = context.WithValue(ctx, postCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) userContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawUserID := chi.URLParam(r, "userID")

		userID, err := uuid.Parse(rawUserID)
		if err != nil {
			s.badRequest(w, r, err)
			return
		}

		ctx := r.Context()

		user, err := s.Store.Users.GetByID(ctx, userID)
		if err != nil {
			s.internalServerError(w, r, err)
			return
		}

		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromCtx(r *http.Request) *store.Post {
	post, _ := r.Context().Value(postCtx).(*store.Post)
	return post
}

func getUserFromCtx(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCtx).(*store.User)
	return user
}
