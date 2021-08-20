package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"time"
	"url-shortener/customError"
	"url-shortener/service"
	"url-shortener/validate"
)

const adminToken = "@dmIn"

type Controller interface {
	Shorten(ctx *gin.Context)
	Redirect(ctx *gin.Context)
	GetUrls(ctx *gin.Context)
	DeleteUrl(ctx *gin.Context)
}

type controller struct {
	service service.Service
}

type Header struct {
	Token string `header:"Token"`
}

type Response struct {
	Code    uint64      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func New(service service.Service) Controller {
	return &controller{
		service,
	}
}

func (c *controller) Shorten(ctx *gin.Context) {
	// Receive input
	var input struct {
		Url    string `json:"url"`
		Expiry string `json:"expiry"`
	}

	// Bind response body to input
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusOK, customError.ValidationError{
			Code:    1,
			Message: fmt.Sprintf("failed to handle input, err: %v", err),
		})
		return
	}

	// Validate url
	uri, err := url.ParseRequestURI(input.Url)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, customError.ValidationError{
			Code:    1,
			Message: fmt.Sprintf("failed to handle url input, err: %v", err),
		})
		return
	}

	// Check a url whether it is in blacklist
	err = validate.CheckBlackList(uri.String())
	if err != nil {
		ctx.JSON(http.StatusBadRequest, customError.ValidationError{
			Code:    1,
			Message: fmt.Sprintf("failed to handle url input, err: %v", err),
		})
		return
	}

	// Convert expiry to time type
	var expiry *time.Time
	if input.Expiry != "" {
		(*expiry), err = time.Parse(time.RFC3339, input.Expiry)
		if err != nil {
			ctx.JSON(http.StatusOK, customError.ValidationError{
				Code:    1,
				Message: fmt.Sprintf("failed to parse expiry, err: %v", err),
			})
			return
		}
	}

	shortCode, err := c.service.Encode(ctx, uri.String(), expiry)
	if err != nil {
		ctx.JSON(http.StatusOK, customError.InternalError{
			Code:    2,
			Message: fmt.Sprintf("internal error, err: %v", err),
		})
		return
	}
	ctx.JSON(http.StatusOK, &Response{
		Code:    0,
		Message: "success",
		Data:    shortCode,
	})
}

func (c *controller) Redirect(ctx *gin.Context) {
	// Receive input
	shortCode := ctx.Param("shortCode")

	fullUrl, err := c.service.Decode(ctx, shortCode)
	if err != nil {
		if ierr, ok := err.(*customError.InternalError); ok {
			ctx.JSON(ierr.HTTPStatusCode, customError.InternalError{
				Code:    2,
				Message: err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusNotFound, customError.InternalError{
			Code:    2,
			Message: fmt.Sprintf("internal error, err: %v", err),
		})
		return
	}
	ctx.Redirect(http.StatusFound, fullUrl)
}

func (c *controller) GetUrls(ctx *gin.Context) {
	// Receive input
	h := Header{}
	if err := ctx.ShouldBindHeader(&h); err != nil {
		ctx.JSON(http.StatusForbidden, customError.InternalError{
			Code:    2,
			Message: fmt.Sprintf("failed to access this api"),
		})
		return
	}
	if h.Token != adminToken {
		ctx.JSON(http.StatusForbidden, customError.InternalError{
			Code:    2,
			Message: fmt.Sprintf("failed to access this api"),
		})
		return
	}

	shortCode := ctx.Query("shortCode")
	fullUrl := ctx.Query("fullUrl")

	var pointerToShortCode *string
	var pointerToFullUrl *string
	if shortCode != "" {
		pointerToShortCode = &shortCode
	}
	if fullUrl != "" {
		pointerToFullUrl = &fullUrl
	}

	urlObjects, err := c.service.GetUrlObjects(ctx, pointerToShortCode, pointerToFullUrl)
	if err != nil {
		ctx.JSON(http.StatusNotFound, customError.InternalError{
			Code:    2,
			Message: fmt.Sprintf("internal error, err: %v", err),
		})
		return
	}
	ctx.JSON(http.StatusOK, &Response{
		Code:    0,
		Message: "success",
		Data:    urlObjects,
	})
}

func (c *controller) DeleteUrl(ctx *gin.Context) {
	// Receive input
	h := Header{}
	if err := ctx.ShouldBindHeader(&h); err != nil {
		ctx.JSON(http.StatusForbidden, customError.InternalError{
			Code:    2,
			Message: fmt.Sprintf("failed to access this api"),
		})
		return
	}
	if h.Token != adminToken {
		ctx.JSON(http.StatusForbidden, customError.InternalError{
			Code:    2,
			Message: fmt.Sprintf("failed to access this api"),
		})
		return
	}

	shortCode := ctx.Param("shortCode")

	_, err := c.service.DeleteUrl(ctx, shortCode)
	if err != nil {
		ctx.JSON(http.StatusNotFound, customError.InternalError{
			Code:    2,
			Message: fmt.Sprintf("internal error, err: %v", err),
		})
		return
	}
	ctx.JSON(http.StatusOK, &Response{
		Code:    0,
		Message: "success",
	})
}
