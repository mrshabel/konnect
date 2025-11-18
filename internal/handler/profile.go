package handler

import (
	"konnect/internal/logger"
	"konnect/internal/model"
	"konnect/internal/service"
	"konnect/internal/util"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ProfileHandler struct {
	profileService    *service.ProfileService
	cloudinaryService *service.CloudinaryService
	logger            *zap.Logger
}

func NewProfileHandler(profileService *service.ProfileService, cloudinaryService *service.CloudinaryService, logger *logger.Logger) *ProfileHandler {
	return &ProfileHandler{
		profileService:    profileService,
		cloudinaryService: cloudinaryService,
		logger:            logger.With(zap.String("component", "profile_handler")),
	}
}

// CreateProfile godoc
// @Summary Create user profile
// @Description Create a new profile for the authenticated user
// @Tags profiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.CreateProfileRequest true "Profile data"
// @Success 201 {object} model.SuccessResponse{data=model.Profile} "Profile created successfully"
// @Failure 400,401,500 {object} model.ErrorResponse
// @Router /profiles [post]
func (h *ProfileHandler) CreateProfile(c *gin.Context) {
	// validate model and date
	var req model.CreateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Message: "Invalid profile data",
			Detail:  err.Error(),
		})
		return
	}

	if !model.ValidateInterests(req.Interests) {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Message: "Unknown interest provided in profile data",
		})
		return
	}

	dob, err := util.ValidateDate(req.DOB)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Message: "Invalid profile data",
			Detail:  err.Error(),
		})
		return
	}

	// Get authenticated user
	user, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Message: "Unauthorized"})
		return
	}

	profile := &model.Profile{
		UserID:             user.ID,
		Fullname:           req.Fullname,
		Interests:          req.Interests,
		Bio:                req.Bio,
		DOB:                dob,
		Gender:             req.Gender,
		IsGenderPublic:     req.IsGenderPublic,
		RelationshipIntent: req.RelationshipIntent,
		Latitude:           req.Latitude,
		Longitude:          req.Longitude,
	}

	if err := h.profileService.CreateProfile(profile); err != nil {
		if err == service.ErrProfileExists {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "Profile already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Failed to create profile"})
		return
	}

	c.JSON(http.StatusCreated, model.SuccessResponse{
		Message: "Profile created successfully",
		Data:    profile,
	})
}

// GetCurrentUserProfile godoc
// @Summary Get current user profile
// @Description Get profile of currently logged in user
// @Tags profiles
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.SuccessResponse{data=model.Profile} "Profile retrieved successfully"
// @Failure 400,401,404,500 {object} model.ErrorResponse
// @Router /profiles/me [get]
func (h *ProfileHandler) GetCurrentUserProfile(c *gin.Context) {
	user, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Message: "Unauthorized"})
		return
	}

	profile, err := h.profileService.GetProfileByUserID(user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Profile not found"})
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse{Message: "Profile retrieved successfully", Data: profile})
}

// GetNearbyProfiles godoc
// @Summary Get nearby profiles
// @Description Get profiles within specified radius of coordinates
// @Tags profiles
// @Produce json
// @Security BearerAuth
// @Param lat query number true "Latitude"
// @Param lng query number true "Longitude"
// @Param radius query number false "Radius in meters" default(5000)
// @Param limit query number false "Limit" default(20)
// @Param offset query number false "Offset" default(0)
// @Success 200 {object} model.SuccessResponse{data=[]model.Profile} "Nearby profiles retrieved successfully"
// @Failure 400,401,500 {object} model.ErrorResponse
// @Router /profiles/nearby [get]
func (h *ProfileHandler) GetNearbyProfiles(c *gin.Context) {
	var query model.GetNearbyProfilesRequest
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

	profiles, err := h.profileService.GetNearbyProfiles(user.ID, query.Lat, query.Lng, query.Radius, query.Offset, query.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Failed to get nearby profiles"})
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse{Message: "Nearby profiles retrieved successfully", Data: profiles})
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get profile by ID
// @Tags profiles
// @Produce json
// @Security BearerAuth
// @Param id path string true "Profile ID"
// @Success 200 {object} model.SuccessResponse{data=model.Profile} "Profile retrieved successfully"
// @Failure 400,401,404,500 {object} model.ErrorResponse
// @Router /profiles/{id} [get]
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	var param model.IDParam
	if err := c.ShouldBindUri(&param); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "Invalid profile ID"})
		return
	}

	profile, err := h.profileService.GetProfile(param.GetID())
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Profile not found"})
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse{Message: "Profile retrieved successfully", Data: profile})
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update authenticated user's profile
// @Tags profiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body model.UpdateProfileRequest true "Profile update data"
// @Success 200 {object} model.SuccessResponse{data=model.Profile} "Profile updated successfully"
// @Failure 400,401,404,500 {object} model.ErrorResponse
// @Router /profiles [patch]
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	var req model.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Message: "Invalid profile data",
			Detail:  err.Error(),
		})
		return
	}

	user, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Message: "Unauthorized"})
		return
	}

	// payload to be updated
	data, err := req.Compose()
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Message: "Invalid profile data",
			Detail:  err.Error(),
		})
		return
	}

	// get validate interests if provided
	if req.Interests != nil {
		if len(req.Interests) != 0 && !model.ValidateInterests(req.Interests) {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Message: "Unknown interest provided in profile data",
			})
			return
		}
	}

	profile, err := h.profileService.UpdateProfileByUserID(user.ID, data)
	if err != nil {
		h.logger.Error("failed to update profile", zap.Error(err), zap.String("user_id", user.ID.String()))
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse{Message: "Profile updated successfully", Data: profile})
}

// UploadProfilePhoto godoc
// @Summary Upload profile photo
// @Description Upload and verify user profile photo
// @Tags profiles
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param photo formData file true "Profile photo"
// @Success 200 {object} model.SuccessResponse{data=model.Profile} "Profile photo uploaded successfully"
// @Failure 400,401,500 {object} model.ErrorResponse
// @Router /profiles/photo [post]
func (h *ProfileHandler) UploadProfilePhoto(c *gin.Context) {
	user, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Message: "Unauthorized"})
		return
	}

	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Message: "No photo provided",
			Detail:  err.Error(),
		})
		return
	}

	if _, err := h.profileService.GetProfileByUserID(user.ID); err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: "Profile not found"})
		return
	}

	// upload photo with userid as filename and get URL
	photoURL, publicID, err := h.cloudinaryService.UploadImage(c.Request.Context(), file, "profile-photos", user.ID.String())
	if err != nil {
		h.logger.Error("failed to upload photo",
			zap.Error(err),
			zap.String("user_id", user.ID.String()))
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Message: "Failed to upload photo",
		})
		return
	}

	// update db with details
	profile, err := h.profileService.UpdateProfileByUserID(user.ID, &model.Profile{PhotoURL: &photoURL, PhotoPublicID: &publicID})
	if err != nil {
		h.logger.Error("failed to update profile",
			zap.Error(err),
			zap.String("user_id", user.ID.String()),
			zap.String("photo_url", photoURL),
			zap.String("public_id", publicID))

		// delete image in background
		go func() {
			if err := h.cloudinaryService.DeleteImage(c.Request.Context(), publicID); err != nil {
				h.logger.Error("failed to delete photo in background after profile update failed",
					zap.Error(err),
					zap.String("public_id", publicID),
					zap.String("photo_url", photoURL))
			}
		}()
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Failed to update user profile"})
		return
	}

	c.JSON(http.StatusOK, model.SuccessResponse{
		Message: "Profile photo uploaded successfully",
		Data:    profile,
	})
}
