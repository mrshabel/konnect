package handler

import (
	"konnect/internal/logger"
	"konnect/internal/model"
	"konnect/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SwipeHandler struct {
	swipeService *service.SwipeService
	logger       *zap.Logger
}

func NewSwipeHandler(swipeService *service.SwipeService, logger *logger.Logger) *SwipeHandler {
	return &SwipeHandler{
		swipeService: swipeService,
		logger:       logger.With(zap.String("component", "swipe_handler")),
	}
}

// CreateSwipe godoc
// @Summary Create a swipe
// @Description Create a new swipe (like or pass) on another user. Returns a match if one is created.
// @Tags swipes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.CreateSwipeRequest true "Swipe data"
// @Success 201 {object} model.SuccessResponse{data=model.SwipeResponse} "Swipe created successfully"
// @Failure 400,401,500 {object} model.ErrorResponse
// @Router /swipes [post]
func (h *SwipeHandler) CreateSwipe(c *gin.Context) {
	var req model.CreateSwipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "Invalid swipe data", Detail: err.Error()})
		return
	}

	user, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Message: "Unauthorized"})
		return
	}

	swipe := &model.Swipe{
		SwiperID:  user.ID,
		SwipeeID:  req.SwipeeID,
		SwipeType: req.SwipeType,
	}

	swipe, match, err := h.swipeService.CreateSwipe(swipe)
	if err != nil {
		if err == service.ErrAlreadySwiped || err == service.ErrSelfSwipe {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Failed to create swipe"})
		return
	}

	// send notification on match creation to the profile that was matched
	if match != nil {
		h.logger.Info("Match created, sending notification to matched party")
		// get swipe details with user preloaded
		swipe, err = h.swipeService.GetSwipeByID(swipe.ID)
		if err == nil {
			if err := h.swipeService.SendMatchNotification(swipe); err != nil {
				h.logger.Error("Failed to send match notification", zap.Error(err))
			}
		} else {
			h.logger.Error("Failed to get swipe details for sending match notification", zap.String("swiper_id", swipe.SwiperID.String()), zap.String("swipee_id", swipe.SwipeeID.String()), zap.Error(err))
		}

	}

	c.JSON(http.StatusCreated, model.SuccessResponse{
		Message: "Swipe created successfully",
		Data:    model.SwipeResponse{Swipe: *swipe, Match: match},
	})
}

// GetUserSwipeHistory godoc
// @Summary Get swipe history
// @Description Get all paginated swipes performed by a given user
// @Tags swipes
// @Produce json
// @Security BearerAuth
// @Param limit query number false "Limit" default(20)
// @Param offset query number false "Offset" default(0)
// @Success 200 {object} model.SuccessResponse{data=[]model.Swipe} "Swipe history retrieved successfully"
// @Failure 400,401,500 {object} model.ErrorResponse
// @Router /swipes/me [get]
func (h *SwipeHandler) GetUserSwipeHistory(c *gin.Context) {
	var query model.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Message: "Invalid query parameters",
			Detail:  err.Error(),
		})
		return
	}

	user, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Message: "Unauthorized"})
		return
	}

	swipes, err := h.swipeService.GetSwipeHistory(user.ID, query.Limit, query.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Failed to get swipe history for user"})
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse{Message: "Swipe history retrieved successfully", Data: swipes})
}
