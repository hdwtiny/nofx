package store

import (
	"testing"
	"time"
)

func TestPriceAlertStore_CreateAndTrigger(t *testing.T) {
	st, err := New(":memory:")
	if err != nil {
		t.Fatalf("failed to init store: %v", err)
	}
	defer st.Close()

	userID := "u1"
	// Create a minimal user so foreign constraints (if any) won't fail.
	_ = st.User().Create(&User{ID: userID, Email: "u1@example.com", PasswordHash: "x"})

	a, err := st.PriceAlert().Create(userID, "BTCUSDT", "binance", 100, 90, PriceAlertDirectionUp)
	if err != nil {
		t.Fatalf("create alert: %v", err)
	}

	pending, err := st.PriceAlert().ListPending()
	if err != nil {
		t.Fatalf("list pending: %v", err)
	}
	if len(pending) == 0 {
		t.Fatalf("expected pending alerts")
	}

	ok, err := st.PriceAlert().MarkTriggered(a.ID, time.Now().UTC(), 101)
	if err != nil {
		t.Fatalf("mark triggered: %v", err)
	}
	if !ok {
		t.Fatalf("expected ok=true")
	}

	ok2, err := st.PriceAlert().MarkTriggered(a.ID, time.Now().UTC(), 102)
	if err != nil {
		t.Fatalf("mark triggered again: %v", err)
	}
	if ok2 {
		t.Fatalf("expected ok=false for second trigger")
	}
}

