## Go charting
Try out golang plotting library: https://github.com/go-echarts/go-echarts
- Kline charts are the main focus (candlestick)
    - In the example code `page_center_layout.html` has some cool stuff too
- Line charts with indicators (moving averages, RSI, MACD, etc)
- ThemeRiver charts for volume and other indicators

## Data Sources
- Coinbase API: DONE
- Binance API: DONE
- Bitstamp API: https://www.bitstamp.net/api/#tag/Market-info/operation/GetOHLCData
- Bybit API : https://bybit-exchange.github.io/docs/v5/market/kline
- Kraken API: https://docs.kraken.com/api/docs/rest-api/get-ohlc-data
    - Note: Kraken only returns the last 720 candles per request, older data cannot be retrieved
    - Historical data can be downloaded: https://support.kraken.com/articles/360047124832-downloadable-historical-ohlcvt-open-high-low-close-volume-trades-data
