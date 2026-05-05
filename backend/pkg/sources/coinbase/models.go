package coinbase

import "fmt"

type CoinbaseAPIError struct {
	Message string `json:"message"`
}

func (e CoinbaseAPIError) Error() error {
	return fmt.Errorf("Coinbase API error: %s", e.Message)
}

type CoinbaseTradingPair struct {
	Id                     string `json:"id"`
	Name                   string `json:"display_name"`
	Base                   string `json:"base_currency"`
	Quote                  string `json:"quote_currency"`
	Status                 string `json:"status"`
	QuoteIncrement         string `json:"quote_increment"`
	BaseIncrement          string `json:"base_increment"`
	MinMarketFunds         string `json:"min_market_funds"`
	MarginEnabled          bool   `json:"margin_enabled"`
	PostOnly               bool   `json:"post_only"`
	LimitOnly              bool   `json:"limit_only"`
	CancelOnly             bool   `json:"cancel_only"`
	StatusMessage          string `json:"status_message"`
	TradingDisabled        bool   `json:"trading_disabled"`
	FxStablecoin           bool   `json:"fx_stablecoin"`
	MaxSlippagePercentage  string `json:"max_slippage_percentage"`
	AuctionMode            bool   `json:"auction_mode"`
	HighBidLimitPercentage string `json:"high_bid_limit_percentage"`
}
