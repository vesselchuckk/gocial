package server

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/vesselchuckk/go-social/internal/store"
	"net/http"
	"strings"
)

func (s *Server) BasicAuth() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				s.unauthorizedBasicError(w, r, fmt.Errorf("authorization header is missing"))
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Basic" {
				s.unauthorizedBasicError(w, r, fmt.Errorf("authorization header is invalid"))
				return
			}

			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				s.unauthorizedBasicError(w, r, err)
				return
			}

			username := s.Config.AdminUsername
			pass := s.Config.AdminPassword

			credentials := strings.SplitN(string(decoded), ":", 2)
			if len(credentials) != 2 || credentials[0] != username || credentials[1] != pass {
				s.unauthorizedBasicError(w, r, fmt.Errorf("invalid credentials"))
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}

func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			s.unauthorizedError(w, r, fmt.Errorf("authorization header is missing"))
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			s.unauthorizedError(w, r, fmt.Errorf("authorization header is invalid"))
			return
		}

		token := parts[1]
		jwtToken, err := s.JWTAuth.ValidateToken(token)
		if err != nil {
			s.unauthorizedError(w, r, err)
			return
		}

		claims, _ := jwtToken.Claims.(jwt.MapClaims)

		rawSub, ok := claims["sub"]
		if !ok {
			s.unauthorizedError(w, r, err)
			return
		}

		subClaim, ok := rawSub.(string)
		if !ok {
			s.unauthorizedError(w, r, err)
			return
		}

		userID, err := uuid.Parse(subClaim)
		if err != nil {
			s.unauthorizedError(w, r, err)
			return
		}

		ctx := r.Context()

		user, err := s.getUser(ctx, userID)
		if err != nil {
			s.unauthorizedError(w, r, err)
			return
		}

		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) checkPostOwnership(role string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromCtx(r)
		post := getPostFromCtx(r)

		if post.UserID == user.ID {
			next.ServeHTTP(w, r)
			return
		}

		allowed, err := s.checkRole(r.Context(), user, role)
		if err != nil {
			s.internalServerError(w, r, err)
			return
		}

		if !allowed {
			s.forbiddenResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) checkRole(ctx context.Context, user *store.User, roleName string) (bool, error) {
	role, err := s.Store.Roles.GetByName(ctx, roleName)
	if err != nil {
		return false, nil
	}

	return user.Role.Level >= role.Level, nil
}

func (s *Server) getUser(ctx context.Context, userID uuid.UUID) (*store.User, error) {
	user, err := s.Redis.Users.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		user, err = s.Store.Users.GetByID(ctx, userID)
		if err != nil {
			return nil, err
		}

		if err := s.Redis.Users.Set(ctx, user); err != nil {
			return nil, err
		}
	}

	return user, nil
}
