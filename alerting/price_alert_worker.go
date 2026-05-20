package alerting

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"nofx/provider/serverchan"
	"nofx/safe"
	"nofx/store"
)

type PriceAlertWorker struct {
	store  *store.Store
	logger *slog.Logger
	stopCh chan struct{}
}

func NewPriceAlertWorker(st *store.Store, logger *slog.Logger) *PriceAlertWorker {
	if logger == nil {
		logger = slog.Default()
	}
	return &PriceAlertWorker{
		store:  st,
		logger: logger,
		stopCh: make(chan struct{}),
	}
}

func (w *PriceAlertWorker) Start() {
	safe.GoNamed("price-alert-worker", func() {
		t := time.NewTicker(1 * time.Minute)
		defer t.Stop()
		for {
			select {
			case <-w.stopCh:
				return
			case <-t.C:
				w.tick()
			}
		}
	})
}

func (w *PriceAlertWorker) Stop() { close(w.stopCh) }

func (w *PriceAlertWorker) tick() {
	if w.store == nil {
		return
	}

	alerts, err := w.store.PriceAlert().ListPending()
	if err != nil {
		w.logger.Warn("price alert: list pending failed", "err", err)
		return
	}
	if len(alerts) == 0 {
		return
	}

	type key struct{ platform, symbol string }
	ranges := map[key]struct {
		high  float64
		low   float64
		close float64
		err   error
	}{}

	getRange := func(platform, symbol string) (float64, float64, float64, error) {
		k := key{platform: platform, symbol: symbol}
		if v, ok := ranges[k]; ok {
			return v.high, v.low, v.close, v.err
		}
		h, l, c, e := FetchLatestKlineRange(platform, symbol)
		ranges[k] = struct {
			high  float64
			low   float64
			close float64
			err   error
		}{h, l, c, e}
		return h, l, c, e
	}

	for _, a := range alerts {
		if a == nil {
			continue
		}
		high, low, closePrice, err := getRange(a.Platform, a.Symbol)
		if err != nil {
			w.logger.Warn("price alert: fetch range failed", "platform", a.Platform, "symbol", a.Symbol, "err", err)
			continue
		}

		triggered := false
		triggerPrice := closePrice
		switch a.Direction {
		case store.PriceAlertDirectionUp:
			if high >= a.TargetPrice {
				triggered = true
				triggerPrice = high
			}
		case store.PriceAlertDirectionDown:
			if low <= a.TargetPrice {
				triggered = true
				triggerPrice = low
			}
		default:
			// fallback based on reference price
			if a.TargetPrice >= a.ReferencePrice && high >= a.TargetPrice {
				triggered = true
				triggerPrice = high
			}
			if a.TargetPrice < a.ReferencePrice && low <= a.TargetPrice {
				triggered = true
				triggerPrice = low
			}
		}

		if !triggered {
			continue
		}

		// Load user ServerChan config
		sc, err := w.store.ServerChanConfig().Get(a.UserID)
		if err != nil || !sc.Enabled {
			continue
		}
		sendKey := strings.TrimSpace(sc.SendKey.String())
		if sendKey == "" {
			continue
		}

		now := time.Now().UTC()
		ok, err := w.store.PriceAlert().MarkTriggered(a.ID, now, triggerPrice)
		if err != nil {
			w.logger.Warn("price alert: mark triggered failed", "id", a.ID, "err", err)
			continue
		}
		if !ok {
			// another worker already triggered it
			continue
		}

		title := fmt.Sprintf("Price alert triggered: %s", a.Symbol)
		desp := fmt.Sprintf(
			"Platform: %s\nSymbol: %s\nTarget: %.8f\nDirection: %s\nTriggeredPrice: %.8f\nTime(UTC): %s\nRangeHigh: %.8f\nRangeLow: %.8f\nClose: %.8f\n",
			a.Platform, a.Symbol, a.TargetPrice, a.Direction, triggerPrice, now.Format(time.RFC3339),
			high, low, closePrice,
		)
		if err := serverchan.New(sendKey).Send(title, desp); err != nil {
			// If sending fails, revert status to pending would be complex; instead log.
			// User can re-create the alert if needed.
			w.logger.Warn("price alert: serverchan send failed", "id", a.ID, "err", err)
		}
	}
}

