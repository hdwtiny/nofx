package api

import (
	"net/http"
	"strings"

	"nofx/alerting"
	"nofx/market"
	"nofx/store"

	"github.com/gin-gonic/gin"
)

type createPriceAlertRequest struct {
	Symbol      string  `json:"symbol"`
	Platform    string  `json:"platform"`
	TargetPrice float64 `json:"target_price"`
}

func (s *Server) handleListPriceAlerts(c *gin.Context) {
	userID := c.GetString("user_id")
	alerts, err := s.store.PriceAlert().List(userID)
	if err != nil {
		SafeInternalError(c, "Failed to list price alerts", err)
		return
	}
	c.JSON(http.StatusOK, alerts)
}

func (s *Server) handleCreatePriceAlert(c *gin.Context) {
	userID := c.GetString("user_id")

	// Require ServerChan configured for this user
	sc, err := s.store.ServerChanConfig().Get(userID)
	if err != nil || !sc.Enabled || strings.TrimSpace(sc.SendKey.String()) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ServerChan is not configured. Please set SendKey first."})
		return
	}

	var req createPriceAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SafeBadRequest(c, "Invalid request body")
		return
	}
	req.Symbol = market.Normalize(strings.TrimSpace(req.Symbol))
	req.Platform = strings.ToLower(strings.TrimSpace(req.Platform))

	if req.Symbol == "" {
		SafeBadRequest(c, "symbol is required")
		return
	}
	if req.Platform == "" {
		SafeBadRequest(c, "platform is required")
		return
	}
	if req.TargetPrice <= 0 {
		SafeBadRequest(c, "target_price must be > 0")
		return
	}

	existing, err := s.store.PriceAlert().List(userID)
	if err != nil {
		SafeInternalError(c, "Failed to check existing alerts", err)
		return
	}
	for _, a := range existing {
		if a.Status == store.PriceAlertStatusPending &&
			strings.EqualFold(a.Symbol, req.Symbol) &&
			strings.EqualFold(a.Platform, req.Platform) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "A pending alert already exists for this symbol on this platform"})
			return
		}
	}

	// Determine direction based on reference price at creation time.
	_, _, closePrice, err := alerting.FetchLatestKlineRange(req.Platform, req.Symbol)
	if err != nil || closePrice <= 0 {
		SafeInternalError(c, "Failed to fetch reference price", err)
		return
	}
	direction := store.PriceAlertDirectionUp
	if req.TargetPrice < closePrice {
		direction = store.PriceAlertDirectionDown
	}

	alert, err := s.store.PriceAlert().Create(userID, req.Symbol, req.Platform, req.TargetPrice, closePrice, direction)
	if err != nil {
		SafeInternalError(c, "Failed to create price alert", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": alert.ID})
}

func (s *Server) handleDeletePriceAlert(c *gin.Context) {
	userID := c.GetString("user_id")
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		SafeBadRequest(c, "id is required")
		return
	}
	if err := s.store.PriceAlert().Delete(userID, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Price alert not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted"})
}

