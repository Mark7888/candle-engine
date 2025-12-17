package binance

import (
	"errors"
	"maps"
	"slices"
	"strconv"
	"testcharting/pkg/charting/sources"
	"time"
)

// https://developers.binance.com/docs/derivatives/usds-margined-futures/market-data/rest-api/Exchange-Information
// https://developers.binance.com/docs/derivatives/usds-margined-futures/market-data/rest-api/Kline-Candlestick-Data
// https://fapi.binance.com/fapi/v1/klines?symbol=BTCUSDC&timeFrame=1d

func NewBinanceSource() sources.Source {
	return &BinanceSource{
		Name:    "Binance",
		BaseURL: "https://fapi.binance.com/fapi/v1",
	}
}

type BinanceSource struct {
	Name         string
	BaseURL      string
	TradingPairs []BinanceTradingPair
}

func (c *BinanceSource) GetName() string {
	return c.Name
}

func (c *BinanceSource) GetTradingPairs() []sources.TradingPair {
	c.fetchTradingPairsIfNeeded()

	pairs := make([]sources.TradingPair, len(c.TradingPairs))
	for i, pair := range c.TradingPairs {
		pairs[i] = sources.TradingPair{
			Id:     pair.Pair, // ? Use Pair or Symbol?
			Name:   pair.Symbol,
			Base:   pair.BaseAsset,
			Quote:  pair.QuoteAsset,
			Status: pair.Status,
		}
	}
	return pairs
}

func (c *BinanceSource) GetValidTimeFrames() []string {
	timeFrameMap := c.getValidTimeFrameMap()
	return slices.Collect(maps.Keys(timeFrameMap))
}

func (c *BinanceSource) GetTimeFrameSeconds(timeFrame string) int64 {
	timeFrameMap := c.getValidTimeFrameMap()
	granularity, ok := timeFrameMap[timeFrame]
	if !ok {
		return 0
	}
	return int64(granularity)
}

func (c *BinanceSource) GetTimeFrameString(timeFrameSeconds int64) string {
	timeFrameMap := c.getValidTimeFrameMap()
	for k, v := range timeFrameMap {
		if int64(v) == timeFrameSeconds {
			return k
		}
	}
	return ""
}

func (c *BinanceSource) GetData(pairId string, timeFrame string, timeRange sources.TimeRange) ([]sources.ChartEntry, error) {
	timeFrameMap := c.getValidTimeFrameMap()
	granularity, ok := timeFrameMap[timeFrame]
	if !ok {
		return nil, errors.New("invalid timeFrame")
	}

	candles, err := c.fetchDataIfNeeded(pairId, granularity, timeRange)
	if err != nil {
		return nil, err
	}

	return candles, nil
}

func (c *BinanceSource) getValidTimeFrameMap() map[string]int {
	// Supported timeFrames by Binance API
	return map[string]int{
		"1m":  60,
		"3m":  180,
		"5m":  300,
		"15m": 900,
		"30m": 1800,
		"1h":  3600,
		"2h":  7200,
		"4h":  14400,
		"6h":  21600,
		"8h":  28800,
		"12h": 43200,
		"1d":  86400,
		"3d":  259200,
		"1w":  604800,
		"1M":  2592000,
	}
}

func (c *BinanceSource) fetchTradingPairsIfNeeded() {
	if len(c.TradingPairs) > 0 {
		return
	}

	url := c.BaseURL + "/exchangeInfo"

	pairs, _, err := sources.GetRequestParse[[]BinanceTradingPair](url, nil, "symbols")
	if err != nil {
		return
	}

	c.TradingPairs = *pairs
}

func (c *BinanceSource) fetchDataIfNeeded(pairId string, granularity int, timeRange sources.TimeRange) ([]sources.ChartEntry, error) {
	url := c.BaseURL + "/klines"

	// TODO - Implement rate limiting handling as per Binance API guidelines.

	const maxCandlesPerRequest = 1500 // Binance allows up to 1500 candles per request
	candleCount := int(timeRange.End.Sub(timeRange.Start).Seconds()) / granularity
	requestCount := (candleCount + maxCandlesPerRequest - 1) / maxCandlesPerRequest

	if requestCount == 0 {
		requestCount = 1
	}

	allCandles := make([]sources.ChartEntry, 0, candleCount)

	// Convert granularity to Binance timeFrame string
	timeFrameStr := ""
	for k, v := range c.getValidTimeFrameMap() {
		if v == granularity {
			timeFrameStr = k
			break
		}
	}

	// Fetch data in segments if needed
	for i := 0; i < requestCount; i++ {
		segmentStart := timeRange.Start.Add(time.Duration(i*maxCandlesPerRequest*granularity) * time.Second)
		segmentEnd := segmentStart.Add(time.Duration(maxCandlesPerRequest*granularity) * time.Second)

		if segmentEnd.After(timeRange.End) {
			segmentEnd = timeRange.End
		}

		params := map[string]string{
			"symbol":    pairId,
			"interval":  timeFrameStr,
			"startTime": strconv.FormatInt(segmentStart.UnixMilli(), 10),
			"endTime":   strconv.FormatInt(segmentEnd.UnixMilli(), 10),
			"limit":     strconv.Itoa(maxCandlesPerRequest),
		}

		candlesRaw, statusCode, err := sources.GetRequestParse[[][]interface{}](url, params)
		if err != nil {
			return nil, err
		}

		_ = statusCode // TODO - handle non-200 status codes

		candles := make([]sources.ChartEntry, 0, len(*candlesRaw))

		for _, candle := range *candlesRaw {
			if len(candle) < 6 {
				continue
			}

			timestamp := int64(candle[0].(float64)) / 1000
			open, _ := strconv.ParseFloat(candle[1].(string), 64)
			high, _ := strconv.ParseFloat(candle[2].(string), 64)
			low, _ := strconv.ParseFloat(candle[3].(string), 64)
			close, _ := strconv.ParseFloat(candle[4].(string), 64)
			volume, _ := strconv.ParseFloat(candle[5].(string), 64)

			candles = append(candles, sources.ChartEntry{
				Timestamp: time.Unix(timestamp, 0),
				Open:      open,
				High:      high,
				Low:       low,
				Close:     close,
				Volume:    volume,
			})
		}

		if candles != nil {
			allCandles = append(allCandles, candles...)
		}
	}

	return allCandles, nil
}
