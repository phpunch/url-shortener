package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"time"
	"url-shortener/customError"
	"url-shortener/model"
	"url-shortener/service"
	"url-shortener/validate"
)

// Set fixed admin token
const adminToken = "@dmIn"

// Controller is an interface for APIs
type Controller interface {
	Shorten(ctx *gin.Context)
	Redirect(ctx *gin.Context)
	GetUrls(ctx *gin.Context)
	DeleteUrl(ctx *gin.Context)
}

// controller is an APIs management

type controller struct {
	service service.Service
}

// New is a constructor of controller
func New(service service.Service) Controller {
	return &controller{
		service,
	}
}

// Shorten godoc
// @Summary Shorten a specified url
// @Description shorten a specified url
// @Accept  json
// @Produce  json
// @Param ShortenInput body model.ShortenInput true "Input for shortening data"
// @Success 200 {object} Response
// @Failure 400,404 {object} customError.ValidationError
// @Router /shorten [post]
func (c *controller) Shorten(ctx *gin.Context) {
	// Receive input
	var input model.ShortenInput

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
	var pointerToExpiry *time.Time
	if input.Expiry != "" {
		expiry, err := time.Parse(time.RFC3339, input.Expiry)
		if err != nil {
			ctx.JSON(http.StatusOK, customError.ValidationError{
				Code:    1,
				Message: fmt.Sprintf("failed to parse expiry, err: %v", err),
			})
			return
		}
		pointerToExpiry = &expiry
	}

	shortCode, err := c.service.Encode(ctx, uri.String(), pointerToExpiry)
	if err != nil {
		ctx.JSON(http.StatusOK, customError.InternalError{
			Code:    2,
			Message: err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, &model.Response{
		Code:    0,
		Message: "success",
		Data:    shortCode,
	})
}

// Redirect godoc
// @summary Redirect to full url
// @description Redirect to full url using short code
// @produce json
// @Param shortCode path string true "Short Code"
// @Success 302 {object} Response
// @Failure 404 {object} customError.InternalError
// @router /{shortCode} [get]
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

// GetUrls godoc
// @summary Get all url for admin
// @description Get all url saved in database and can be filtered with a short code and a full url
// @produce json
// @Param token header string true "Admin token -> enter `@dmIn`"
// @Param shortCode query string false "Short Code"
// @Param fullUrl query string false "Full URL"
// @Success 200 {object} Response
// @Failure 400 {object} customError.InternalError
// @router /admin/urls [get]
func (c *controller) GetUrls(ctx *gin.Context) {
	// Receive input
	h := model.Header{}
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
	ctx.JSON(http.StatusOK, &model.Response{
		Code:    0,
		Message: "success",
		Data:    urlObjects,
	})
}

// GetUrls godoc
// @summary Get all url for admin
// @description Get all url saved in database and can be filtered with a short code and a full url
// @produce json
// @Param token header string true "Admin token -> enter `@dmIn`"
// @Param shortCode path string true "Short Code"
// @Success 200 {object} Response
// @Failure 403 {object} customError.InternalError
// @Failure 404 {object} customError.InternalError
// @router /{shortCode} [delete]
func (c *controller) DeleteUrl(ctx *gin.Context) {
	// Receive input
	h := model.Header{}
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
			Message: err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, &model.Response{
		Code:    0,
		Message: "success",
	})
}
