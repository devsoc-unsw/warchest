// receive HTTP requests and pass data to the service

package handler

import (
	"backend/db"
	"backend/internal/service"
	"encoding/json"
	"io"
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
	// Temporary legacy input: replace with authenticated identity before enabling member access.
	CreatedBy string `json:"created_by"`
}

// read json, use EventService, return http code
// parses the JSON request body into req
func (h *EventHandler) CreateEvent(c *gin.Context) {
	// declares a variable req createEventRequest
	var req createEventRequest

	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request body must contain one JSON object"})
		return
	}
	// := declare and assign
	event, err := h.service.CreateEvent(c.Request.Context(), service.CreateEventInput{
		Name:        req.Name,
		EventTime:   req.EventTime,
		Location:    req.Location,
		Description: req.Description,
		SocietyID:   req.SocietyID,
		CreatedBy:   req.CreatedBy,
		Status:      req.Status,
	})

	// check err
	if err != nil {
		// convert error type into a string
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// return 201, and created data.
	c.JSON(http.StatusCreated, event)
}
