package server

import (
	"github.com/vesselchuckk/go-social/internal/auth"
	"github.com/vesselchuckk/go-social/internal/store"
	"github.com/vesselchuckk/go-social/internal/store/cache"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
)

func NewTestServer(t *testing.T) *Server {
	t.Helper()

	logger := zap.NewNop().Sugar()
	mockStore := store.NewMockStore()
	mockCache := cache.NewMockStore()

	testAuth := &auth.TestAuth{}

	return &Server{
		Logger:  logger,
		Store:   mockStore,
		Redis:   mockCache,
		JWTAuth: testAuth,
	}
}

func execute_req(req *http.Request, mux http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	return rr
}
