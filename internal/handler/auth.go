package handler

import (
	"fmt"
	"konnect/internal/model"
	"konnect/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// GoogleLogin godoc
// @Summary Initiate Google OAuth login
// @Description Redirects user to Google OAuth consent screen
// @Tags auth
// @Produce json
// @Success 307 "Redirect to Google OAuth"
// @Failure 400 {object} model.ErrorResponse
// @Router /auth/google/login [get]
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	q := c.Request.URL.Query()
	// register google oauth provider and begin authentication
	q.Add("provider", "google")
	c.Request.URL.RawQuery = q.Encode()
	gothic.BeginAuthHandler(c.Writer, c.Request)
}

// GoogleCallback godoc
// @Summary Google OAuth callback handler
// @Description Handles OAuth callback from Google and returns JWT token
// @Tags auth
// @Produce json
// @Param code query string true "Authorization code from Google"
// @Param state query string true "State parameter for CSRF protection"
// @Success 200 {object} model.SuccessResponse{data=model.AuthResponse} "Login successful"
// @Failure 400,401,500 {object} model.ErrorResponse
// @Router /auth/google/callback [get]
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	q := c.Request.URL.Query()
	q.Add("provider", "google")
	c.Request.URL.RawQuery = q.Encode()

	// complete the authentication process
	gothUser, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to complete auth: %w", err))
		return
	}

	// upsert user
	dbUser, err := h.authService.UpsertUserFromProvider(gothUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Failed to login user"})
		return
	}

	// get user with profile
	user, err := h.authService.GetUserByID(dbUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Failed to login user"})
		return
	}

	// generate tokens
	token, err := h.authService.GenerateAccessToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Failed to generate access token"})
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse{Message: "Login successful", Data: model.AuthResponse{
		Token: token,
		User:  *user,
	}})
}
