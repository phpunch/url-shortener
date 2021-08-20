package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"gotest.tools/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"url-shortener/mock"
	"url-shortener/model"
)

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

	var resp model.Response
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
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, output, w.HeaderMap.Get("Location"))
}
func TestGetUrlsRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	c, router := gin.CreateTestContext(httptest.NewRecorder())
	serv := mock.NewMockService(mockCtrl)
	ctrl := New(serv)

	mockedTime, _ := time.Parse(time.RFC3339Nano, "2021-08-20T22:06:32.6162088+07:00")
	output := []*model.UrlObject{
		&model.UrlObject{
			FullURL:   "http://www.facebook.com",
			ShortCode: "7XxYzjImrg6",
			Expiry:    &mockedTime,
			Hits:      2,
		},
		&model.UrlObject{
			FullURL:   "http://www.netflix.com",
			ShortCode: "4oEQByEsvg4",
			Hits:      25,
		},
	}
	expected := []interface{}{
		map[string]interface{}{
			"expiry":    "2021-08-20T22:06:32.6162088+07:00",
			"fullUrl":   "http://www.facebook.com",
			"hits":      float64(2),
			"shortCode": "7XxYzjImrg6",
		},
		map[string]interface{}{
			"fullUrl":   "http://www.netflix.com",
			"hits":      float64(25),
			"shortCode": "4oEQByEsvg4",
		},
	}
	serv.EXPECT().
		GetUrlObjects(gomock.Any(), nil, nil).
		Return(output, nil)

	router.GET("/admin/urls", ctrl.GetUrls)

	w := httptest.NewRecorder()

	// completeUrl := fmt.Sprintf("/admin/urls?shortCode=%s&fullUrl=%s", shortCode, fullUrl)
	c.Request, _ = http.NewRequest("GET", "/admin/urls", nil)
	c.Request.Header.Set("Token", adminToken)
	router.ServeHTTP(w, c.Request)

	var resp model.Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode, err: %v", err)
	}
	t.Log(resp)
	assert.Equal(t, 200, w.Code)
	assert.DeepEqual(t, expected, resp.Data)
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
