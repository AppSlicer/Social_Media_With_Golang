package handlers

import (
	"math"
	"net/http"
	"strconv"

	"github.com/anonto42/nano-midea/backend/internal/models"
	"github.com/anonto42/nano-midea/backend/internal/repositories"
	"github.com/labstack/echo/v4"
)

// NotificationHandler handles notification-related HTTP requests
type NotificationHandler struct {
	notificationRepository repositories.NotificationRepository
	userRepository         repositories.UserRepository
}

// NewNotificationHandler creates a new NotificationHandler
func NewNotificationHandler(notifRepo repositories.NotificationRepository, userRepo repositories.UserRepository) *NotificationHandler {
	return &NotificationHandler{
		notificationRepository: notifRepo,
		userRepository:         userRepo,
	}
}

// RegisterNotificationRoutes registers notification routes
func (h *NotificationHandler) RegisterNotificationRoutes(g *echo.Group) {
	g.GET("/notifications", h.GetNotifications)
	g.GET("/notifications/grouped", h.GetGroupedNotifications)
	g.GET("/notifications/unread-count", h.GetUnreadCount)
	g.PUT("/notifications/:id/read", h.MarkAsRead)
	g.PUT("/notifications/read-all", h.MarkAllAsRead)
}

// EnrichedNotification includes actor info
type EnrichedNotification struct {
	models.Notification
	Actor models.UserCompact `json:"actor"`
}

func (h *NotificationHandler) enrichNotifications(notifications []models.Notification) []EnrichedNotification {
	enriched := make([]EnrichedNotification, len(notifications))
	userCache := make(map[uint]models.UserCompact)

	for i, n := range notifications {
		enriched[i] = EnrichedNotification{Notification: n}
		if actor, ok := userCache[n.ActorID]; ok {
			enriched[i].Actor = actor
		} else {
			user, err := h.userRepository.GetUserByID(n.ActorID)
			if err == nil {
				compact := user.ToCompact()
				userCache[n.ActorID] = compact
				enriched[i].Actor = compact
			}
		}
	}
	return enriched
}

// GetNotifications returns paginated notifications
func (h *NotificationHandler) GetNotifications(c echo.Context) error {
	currentUserID := getUserIDFromContext(c)
	if currentUserID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}

	notifications, total, err := h.notificationRepository.GetByRecipientID(currentUserID, page, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	enriched := h.enrichNotifications(notifications)

	return c.JSON(http.StatusOK, echo.Map{
		"success": true,
		"data": echo.Map{
			"notifications": enriched,
		},
		"meta": echo.Map{
			"currentPage":     page,
			"totalPages":      totalPages,
			"totalItems":      total,
			"itemsPerPage":    limit,
			"hasNextPage":     page < totalPages,
			"hasPreviousPage": page > 1,
		},
	})
}

// GetGroupedNotifications returns notifications grouped by time period
func (h *NotificationHandler) GetGroupedNotifications(c echo.Context) error {
	currentUserID := getUserIDFromContext(c)
	if currentUserID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	today, yesterday, thisWeek, older, err := h.notificationRepository.GetGrouped(currentUserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	unreadCount, _ := h.notificationRepository.GetUnreadCount(currentUserID)

	return c.JSON(http.StatusOK, echo.Map{
		"success": true,
		"data": echo.Map{
			"notifications": echo.Map{
				"today":     h.enrichNotifications(today),
				"yesterday": h.enrichNotifications(yesterday),
				"thisWeek":  h.enrichNotifications(thisWeek),
				"older":     h.enrichNotifications(older),
			},
			"unreadCount": unreadCount,
		},
	})
}

// GetUnreadCount returns the unread notification count
func (h *NotificationHandler) GetUnreadCount(c echo.Context) error {
	currentUserID := getUserIDFromContext(c)
	if currentUserID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	count, err := h.notificationRepository.GetUnreadCount(currentUserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{"success": true, "data": echo.Map{"count": count}})
}

// MarkAsRead marks a notification as read
func (h *NotificationHandler) MarkAsRead(c echo.Context) error {
	currentUserID := getUserIDFromContext(c)
	if currentUserID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	notifID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid notification ID")
	}

	if err := h.notificationRepository.MarkAsRead(uint(notifID)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{"success": true, "data": echo.Map{"success": true}})
}

// MarkAllAsRead marks all notifications as read
func (h *NotificationHandler) MarkAllAsRead(c echo.Context) error {
	currentUserID := getUserIDFromContext(c)
	if currentUserID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	if err := h.notificationRepository.MarkAllAsRead(currentUserID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{"success": true, "data": echo.Map{"success": true}})
}
