package store

import (
	"errors"
	"fmt"
	"nofx/crypto"
	"time"

	"gorm.io/gorm"
)

// ServerChanConfig stores per-user ServerChan send key (SendKey/SCKEY).
type ServerChanConfig struct {
	UserID    string                 `gorm:"primaryKey;column:user_id" json:"user_id"`
	SendKey   crypto.EncryptedString `gorm:"column:send_key;default:''" json:"-"`
	Enabled   bool                   `gorm:"column:enabled;default:true" json:"enabled"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

func (ServerChanConfig) TableName() string { return "serverchan_configs" }

type ServerChanConfigStore struct {
	db *gorm.DB
}

func NewServerChanConfigStore(db *gorm.DB) *ServerChanConfigStore {
	return &ServerChanConfigStore{db: db}
}

func (s *ServerChanConfigStore) initTables() error {
	// For PostgreSQL with existing table, skip AutoMigrate
	if s.db.Dialector.Name() == "postgres" {
		var tableExists int64
		s.db.Raw(`SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'serverchan_configs'`).Scan(&tableExists)
		if tableExists > 0 {
			return nil
		}
	}
	return s.db.AutoMigrate(&ServerChanConfig{})
}

func (s *ServerChanConfigStore) Get(userID string) (*ServerChanConfig, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID required")
	}
	var cfg ServerChanConfig
	if err := s.db.Where("user_id = ?", userID).First(&cfg).Error; err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s *ServerChanConfigStore) Upsert(userID, sendKey string, enabled bool) error {
	if userID == "" {
		return fmt.Errorf("userID required")
	}
	if sendKey == "" {
		return fmt.Errorf("sendKey required")
	}
	var cfg ServerChanConfig
	err := s.db.Where("user_id = ?", userID).First(&cfg).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	cfg.UserID = userID
	cfg.SendKey = crypto.EncryptedString(sendKey)
	cfg.Enabled = enabled
	cfg.UpdatedAt = time.Now().UTC()
	return s.db.Save(&cfg).Error
}

func (s *ServerChanConfigStore) Disable(userID string) error {
	if userID == "" {
		return fmt.Errorf("userID required")
	}
	res := s.db.Model(&ServerChanConfig{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"enabled":    false,
			"updated_at": time.Now().UTC(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		// no-op if missing
		return nil
	}
	return nil
}

