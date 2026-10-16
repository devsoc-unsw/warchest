// receive HTTP requests and pass data to the service

package handler

import (
	"backend/db"
	"backend/internal/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	service service.EventService
}

func NewEventHandler(service service.EventService) *EventHandler {
	return &EventHandler{service: service}
}

// JSON from frontend
type createEventRequest struct {
	Name        string         `json:"event_name"`
	EventTime   time.Time      `json:"event_time"`
	Location    string         `json:"location"`
	Description string         `json:"description"`
	SocietyID   string         `json:"society_id"`
	Status      db.EventStatus `json:"status"`
	// Temporary input: use authenticated identity before enabling member access.
	CreatedBy string `json:"created_by"`
}

// Read the frontend JSON, create the Event, and return the result.
func (h *EventHandler) CreateEvent(c *gin.Context) {
	var req createEventRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	event, err := h.service.CreateEvent(c.Request.Context(), service.CreateEventInput{
		Name:        req.Name,
		EventTime:   req.EventTime,
		Location:    req.Location,
		Description: req.Description,
		SocietyID:   req.SocietyID,
		CreatedBy:   req.CreatedBy,
		Status:      req.Status,
	})

	if err != nil {
		if err == service.ErrCreateEvent {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id":          event.ID.String(),
		"type":        event.Type,
		"event_name":  event.Name,
		"society_id":  event.SocietyID,
		"status":      event.Status.EventStatus,
		"created_by":  event.CreatedBy.String(),
		"event_time":  event.EventTime,
		"location":    event.Location,
		"description": event.Description,
		"created_at":  event.CreatedAt.Time,
		"updated_at":  event.UpdatedAt.Time,
	})
}
