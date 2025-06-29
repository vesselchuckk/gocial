package server

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

const (
	testToken = "abc123"
)

func TestGetUser(t *testing.T) {
	srv := NewTestServer(t)
	mux := srv.Make()

	t.Run("should not allow unauthed req", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/v1/users/55e13943-a3c0-4ec5-a9bf-4a9d0883c9ca", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := execute_req(req, mux)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("should allow authed req", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/v1/users/55e13943-a3c0-4ec5-a9bf-4a9d0883c9ca", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", "Bearer "+testToken)

		rr := execute_req(req, mux)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
