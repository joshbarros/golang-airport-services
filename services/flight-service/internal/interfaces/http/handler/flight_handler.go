package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joshbarros/golang-airport-services/pkg/errors"
	"github.com/joshbarros/golang-airport-services/pkg/logger"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/interfaces/http/dto"
	"github.com/joshbarros/golang-airport-services/services/flight-service/internal/usecase"
	"go.uber.org/zap"
)

// FlightHandler handles HTTP requests for flight operations
type FlightHandler struct {
	createFlightUC       *usecase.CreateFlightUseCase
	getFlightUC          *usecase.GetFlightUseCase
	listFlightsUC        *usecase.ListFlightsUseCase
	updateFlightStatusUC *usecase.UpdateFlightStatusUseCase
	delayFlightUC        *usecase.DelayFlightUseCase
	cancelFlightUC       *usecase.CancelFlightUseCase
	logger               *logger.Logger
}

// NewFlightHandler creates a new FlightHandler
func NewFlightHandler(
	createFlightUC *usecase.CreateFlightUseCase,
	getFlightUC *usecase.GetFlightUseCase,
	listFlightsUC *usecase.ListFlightsUseCase,
	updateFlightStatusUC *usecase.UpdateFlightStatusUseCase,
	delayFlightUC *usecase.DelayFlightUseCase,
	cancelFlightUC *usecase.CancelFlightUseCase,
	logger *logger.Logger,
) *FlightHandler {
	return &FlightHandler{
		createFlightUC:       createFlightUC,
		getFlightUC:          getFlightUC,
		listFlightsUC:        listFlightsUC,
		updateFlightStatusUC: updateFlightStatusUC,
		delayFlightUC:        delayFlightUC,
		cancelFlightUC:       cancelFlightUC,
		logger:               logger,
	}
}

// CreateFlight handles POST /api/v1/flights
// @Summary Create a new flight
// @Tags flights
// @Accept json
// @Produce json
// @Param request body dto.CreateFlightRequest true "Create flight request"
// @Success 201 {object} dto.FlightResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/flights [post]
func (h *FlightHandler) CreateFlight(c *gin.Context) {
	var req dto.CreateFlightRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, errors.BadRequest("invalid request body: "+err.Error()))
		return
	}

	requestID := c.GetString("request_id")
	h.logger.Info("Creating flight",
		zap.String("request_id", requestID),
		zap.String("flight_number", req.FlightNumber),
	)

	input := usecase.CreateFlightInput{
		FlightNumber:  req.FlightNumber,
		Origin:        req.Origin,
		Destination:   req.Destination,
		DepartureTime: req.DepartureTime,
		ArrivalTime:   req.ArrivalTime,
		AircraftType:  req.AircraftType,
		Gate:          req.Gate,
		Terminal:      req.Terminal,
	}

	output, err := h.createFlightUC.Execute(c.Request.Context(), input)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	response := dto.FlightResponse{
		ID:            output.ID,
		FlightNumber:  output.FlightNumber,
		Origin:        output.Origin,
		Destination:   output.Destination,
		DepartureTime: output.DepartureTime,
		ArrivalTime:   output.ArrivalTime,
		Status:        output.Status,
		AircraftType:  output.AircraftType,
		Gate:          output.Gate,
		Terminal:      output.Terminal,
		CreatedAt:     output.CreatedAt,
		UpdatedAt:     output.CreatedAt,
	}
	c.JSON(http.StatusCreated, response)
}

// GetFlight handles GET /api/v1/flights/:id
// @Summary Get a flight by ID
// @Tags flights
// @Produce json
// @Param id path string true "Flight ID"
// @Success 200 {object} dto.FlightResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/flights/{id} [get]
func (h *FlightHandler) GetFlight(c *gin.Context) {
	flightID := c.Param("id")

	requestID := c.GetString("request_id")
	h.logger.Debug("Getting flight",
		zap.String("request_id", requestID),
		zap.String("flight_id", flightID),
	)

	output, err := h.getFlightUC.Execute(c.Request.Context(), flightID)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	response := toFlightResponse(output)
	c.JSON(http.StatusOK, response)
}

// ListFlights handles GET /api/v1/flights
// @Summary List flights with optional filters
// @Tags flights
// @Produce json
// @Param status query string false "Filter by status (scheduled, boarding, departed, in_flight, landed, arrived, delayed, cancelled)"
// @Param origin query string false "Filter by origin airport code (e.g., JFK)"
// @Param destination query string false "Filter by destination airport code (e.g., LAX)"
// @Param start_date query string false "Filter by start date (RFC3339 format)"
// @Param end_date query string false "Filter by end date (RFC3339 format)"
// @Param active_only query boolean false "Show only active flights"
// @Param delayed_only query boolean false "Show only delayed flights"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Page size (default: 20, max: 100)"
// @Success 200 {object} dto.FlightListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/flights [get]
func (h *FlightHandler) ListFlights(c *gin.Context) {
	requestID := c.GetString("request_id")

	// Parse query parameters
	status := c.Query("status")
	origin := c.Query("origin")
	destination := c.Query("destination")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	activeOnly := c.Query("active_only") == "true"
	delayedOnly := c.Query("delayed_only") == "true"

	// Parse pagination
	limit, offset := ParsePagination(c)

	// Parse dates if provided
	var startDate, endDate *time.Time
	if startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &parsed
		} else {
			h.respondWithError(c, errors.BadRequest("invalid start_date format, use RFC3339"))
			return
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &parsed
		} else {
			h.respondWithError(c, errors.BadRequest("invalid end_date format, use RFC3339"))
			return
		}
	}

	h.logger.Debug("Listing flights",
		zap.String("request_id", requestID),
		zap.String("status", status),
		zap.String("origin", origin),
		zap.String("destination", destination),
		zap.Bool("active_only", activeOnly),
		zap.Bool("delayed_only", delayedOnly),
		zap.Int("limit", limit),
		zap.Int("offset", offset),
	)

	input := usecase.ListFlightsInput{
		Status:      status,
		Origin:      origin,
		Destination: destination,
		StartDate:   startDate,
		EndDate:     endDate,
		ActiveOnly:  activeOnly,
		DelayedOnly: delayedOnly,
		Limit:       limit,
		Offset:      offset,
	}

	output, err := h.listFlightsUC.Execute(c.Request.Context(), input)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	// Convert to response
	flights := make([]dto.FlightResponse, 0, len(output.Flights))
	for _, flight := range output.Flights {
		flights = append(flights, toFlightResponse(flight))
	}

	response := dto.FlightListResponse{
		Flights:    flights,
		TotalCount: output.TotalCount,
		Page:       output.Page,
		PageSize:   output.PageSize,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateFlightStatus handles PATCH /api/v1/flights/:id/status
// @Summary Update flight status
// @Tags flights
// @Accept json
// @Produce json
// @Param id path string true "Flight ID"
// @Param request body dto.UpdateFlightStatusRequest true "Update status request"
// @Success 200 {object} dto.FlightResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/flights/{id}/status [patch]
func (h *FlightHandler) UpdateFlightStatus(c *gin.Context) {
	flightID := c.Param("id")
	var req dto.UpdateFlightStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, errors.BadRequest("invalid request body: "+err.Error()))
		return
	}

	requestID := c.GetString("request_id")
	h.logger.Info("Updating flight status",
		zap.String("request_id", requestID),
		zap.String("flight_id", flightID),
		zap.String("new_status", req.Status),
	)

	input := usecase.UpdateFlightStatusInput{
		FlightID: flightID,
		Status:   req.Status,
	}

	output, err := h.updateFlightStatusUC.Execute(c.Request.Context(), input)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	response := toFlightResponse(output)
	c.JSON(http.StatusOK, response)
}

// DelayFlight handles POST /api/v1/flights/:id/delay
// @Summary Delay a flight
// @Tags flights
// @Accept json
// @Produce json
// @Param id path string true "Flight ID"
// @Param request body dto.DelayFlightRequest true "Delay flight request"
// @Success 200 {object} dto.FlightResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/flights/{id}/delay [post]
func (h *FlightHandler) DelayFlight(c *gin.Context) {
	flightID := c.Param("id")
	var req dto.DelayFlightRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, errors.BadRequest("invalid request body: "+err.Error()))
		return
	}

	requestID := c.GetString("request_id")
	h.logger.Info("Delaying flight",
		zap.String("request_id", requestID),
		zap.String("flight_id", flightID),
		zap.Int("delay_minutes", req.DelayMinutes),
		zap.String("reason", req.Reason),
	)

	input := usecase.DelayFlightInput{
		FlightID:      flightID,
		DelayDuration: time.Duration(req.DelayMinutes) * time.Minute,
		Reason:        req.Reason,
	}

	output, err := h.delayFlightUC.Execute(c.Request.Context(), input)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	response := toFlightResponse(output)
	c.JSON(http.StatusOK, response)
}

// CancelFlight handles POST /api/v1/flights/:id/cancel
// @Summary Cancel a flight
// @Tags flights
// @Accept json
// @Produce json
// @Param id path string true "Flight ID"
// @Param request body dto.CancelFlightRequest true "Cancel flight request"
// @Success 200 {object} dto.FlightResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/flights/{id}/cancel [post]
func (h *FlightHandler) CancelFlight(c *gin.Context) {
	flightID := c.Param("id")
	var req dto.CancelFlightRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondWithError(c, errors.BadRequest("invalid request body: "+err.Error()))
		return
	}

	requestID := c.GetString("request_id")
	h.logger.Info("Cancelling flight",
		zap.String("request_id", requestID),
		zap.String("flight_id", flightID),
		zap.String("reason", req.Reason),
	)

	input := usecase.CancelFlightInput{
		FlightID: flightID,
		Reason:   req.Reason,
	}

	output, err := h.cancelFlightUC.Execute(c.Request.Context(), input)
	if err != nil {
		h.respondWithError(c, err)
		return
	}

	response := toFlightResponse(output)
	c.JSON(http.StatusOK, response)
}

// Helper methods

func (h *FlightHandler) respondWithError(c *gin.Context, err error) {
	// Check if it's an AppError
	if appErr, ok := err.(*errors.AppError); ok {
		response := dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:     appErr.Code,
				Message:  appErr.Message,
				Metadata: appErr.Metadata,
			},
		}

		h.logger.Warn("Request failed",
			zap.String("request_id", c.GetString("request_id")),
			zap.String("error_code", appErr.Code),
			zap.String("error_message", appErr.Message),
			zap.Int("status_code", appErr.StatusCode),
		)

		c.JSON(appErr.StatusCode, response)
		return
	}

	// Unknown error - log and return 500
	h.logger.Error("Unexpected error",
		zap.String("request_id", c.GetString("request_id")),
		zap.Error(err),
	)

	response := dto.ErrorResponse{
		Error: dto.ErrorDetail{
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "An unexpected error occurred",
		},
	}
	c.JSON(http.StatusInternalServerError, response)
}

func toFlightResponse(output *usecase.GetFlightOutput) dto.FlightResponse {
	return dto.FlightResponse{
		ID:                  output.ID,
		FlightNumber:        output.FlightNumber,
		Origin:              output.Origin,
		Destination:         output.Destination,
		DepartureTime:       output.DepartureTime,
		ArrivalTime:         output.ArrivalTime,
		ActualDepartureTime: output.ActualDepartureTime,
		ActualArrivalTime:   output.ActualArrivalTime,
		Status:              output.Status,
		AircraftType:        output.AircraftType,
		Gate:                output.Gate,
		Terminal:            output.Terminal,
		DelayReason:         output.DelayReason,
		CancellationReason:  output.CancellationReason,
		CreatedAt:           output.CreatedAt,
		UpdatedAt:           output.UpdatedAt,
	}
}

// ParsePagination extracts pagination parameters from query string
func ParsePagination(c *gin.Context) (limit, offset int) {
	limit = 20 // default
	offset = 0

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			offset = (page - 1) * limit
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	return limit, offset
}
