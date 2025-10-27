package coinbase

import (
	"errors"
	"maps"
	"slices"
	"testcharting/pkg/charting/sources"
	"time"
)

func NewCoinbaseSource() sources.Source {
	return &CoinbaseSource{
		Name:    "Coinbase",
		BaseURL: "https://api.exchange.coinbase.com",
	}
}

type CoinbaseSource struct {
	Name         string
	BaseURL      string
	TradingPairs []CoinbaseTradingPair
}

func (c *CoinbaseSource) GetName() string {
	return c.Name
}

func (c *CoinbaseSource) GetTradingPairs() []sources.TradingPair {
	c.fetchTradingPairsIfNeeded()

	pairs := make([]sources.TradingPair, len(c.TradingPairs))
	for i, pair := range c.TradingPairs {
		pairs[i] = sources.TradingPair{
			Id:     pair.Id,
			Name:   pair.Name,
			Base:   pair.Base,
			Quote:  pair.Quote,
			Status: pair.Status,
		}
	}
	return pairs
}

func (c *CoinbaseSource) GetValidIntervals() []string {
	intervalMap := c.getValidIntervalMap()
	return slices.Collect(maps.Keys(intervalMap))
}

func (c *CoinbaseSource) GetData(pairId string, interval string, start time.Time, end time.Time) ([]sources.ChartEntry, error) {
	intervalMap := c.getValidIntervalMap()
	granularity, ok := intervalMap[interval]
	if !ok {
		return nil, errors.New("invalid interval")
	}

	candles, err := c.fetchDataIfNeeded(pairId, granularity, start, end)
	if err != nil {
		return nil, err
	}

	entries := make([]sources.ChartEntry, len(candles))
	for i, candle := range candles {
		entries[i] = sources.ChartEntry{
			Timestamp: time.Unix(candle.Time, 0),
			Open:      candle.Open,
			High:      candle.High,
			Low:       candle.Low,
			Close:     candle.Close,
			Volume:    candle.Volume,
		}
	}
	return entries, nil
}

func (c *CoinbaseSource) getValidIntervalMap() map[string]int {
	// Supported intervals by Coinbase API in seconds: {60, 300, 900, 3600, 21600, 86400}
	return map[string]int{
		"1m":  60,
		"5m":  300,
		"15m": 900,
		"1h":  3600,
		"6h":  21600,
		"1d":  86400,
	}
}

func (c *CoinbaseSource) fetchTradingPairsIfNeeded() {
	if len(c.TradingPairs) > 0 {
		return
	}

	url := c.BaseURL + "/products"

	pairs, _, err := sources.GetRequestParse[[]CoinbaseTradingPair](url, nil)
	if err != nil {
		return
	}

	c.TradingPairs = *pairs
}

func (c *CoinbaseSource) fetchDataIfNeeded(pairId string, granularity int, start time.Time, end time.Time) ([]CoinbaseCandle, error) {
	url := c.BaseURL + "/products/" + pairId + "/candles"

	// TODO - Coinbase will only return max 300 candles per request. You may need to make multiple requests if the time range is large.
	// TODO - Implement rate limiting handling as per Coinbase API guidelines.
	// https://docs.cdp.coinbase.com/exchange/rest-api/rate-limits

	candleCount := int(end.Sub(start).Seconds()) / granularity
	requestCount := (candleCount + 299) / 300
	// TODO - implement multiple requests if needed
	if requestCount > 1 {
		return nil, errors.New("time range too large, multiple requests needed - not implemented yet")
	}

	params := map[string]string{
		"start":       start.UTC().Format(time.RFC3339),
		"end":         end.UTC().Format(time.RFC3339),
		"granularity": string(rune(granularity)),
	}

	candles, statusCode, err := sources.GetRequestParse[[]CoinbaseCandle](url, params)
	if err != nil {
		return nil, err
	}

	_ = statusCode // TODO - handle non-200 status codes

	return *candles, nil
}
