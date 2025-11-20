package binance

type BinanceTradingPair struct {
	Symbol             string `json:"symbol"`
	Pair               string `json:"pair"`
	ContractType       string `json:"contractType"`
	DeliveryDate       int64  `json:"deliveryDate"`
	OnboardDate        int64  `json:"onboardDate"`
	Status             string `json:"status"`
	BaseAsset          string `json:"baseAsset"`
	QuoteAsset         string `json:"quoteAsset"`
	MarginAsset        string `json:"marginAsset"`
	PricePrecision     int    `json:"pricePrecision"`
	QuantityPrecision  int    `json:"quantityPrecision"`
	BaseAssetPrecision int    `json:"baseAssetPrecision"`
	QuotePrecision     int    `json:"quotePrecision"`
	UnderlyingType     string `json:"underlyingType"`
	TriggerProtect     string `json:"triggerProtect"`
	LiquidationFee     string `json:"liquidationFee"`
	MarketTakeBound    string `json:"marketTakeBound"`
	MaxMoveOrderLimit  int    `json:"maxMoveOrderLimit"`
}
