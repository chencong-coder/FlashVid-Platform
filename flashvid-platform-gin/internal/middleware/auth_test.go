package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"flashvid-platform-gin/api"
	"flashvid-platform-gin/internal/middleware"
	jwtutil "flashvid-platform-gin/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func TestOptionalAuthAllowsAnonymousRequests(t *testing.T) {
	router := optionalAuthTestRouter(t)
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("anonymous request status = %d, want %d", response.Code, http.StatusNoContent)
	}
}

func TestOptionalAuthInjectsValidUserID(t *testing.T) {
	router := optionalAuthTestRouter(t)
	token, err := jwtutil.GenAccessToken(largeTestUserID, "tester")
	if err != nil {
		t.Fatalf("generate access token: %v", err)
	}
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("authenticated request status = %d, want %d", response.Code, http.StatusAccepted)
	}
}

func TestOptionalAuthRejectsInvalidToken(t *testing.T) {
	router := optionalAuthTestRouter(t)
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer invalid-token")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	var body api.ResponseData[any]
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body.Code != api.CodeInvalidToken {
		t.Fatalf("response code = %d, want %d", body.Code, api.CodeInvalidToken)
	}
}

const largeTestUserID int64 = 9_007_199_254_740_993

func optionalAuthTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	config := viper.New()
	config.Set("jwt.access_secret", "optional-auth-test-secret")
	config.Set("jwt.refresh_secret", "optional-auth-test-refresh-secret")
	config.Set("jwt.access_expire_seconds", 60)
	config.Set("jwt.refresh_expire_seconds", 120)
	jwtutil.MustInit(config)

	router := gin.New()
	router.GET("/", middleware.OptionalAuth(), func(c *gin.Context) {
		value, exists := c.Get(middleware.CtxKeyUserID)
		if !exists {
			c.Status(http.StatusNoContent)
			return
		}
		userID, ok := value.(int64)
		if !ok || userID != largeTestUserID {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusAccepted)
	})
	return router
}
