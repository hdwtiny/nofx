package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"nofx/config"
	"nofx/crypto"
	"nofx/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type upsertServerChanRequest struct {
	SendKey string `json:"send_key"`
	Enabled *bool  `json:"enabled"`
}

func (s *Server) handleGetServerChanConfig(c *gin.Context) {
	userID := c.GetString("user_id")
	cfg, err := s.store.ServerChanConfig().Get(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"enabled":    false,
				"configured": false,
			})
			return
		}
		SafeInternalError(c, "Failed to load ServerChan config", err)
		return
	}

	sendKey := strings.TrimSpace(cfg.SendKey.String())
	c.JSON(http.StatusOK, gin.H{
		"enabled":    cfg.Enabled,
		"configured": sendKey != "",
	})
}

func (s *Server) handleUpsertServerChanConfig(c *gin.Context) {
	userID := c.GetString("user_id")
	appCfg := config.Get()

	bodyBytes, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	var req upsertServerChanRequest

	if !appCfg.TransportEncryption {
		if err := json.Unmarshal(bodyBytes, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}
	} else {
		var encryptedPayload crypto.EncryptedPayload
		if err := json.Unmarshal(bodyBytes, &encryptedPayload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format, encrypted transmission required"})
			return
		}
		if encryptedPayload.WrappedKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "This endpoint only supports encrypted transmission, please use encrypted client",
				"code":    "ENCRYPTION_REQUIRED",
				"message": "Encrypted transmission is required for security reasons",
			})
			return
		}

		decrypted, err := s.cryptoHandler.cryptoService.DecryptSensitiveData(&encryptedPayload)
		if err != nil {
			logger.Infof("❌ Failed to decrypt serverchan config (UserID: %s): %v", userID, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decrypt data"})
			return
		}
		if err := json.Unmarshal([]byte(decrypted), &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse decrypted data"})
			return
		}
	}

	sendKey := strings.TrimSpace(req.SendKey)
	if sendKey == "" {
		SafeBadRequest(c, "send_key is required")
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	if err := s.store.ServerChanConfig().Upsert(userID, sendKey, enabled); err != nil {
		SafeInternalError(c, "Failed to save ServerChan config", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ServerChan config saved"})
}

