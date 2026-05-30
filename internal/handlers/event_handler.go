package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"projectvows/internal/dto"
	"projectvows/internal/repositories"
	"projectvows/internal/services"
	"projectvows/internal/utils"
)

type EventHandler struct {
	svc *services.EventService
}

func NewEventHandler(svc *services.EventService) *EventHandler {
	return &EventHandler{svc: svc}
}

// POST /api/events
func (h *EventHandler) Create(c *gin.Context) {
	var req dto.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}
	event, err := h.svc.Create(req)
	if err != nil {
		utils.Error(c, http.StatusConflict, "create_failed", err.Error())
		return
	}
	utils.Success(c, http.StatusCreated, "Event created", event)
}

// GET /api/events
func (h *EventHandler) List(c *gin.Context) {
	events, err := h.svc.List()
	if err != nil {
		utils.InternalError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "", events)
}

// GET /api/events/:id
func (h *EventHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid event id")
		return
	}
	event, err := h.svc.GetByID(id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			utils.Error(c, http.StatusNotFound, "event_not_found", "event not found")
			return
		}
		utils.InternalError(c, err)
		return
	}
	utils.Success(c, http.StatusOK, "", event)
}
