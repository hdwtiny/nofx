package store

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PriceAlertStatus string

const (
	PriceAlertStatusPending   PriceAlertStatus = "pending"
	PriceAlertStatusTriggered PriceAlertStatus = "triggered"
	PriceAlertStatusCancelled PriceAlertStatus = "cancelled"
)

type PriceAlertDirection string

const (
	PriceAlertDirectionUp   PriceAlertDirection = "up"
	PriceAlertDirectionDown PriceAlertDirection = "down"
)

// PriceAlert represents a one-time price alert for a symbol on a platform.
type PriceAlert struct {
	ID            string           `gorm:"primaryKey" json:"id"`
	UserID        string           `gorm:"column:user_id;not null;index" json:"user_id"`
	Symbol        string           `gorm:"column:symbol;not null;index" json:"symbol"`
	Platform      string           `gorm:"column:platform;not null;index" json:"platform"`
	TargetPrice   float64          `gorm:"column:target_price;not null" json:"target_price"`
	ReferencePrice float64         `gorm:"column:reference_price;not null;default:0" json:"reference_price"`
	Direction     PriceAlertDirection `gorm:"column:direction;not null;default:up" json:"direction"`
	Status        PriceAlertStatus `gorm:"column:status;not null;default:pending;index" json:"status"`
	TriggeredAt   *time.Time       `gorm:"column:triggered_at" json:"triggered_at,omitempty"`
	TriggeredPrice *float64        `gorm:"column:triggered_price" json:"triggered_price,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

func (PriceAlert) TableName() string { return "price_alerts" }

type PriceAlertStore struct {
	db *gorm.DB
}

func NewPriceAlertStore(db *gorm.DB) *PriceAlertStore {
	return &PriceAlertStore{db: db}
}

func (s *PriceAlertStore) initTables() error {
	// For PostgreSQL with existing table, skip AutoMigrate
	if s.db.Dialector.Name() == "postgres" {
		var tableExists int64
		s.db.Raw(`SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'price_alerts'`).Scan(&tableExists)
		if tableExists > 0 {
			return nil
		}
	}
	return s.db.AutoMigrate(&PriceAlert{})
}

func (s *PriceAlertStore) Create(userID, symbol, platform string, targetPrice float64, referencePrice float64, direction PriceAlertDirection) (*PriceAlert, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID required")
	}
	if symbol == "" {
		return nil, fmt.Errorf("symbol required")
	}
	if platform == "" {
		return nil, fmt.Errorf("platform required")
	}
	if targetPrice <= 0 {
		return nil, fmt.Errorf("targetPrice must be > 0")
	}
	if referencePrice <= 0 {
		return nil, fmt.Errorf("referencePrice must be > 0")
	}
	if direction != PriceAlertDirectionUp && direction != PriceAlertDirectionDown {
		return nil, fmt.Errorf("invalid direction")
	}

	a := &PriceAlert{
		ID:          uuid.New().String(),
		UserID:      userID,
		Symbol:      symbol,
		Platform:    platform,
		TargetPrice: targetPrice,
		ReferencePrice: referencePrice,
		Direction:   direction,
		Status:      PriceAlertStatusPending,
	}
	if err := s.db.Create(a).Error; err != nil {
		return nil, err
	}
	return a, nil
}

func (s *PriceAlertStore) List(userID string) ([]*PriceAlert, error) {
	var alerts []*PriceAlert
	if err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&alerts).Error; err != nil {
		return nil, err
	}
	return alerts, nil
}

func (s *PriceAlertStore) GetByID(userID, id string) (*PriceAlert, error) {
	var a PriceAlert
	if err := s.db.Where("user_id = ? AND id = ?", userID, id).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *PriceAlertStore) Cancel(userID, id string) error {
	res := s.db.Model(&PriceAlert{}).
		Where("user_id = ? AND id = ? AND status = ?", userID, id, PriceAlertStatusPending).
		Updates(map[string]interface{}{
			"status":     PriceAlertStatusCancelled,
			"updated_at": time.Now().UTC(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		// Treat as not found/pending
		return fmt.Errorf("price alert not found or not pending")
	}
	return nil
}

func (s *PriceAlertStore) Delete(userID, id string) error {
	res := s.db.Where("user_id = ? AND id = ?", userID, id).Delete(&PriceAlert{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("price alert not found")
	}
	return nil
}

// ListPending returns all pending alerts across all users.
func (s *PriceAlertStore) ListPending() ([]*PriceAlert, error) {
	var alerts []*PriceAlert
	if err := s.db.Where("status = ?", PriceAlertStatusPending).
		Order("created_at ASC").
		Find(&alerts).Error; err != nil {
		return nil, err
	}
	return alerts, nil
}

// MarkTriggered marks an alert as triggered, only if it is currently pending.
func (s *PriceAlertStore) MarkTriggered(id string, triggeredAt time.Time, triggeredPrice float64) (bool, error) {
	res := s.db.Model(&PriceAlert{}).
		Where("id = ? AND status = ?", id, PriceAlertStatusPending).
		Updates(map[string]interface{}{
			"status":          PriceAlertStatusTriggered,
			"triggered_at":    triggeredAt,
			"triggered_price": triggeredPrice,
			"updated_at":      time.Now().UTC(),
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

