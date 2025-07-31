package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rendiffdev/ffprobe-api/internal/errors"
	"github.com/rendiffdev/ffprobe-api/internal/services"
)

// AdminHandler handles administrative API endpoints
type AdminHandler struct {
	userService *services.UserService
	logger      zerolog.Logger
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(userService *services.UserService, logger zerolog.Logger) *AdminHandler {
	return &AdminHandler{
		userService: userService,
		logger:      logger,
	}
}

// UpdateUserRoleRequest represents a request to update a user's role
type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin user pro premium"`
}

// GetUserRole retrieves a user's current role
// @Summary Get user role
// @Description Get the role of a specific user
// @Tags admin
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/admin/users/{id}/role [get]
func (h *AdminHandler) GetUserRole(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", userIDStr).Msg("Invalid user ID")
		errors.ValidationError(c, "Invalid user ID", err.Error())
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to get user")
		errors.NotFound(c, "User not found", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"status":  user.Status,
	})
}

// UpdateUserRole updates a user's role
// @Summary Update user role
// @Description Update the role of a specific user
// @Tags admin
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body UpdateUserRoleRequest true "Update role request"
// @Success 200 {object} gin.H
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/admin/users/{id}/role [put]
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", userIDStr).Msg("Invalid user ID")
		errors.ValidationError(c, "Invalid user ID", err.Error())
		return
	}

	var req UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Invalid request body")
		errors.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	// Get current admin user from context
	adminID, exists := c.Get("user_id")
	if !exists {
		errors.Unauthorized(c, "Admin user ID not found in context", "")
		return
	}

	// Prevent self-role change from admin to lower role
	if adminIDStr, ok := adminID.(string); ok {
		if adminUUID, _ := uuid.Parse(adminIDStr); adminUUID == userID && req.Role != "admin" {
			errors.ValidationError(c, "Cannot change your own admin role to a lower role", "")
			return
		}
	}

	// Update user role
	if err := h.userService.UpdateUserRole(c.Request.Context(), userID, req.Role); err != nil {
		h.logger.Error().Err(err).Str("user_id", userID.String()).Str("role", req.Role).Msg("Failed to update user role")
		errors.InternalError(c, "Failed to update user role", err.Error())
		return
	}

	// Get updated user
	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to get updated user")
		errors.InternalError(c, "Role updated but failed to retrieve user details", err.Error())
		return
	}

	h.logger.Info().
		Str("user_id", userID.String()).
		Str("new_role", req.Role).
		Str("admin_id", adminID.(string)).
		Msg("User role updated")

	c.JSON(http.StatusOK, gin.H{
		"message": "User role updated successfully",
		"user": gin.H{
			"user_id": user.ID,
			"email":   user.Email,
			"role":    user.Role,
			"status":  user.Status,
		},
	})
}

// ListUsers lists all users with pagination
// @Summary List users
// @Description List all users with optional role filtering
// @Tags admin
// @Accept json
// @Produce json
// @Param role query string false "Filter by role" Enums(admin, user, pro, premium)
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} gin.H
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/admin/users [get]
func (h *AdminHandler) ListUsers(c *gin.Context) {
	// Parse query parameters
	role := c.Query("role")
	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Get users
	users, total, err := h.userService.ListUsers(c.Request.Context(), role, limit, offset)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list users")
		errors.InternalError(c, "Failed to list users", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users":  users,
		"total":  total,
		"limit":  limit,
		"offset": offset,
		"count":  len(users),
	})
}

// GetSystemStats returns system statistics
// @Summary Get system statistics
// @Description Get system-wide statistics and metrics
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Failure 500 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /api/v1/admin/stats [get]
func (h *AdminHandler) GetSystemStats(c *gin.Context) {
	stats, err := h.userService.GetSystemStats(c.Request.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to get system stats")
		errors.InternalError(c, "Failed to get system statistics", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
		"timestamp": time.Now().Unix(),
	})
}