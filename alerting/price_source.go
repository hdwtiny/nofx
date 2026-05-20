package alerting

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"nofx/logger"
	"nofx/provider/coinank/coinank_api"
	"nofx/provider/coinank/coinank_enum"
	"nofx/provider/hyperliquid"
)

// FetchLatestKlineRange returns (high, low, close) from the latest available 1m kline range.
// It may merge the last 2 klines to reduce missing intra-minute moves on some data sources.
func FetchLatestKlineRange(platform, symbol string) (high float64, low float64, close float64, err error) {
	platform = strings.ToLower(strings.TrimSpace(platform))
	symbol = strings.TrimSpace(symbol)
	if platform == "" || symbol == "" {
		return 0, 0, 0, fmt.Errorf("platform and symbol are required")
	}

	switch platform {
	case "hyperliquid", "hyperliquid-xyz", "xyz":
		return fetchFromHyperliquid(platform, symbol)
	default:
		return fetchFromCoinAnk(platform, symbol)
	}
}

func fetchFromCoinAnk(platform, symbol string) (high, low, close float64, err error) {
	ex := coinank_enum.Binance
	switch strings.ToLower(platform) {
	case "binance":
		ex = coinank_enum.Binance
	case "bybit":
		ex = coinank_enum.Bybit
	case "okx":
		ex = coinank_enum.Okex
	case "bitget":
		ex = coinank_enum.Bitget
	case "gate":
		ex = coinank_enum.Gate
	case "aster":
		ex = coinank_enum.Aster
	case "kucoin", "lighter":
		// fall back to Binance data like api/handler_klines.go
		ex = coinank_enum.Binance
	default:
		ex = coinank_enum.Binance
	}

	// CoinAnk OKX uses BTC-USDT-SWAP; keep consistent with api/handler_klines.go
	apiSymbol := symbol
	if ex == coinank_enum.Okex && strings.HasSuffix(symbol, "USDT") {
		base := strings.TrimSuffix(symbol, "USDT")
		apiSymbol = fmt.Sprintf("%s-USDT-SWAP", base)
	}

	ctx := context.Background()
	ts := time.Now().UnixMilli()
	klines, err := coinank_api.Kline(ctx, apiSymbol, ex, ts, coinank_enum.To, 2, coinank_enum.Minute1)
	if err != nil {
		// Try fallback to Binance if not Binance already
		if ex != coinank_enum.Binance {
			logger.Warnf("⚠️ CoinAnk Kline failed for %s/%s, fallback to Binance: %v", platform, symbol, err)
			klines, err = coinank_api.Kline(ctx, symbol, coinank_enum.Binance, ts, coinank_enum.To, 2, coinank_enum.Minute1)
		}
	}
	if err != nil {
		return 0, 0, 0, err
	}
	if len(klines) == 0 {
		return 0, 0, 0, fmt.Errorf("no klines returned")
	}

	high = -math.MaxFloat64
	low = math.MaxFloat64
	for _, k := range klines {
		if k.High > high {
			high = k.High
		}
		if k.Low < low {
			low = k.Low
		}
		close = k.Close
	}
	if high <= 0 || low <= 0 || close <= 0 {
		return 0, 0, 0, fmt.Errorf("invalid kline values")
	}
	return high, low, close, nil
}

func fetchFromHyperliquid(platform, symbol string) (high, low, close float64, err error) {
	client := hyperliquid.NewClient()
	ctx := context.Background()

	candles, err := client.GetCandles(ctx, symbol, hyperliquid.MapTimeframe("1m"), 2)
	if err != nil {
		return 0, 0, 0, err
	}
	if len(candles) == 0 {
		return 0, 0, 0, fmt.Errorf("no candles returned")
	}

	high = -math.MaxFloat64
	low = math.MaxFloat64
	for _, c := range candles {
		h, _ := parseFloat(c.High)
		l, _ := parseFloat(c.Low)
		cl, _ := parseFloat(c.Close)
		if h > high {
			high = h
		}
		if l < low {
			low = l
		}
		close = cl
	}
	if high <= 0 || low <= 0 || close <= 0 {
		return 0, 0, 0, fmt.Errorf("invalid candle values")
	}
	return high, low, close, nil
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(s), 64)
}

