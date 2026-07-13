package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/safescope/safescope/apps/api/internal/domain"
	"github.com/safescope/safescope/apps/api/internal/platform/security"
)

func TestAuthenticateRejectsMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(authenticate(security.JWTManager{Secret: []byte("test-secret"), TTL: time.Hour}))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}
}

func TestAuthenticateAcceptsValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := security.JWTManager{Secret: []byte("test-secret"), TTL: time.Hour}
	token, _, err := manager.Issue(&domain.User{ID: "u1", Role: domain.RoleAnalyst})
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.Use(authenticate(manager))
	router.GET("/", func(c *gin.Context) { c.JSON(http.StatusOK, actor(c)) })
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
}
