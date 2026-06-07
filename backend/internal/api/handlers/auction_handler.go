package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/drayauction/auctionxi/internal/api/dto"
	"github.com/drayauction/auctionxi/internal/domain"
	"github.com/drayauction/auctionxi/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuctionHandler struct {
	auctionService *services.AuctionService
}

func NewAuctionHandler(auctionService *services.AuctionService) *AuctionHandler {
	return &AuctionHandler{auctionService: auctionService}
}

func (h *AuctionHandler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	var req dto.CreateAuctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{Code: "VALIDATION_ERROR", Message: err.Error()},
		})
		return
	}

	result, err := h.auctionService.Create(c.Request.Context(), userID, services.CreateAuctionInput{
		AIOpponents: req.AIOpponents,
		Difficulty:  domain.Difficulty(req.Difficulty),
	})
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *AuctionHandler) Get(c *gin.Context) {
	auctionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{Code: "VALIDATION_ERROR", Message: "Invalid auction ID"},
		})
		return
	}

	result, err := h.auctionService.GetState(c.Request.Context(), auctionID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AuctionHandler) Start(c *gin.Context) {
	auctionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{Code: "VALIDATION_ERROR", Message: "Invalid auction ID"},
		})
		return
	}

	result, err := h.auctionService.Start(c.Request.Context(), auctionID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AuctionHandler) Bid(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	auctionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{Code: "VALIDATION_ERROR", Message: "Invalid auction ID"},
		})
		return
	}

	var amount int64
	if inc := c.Query("increment"); inc != "" {
		incM, _ := strconv.ParseInt(inc, 10, 64)
		amount, err = h.auctionService.BidWithIncrement(c.Request.Context(), auctionID, userID, incM)
	} else {
		var req dto.BidRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error: dto.ErrorDetail{Code: "VALIDATION_ERROR", Message: err.Error()},
			})
			return
		}
		amount, err = h.auctionService.PlaceBid(c.Request.Context(), auctionID, userID, req.Amount)
	}
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"accepted": true, "amount": amount})
}

func (h *AuctionHandler) Pass(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	auctionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{Code: "VALIDATION_ERROR", Message: "Invalid auction ID"},
		})
		return
	}

	if err := h.auctionService.Pass(c.Request.Context(), auctionID, userID); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"passed": true})
}

func (h *AuctionHandler) Events(c *gin.Context) {
	auctionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{Code: "VALIDATION_ERROR", Message: "Invalid auction ID"},
		})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	eventCh := h.auctionService.SubscribeEvents(auctionID)
	c.Stream(func(w io.Writer) bool {
		select {
		case event, ok := <-eventCh:
			if !ok {
				return false
			}
			data, _ := json.Marshal(event)
			c.SSEvent(event.Type, string(data))
			return true
		case <-c.Request.Context().Done():
			return false
		case <-time.After(30 * time.Second):
			c.SSEvent("heartbeat", fmt.Sprintf(`{"ts":%d}`, time.Now().Unix()))
			return true
		}
	})
}

func (h *AuctionHandler) Results(c *gin.Context) {
	auctionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{Code: "VALIDATION_ERROR", Message: "Invalid auction ID"},
		})
		return
	}

	result, err := h.auctionService.GetResults(c.Request.Context(), auctionID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
