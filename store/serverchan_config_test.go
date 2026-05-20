package store

import "testing"

func TestServerChanConfigStore_UpsertAndGet(t *testing.T) {
	st, err := New(":memory:")
	if err != nil {
		t.Fatalf("failed to init store: %v", err)
	}
	defer st.Close()

	userID := "u1"
	_ = st.User().Create(&User{ID: userID, Email: "u1@example.com", PasswordHash: "x"})

	if err := st.ServerChanConfig().Upsert(userID, "SCT_TEST", true); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	cfg, err := st.ServerChanConfig().Get(userID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !cfg.Enabled {
		t.Fatalf("expected enabled")
	}
	if cfg.SendKey.String() == "" {
		t.Fatalf("expected send key")
	}
}

