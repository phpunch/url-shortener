package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	"url-shortener/customError"
	"url-shortener/service"
)

type Controller interface {
}

type controller struct {
	service service.Service
}

type Response struct {
	Code    uint64      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
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
