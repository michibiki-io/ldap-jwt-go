package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/michibiki-io/ldap-jwt-go/model"
	"github.com/michibiki-io/ldap-jwt-go/service"
	"github.com/michibiki-io/ldap-jwt-go/utility"
)

func Authorize(c *gin.Context) {
	if c.Request.Method != "POST" {
		c.String(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	authModel := model.Auth{}

	if err := c.Bind(&authModel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and Password are required"})
		return
	}

	userService := service.UserService{}

	if userModel, err := userService.Authorize(&authModel); err != nil {
		c.String(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	} else if tokenSet, expire_in, err := userService.CreateAuth(&userModel); err != nil {
		statusCode, message := errorToHttpStatus(err)
		c.JSON(statusCode, message)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"access_token":  tokenSet.AccessToken.Token,
			"refresh_token": tokenSet.RefreshToken.Token,
			"expire_in":     expire_in.AccessToken,
			"token_type":    "Bearer"})
	}
}

func Verify(c *gin.Context) {
	if c.Request.Method != "POST" {
		c.String(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	mapToken := map[string]string{}
	if err := c.ShouldBindJSON(&mapToken); err != nil {
		c.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}

	userService := service.UserService{}

	if accessToken, ok := mapToken["access_token"]; !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "access_token is required."})
	} else if userModel, expire_in, err := userService.VerifyAuth(accessToken); err != nil {
		statusCode, message := errorToHttpStatus(err)
		c.JSON(statusCode, message)
	} else {
		c.JSON(http.StatusOK, gin.H{"user": userModel, "expire_in": expire_in})
	}
}

func Refresh(c *gin.Context) {
	if c.Request.Method != "POST" {
		c.String(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	mapToken := map[string]string{}
	if err := c.ShouldBindJSON(&mapToken); err != nil {
		c.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}

	userService := service.UserService{}

	if refreshToken, ok := mapToken["refresh_token"]; !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh_token is required."})
	} else if tokenSet, expire_in, err := userService.RefreshAuth(refreshToken); err != nil {
		statusCode, message := errorToHttpStatus(err)
		c.JSON(statusCode, message)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"access_token":  tokenSet.AccessToken.Token,
			"refresh_token": tokenSet.RefreshToken.Token,
			"expire_in":     expire_in,
			"token_type":    "Bearer"})
	}
}

func Deauthorize(c *gin.Context) {
	if c.Request.Method != "POST" {
		c.String(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	mapToken := map[string]string{}
	if err := c.ShouldBindJSON(&mapToken); err != nil {
		c.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}

	userService := service.UserService{}

	if accessToken, ok := mapToken["access_token"]; !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "access_token is required."})
	} else if err := userService.DeleteAuth(accessToken); err != nil {
		statusCode, message := errorToHttpStatus(err)
		c.JSON(statusCode, message)
	} else {
		c.JSON(http.StatusOK, gin.H{"result": true})
	}
}

func errorToHttpStatus(error error) (statusCode int, message gin.H) {
	if error, ok := error.(*utility.Error); ok {
		errNo := error.No()

		statusCode = http.StatusUnauthorized
		message = gin.H{"error": error.Error(), "no": errNo}

		switch errNo {
		case utility.Unauthorized:
			statusCode = http.StatusUnauthorized
		case utility.UnexpectedSigningMethod:
			statusCode = http.StatusUnprocessableEntity
		case utility.Forbidden:
			statusCode = http.StatusForbidden
		case utility.Expired:
			statusCode = http.StatusUnauthorized
		default:
			statusCode = http.StatusUnauthorized
		}
	} else {
		statusCode = http.StatusInternalServerError
		message = gin.H{"error": error.Error()}
	}

	return
}
