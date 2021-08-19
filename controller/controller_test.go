package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/mock"
)

func setupRouter(t *testing.T) *gin.Engine {
	mockCtrl := gomock.NewController(t)
	serv := mock.NewMockService(mockCtrl)
	ctrl := New(serv)

	router := gin.Default()
	router.POST("/shorten", ctrl.Shorten)
	router.GET("/:shortCode", ctrl.Redirect)
	router.GET("/admin/urls", ctrl.GetUrls)
	router.DELETE("/:shortCode", ctrl.DeleteUrl)
	router.Run(":8080")

	return router
}

func TestShortenRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	c, router := gin.CreateTestContext(httptest.NewRecorder())
	serv := mock.NewMockService(mockCtrl)
	ctrl := New(serv)

	output := "mockedShortCode"
	serv.EXPECT().
		Encode(gomock.Any(), "https://www.facebook.com", nil).
		Return(output, nil)

	router.POST("/shorten", ctrl.Shorten)

	w := httptest.NewRecorder()

	jsonBytes, _ := json.Marshal(map[string]string{
		"url": "https://www.facebook.com",
	})
	c.Request, _ = http.NewRequest("POST", "/shorten", bytes.NewReader(jsonBytes))
	router.ServeHTTP(w, c.Request)

	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode, err: %v", err)
	}
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, output, resp.Data)
}
func TestRedirectRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	c, router := gin.CreateTestContext(httptest.NewRecorder())
	serv := mock.NewMockService(mockCtrl)
	ctrl := New(serv)

	input := "mockedShortCode"
	output := "/mockedFullCode"
	serv.EXPECT().
		Decode(gomock.Any(), input).
		Return(output, nil)

	router.GET("/:shortCode", ctrl.Redirect)

	w := httptest.NewRecorder()

	c.Request, _ = http.NewRequest("GET", fmt.Sprintf("/%s", input), nil)
	router.ServeHTTP(w, c.Request)

	t.Log(w.HeaderMap.Get("Location"))
	assert.Equal(t, http.StatusMovedPermanently, w.Code)
	assert.Equal(t, output, w.HeaderMap.Get("Location"))
}
func TestDeleteUrlRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	c, router := gin.CreateTestContext(httptest.NewRecorder())
	serv := mock.NewMockService(mockCtrl)
	ctrl := New(serv)

	input := "mockedFullCode"
	serv.EXPECT().
		DeleteUrl(gomock.Any(), input).
		Return(true, nil)

	// register route
	router.GET("/:shortCode", ctrl.Redirect)
	router.DELETE("/:shortCode", ctrl.DeleteUrl)

	w := httptest.NewRecorder()

	c.Request, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/%s", input), nil)
	c.Request.Header.Set("Token", adminToken)
	router.ServeHTTP(w, c.Request)

	// validate output
	assert.Equal(t, http.StatusOK, w.Code)
}
