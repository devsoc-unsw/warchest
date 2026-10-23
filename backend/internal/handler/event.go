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
	Budget      uint64         `json:"budget"`

	// Check that the creator ID is given and has UUID format.
	// handler handles request validation
	CreatedBy string `json:"created_by" binding:"required,uuid"`
}

// Read the frontend JSON, create the Event, and return the result.
func (h *EventHandler) CreateEvent(c *gin.Context) {
	var req createEventRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	// New Events can only be DRAFT or IN_PROGRESS.
	if req.Status != db.EventStatusDRAFT &&
		req.Status != db.EventStatusINPROGRESS {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "status must be DRAFT or IN_PROGRESS",
		})
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
		Budget:      req.Budget,
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
		"budget":      req.Budget,
		"created_by":  event.CreatedBy.String(),
		"event_time":  event.EventTime,
		"location":    event.Location,
		"description": event.Description,
		"created_at":  event.CreatedAt.Time,
		"updated_at":  event.UpdatedAt.Time,
	})
}
