package server

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/vesselchuckk/go-social/internal/store"
	"log"
	"net/http"
	"strconv"
)

// POSTS PAYLOAD
type CreatePostRequest struct {
	Title   string         `json:"title" validate:"required,max=100"`
	Content string         `json:"content" validate:"required,max=1000"`
	Tags    pq.StringArray `json:"tags"`
}

// USER PAYLOAD
type UpdatePostRequest struct {
	Title   *string `json:"title" validate:"omitempty,max=100"`
	Content *string `json:"content" validate:"omitempty,max=1000"`
}

type FollowRequest struct {
	UserID     uuid.UUID `json:"user_id"`
	FollowerID uuid.UUID `json:"follower_id"`
}

// HEALTH HANDLER

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "ok",
		"env":     s.Config.ENV,
		"version": s.Config.Version,
	}

	if err := s.jsonResponse(w, http.StatusOK, data); err != nil {
		s.internalServerError(w, r, err)
	}
}

// POSTS HANDLER

func (s *Server) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var req CreatePostRequest
	if err := ReadJSON(w, r, &req); err != nil {
		s.badRequest(w, r, err)
		return
	}

	if err := Validate.Struct(req); err != nil {
		s.badRequest(w, r, err)
		return
	}

	user := getUserFromCtx(r)

	post := &store.Post{
		Title:   req.Title,
		Content: req.Content,
		UserID:  user.ID,
	}

	ctx := r.Context()

	err := s.Store.Posts.CreatePost(ctx, post)
	if err != nil {
		s.internalServerError(w, r, err)
		log.Fatalf("error creating post: %v", err)
		return
	}

	if err := s.jsonResponse(w, http.StatusCreated, post); err != nil {
		s.badRequest(w, r, err)
		return
	}
}

func (s *Server) getPostByID(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	comments, err := s.Store.Comments.GetByPostID(r.Context(), post.ID)
	if err != nil {
		s.internalServerError(w, r, err)
		return
	}

	post.Comments = comments

	if err := s.jsonResponse(w, http.StatusOK, post); err != nil {
		s.internalServerError(w, r, err)
		return
	}
}

func (s *Server) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	var req UpdatePostRequest
	if err := ReadJSON(w, r, &req); err != nil {
		s.badRequest(w, r, err)
		return
	}

	if err := Validate.Struct(req); err != nil {
		s.badRequest(w, r, err)
		return
	}

	if req.Title != nil {
		post.Title = *req.Title
	}
	if req.Content != nil {
		post.Content = *req.Content
	}

	if err := s.Store.Posts.Update(r.Context(), post); err != nil {
		s.internalServerError(w, r, err)
		return
	}

	if err := s.jsonResponse(w, http.StatusOK, post); err != nil {
		s.internalServerError(w, r, err)
	}
}

func (s *Server) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "postID")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		s.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	if err := s.Store.Posts.Delete(ctx, id); err != nil {
		s.internalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) getUserHandler(w http.ResponseWriter, r *http.Request) {
	rawUserID := chi.URLParam(r, "userID")
	if rawUserID == "" {
		s.badRequest(w, r, errors.New("userID is required"))
		return
	}

	userID, err := uuid.Parse(rawUserID)
	if err != nil {
		s.badRequest(w, r, err)
		return
	}

	user, err := s.Redis.Users.Get(r.Context(), userID)
	if err != nil {
		s.internalServerError(w, r, err)
		return
	}

	if user == nil {
		user, err = s.Store.Users.GetByID(r.Context(), userID)
		if err != nil {
			s.notFoundError(w, r, err)
			return
		}
	}

	if err := s.jsonResponse(w, http.StatusOK, user); err != nil {
		s.internalServerError(w, r, err)
	}
}

func (s *Server) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followedUser := getUserFromCtx(r)
	followedID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		s.badRequest(w, r, err)
		return
	}

	ctx := r.Context()

	if err := s.Store.Followers.Follow(ctx, followedUser.ID, followedID); err != nil {
		s.internalServerError(w, r, err)
		return
	}

	if err := s.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		s.internalServerError(w, r, err)
	}
}

func (s *Server) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	unfollowedUser := getUserFromCtx(r)

	var req FollowRequest
	if err := ReadJSON(w, r, &req); err != nil {
		s.badRequest(w, r, err)
		return
	}

	ctx := r.Context()

	if err := s.Store.Followers.Unfollow(ctx, unfollowedUser.ID, req.UserID); err != nil {
		s.internalServerError(w, r, err)
		return
	}

	if err := s.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		s.internalServerError(w, r, err)
	}
}

func (s *Server) getUserFeed(w http.ResponseWriter, r *http.Request) {
	fq := store.PaginatedQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
	}

	fq, err := fq.Parse(r)
	if err != nil {
		s.badRequest(w, r, err)
		return
	}

	if err := Validate.Struct(fq); err != nil {
		s.badRequest(w, r, err)
		return
	}

	user := getUserFromCtx(r)

	ctx := r.Context()

	feed, err := s.Store.Posts.GetUserFeed(ctx, user, fq)
	if err != nil {
		s.internalServerError(w, r, err)
	}

	if err := s.jsonResponse(w, http.StatusOK, feed); err != nil {
		s.internalServerError(w, r, err)
	}
}
