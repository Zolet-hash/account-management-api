package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCreateUserRejectsInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/users", New(nil).CreateUser)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/users", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != 400 {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
}

func TestBalancesRejectsInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/users/:userId/balances", New(nil).Balances)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/users/nope/balances", nil))
	if w.Code != 400 {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
}
