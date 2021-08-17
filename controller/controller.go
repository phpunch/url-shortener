package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	"url-shortener/customError"
	"url-shortener/service"
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

	// Convert expiry to time type
	var expiry *time.Time
	var err error
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

	shortCode, err := c.service.Encode(ctx, input.Url, expiry)
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
		ctx.JSON(http.StatusNotFound, customError.InternalError{
			Code:    2,
			Message: fmt.Sprintf("internal error, err: %v", err),
		})
		return
	}
	ctx.Redirect(http.StatusMovedPermanently, fullUrl)
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

	shortCode := ctx.Param("shortCode")
	fullUrl := ctx.Param("fullUrl")

	urlObjects, err := c.service.GetUrlObjects(ctx, &shortCode, &fullUrl)
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
