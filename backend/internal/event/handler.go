// receive HTTP requests and pass data to the service

package event

import (
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	service EventService
}

func NewEventHandler(service EventService) *EventHandler {
	return &EventHandler{service: service}
}

// JSON from frontend
type createEventRequest struct {
	Name        string    `json:"event_name"`
	EventTime   time.Time `json:"event_time"`
	Budget      float64   `json:"budget"`
	Location    string    `json:"location"`
	Description string    `json:"description"`
	SocietyID   string    `json:"society_id"`
	CreatedBy   int64     `json:"created_by"`
}

// read json, use EventService, return http code
// parses the JSON request body into req
func(h *EventHandler) CreateEvent(c *gin.Context) {
	// declares a variable req createEventRequest
	var req createEventRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	// := declare and assign
	event, err := h.service.CreateEvent(
		c.Request.Context(), // read context from http request
		req.Name,
		req.EventTime,
		req.Budget,
		req.Location,
		req.Description,
		req.SocietyID,
		req.CreatedBy,
	)

	// check err
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) // convert error type into a string
		return
	}
	// return 201, and created data.
	c.JSON(http.StatusCreated, event)
}