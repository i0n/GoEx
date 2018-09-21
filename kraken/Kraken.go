package kraken

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/nntaoli-project/GoEx"
)

type BaseResponse struct {
	Error  []string    `json:"error"`
	Result interface{} `json:"result"`
}

type NewOrderResponse struct {
	Description interface{} `json:"descr"`
	TxIds       []string    `json:"txid"`
}

type Kraken struct {
	httpClient *http.Client
	accessKey,
	secretKey string
}

type WithdrawStatus struct {
	Method string `json:"method"`
	AClass string `json:"aclass"`
	Asset  string `json:"asset"`
	RefID  string `json:"refid"`
	TXID   string `json:"txid"`
	Info   string `json:"vol"`
	Status string `json:"status"`
}

var (
	BASE_URL   = "https://api.kraken.com"
	API_V0     = "/0/"
	API_DOMAIN = BASE_URL + API_V0
	PUBLIC     = "public/"
	PRIVATE    = "private/"
)

func New(client *http.Client, accesskey, secretkey string) *Kraken {
	return &Kraken{client, accesskey, secretkey}
}

func (k *Kraken) placeOrder(orderType, side, amount, price string, pair goex.CurrencyPair) (*goex.Order, error) {
	apiuri := "private/AddOrder"

	params := url.Values{}
	params.Set("pair", k.convertPair(pair).ToSymbol(""))
	params.Set("type", side)
	params.Set("ordertype", orderType)
	params.Set("price", price)
	params.Set("volume", amount)

	var resp NewOrderResponse
	err := k.doAuthenticatedRequest("POST", apiuri, params, &resp)
	//log.Println
	if err != nil {
		return nil, err
	}

	var tradeSide goex.TradeSide = goex.SELL
	if "buy" == side {
		tradeSide = goex.BUY
	}

	return &goex.Order{
		Currency: pair,
		OrderID2: resp.TxIds[0],
		Amount:   goex.ToFloat64(amount),
		Price:    goex.ToFloat64(price),
		Side:     tradeSide,
		Status:   goex.ORDER_UNFINISH}, nil
}

func (k *Kraken) LimitBuy(amount, price string, currency goex.CurrencyPair) (*goex.Order, error) {
	return k.placeOrder("limit", "buy", amount, price, currency)
}

func (k *Kraken) LimitSell(amount, price string, currency goex.CurrencyPair) (*goex.Order, error) {
	return k.placeOrder("limit", "sell", amount, price, currency)
}

func (k *Kraken) MarketBuy(amount, price string, currency goex.CurrencyPair) (*goex.Order, error) {
	return k.placeOrder("market", "buy", amount, price, currency)
}

func (k *Kraken) MarketSell(amount, price string, currency goex.CurrencyPair) (*goex.Order, error) {
	return k.placeOrder("market", "sell", amount, price, currency)
}

func (k *Kraken) CancelOrder(orderId string, currency goex.CurrencyPair) (bool, error) {
	params := url.Values{}
	apiuri := "private/CancelOrder"
	params.Set("txid", orderId)

	var respmap map[string]interface{}
	err := k.doAuthenticatedRequest("POST", apiuri, params, &respmap)
	if err != nil {
		return false, err
	}
	//log.Println(respmap)
	return true, nil
}

func (k *Kraken) toOrder(orderinfo interface{}) goex.Order {
	omap := orderinfo.(map[string]interface{})
	descmap := omap["descr"].(map[string]interface{})
	return goex.Order{
		Amount:     goex.ToFloat64(omap["vol"]),
		Price:      goex.ToFloat64(descmap["price"]),
		DealAmount: goex.ToFloat64(omap["vol_exec"]),
		AvgPrice:   goex.ToFloat64(omap["price"]),
		Side:       k.convertSide(descmap["type"].(string)),
		Status:     k.convertOrderStatus(omap["status"].(string)),
		OrderTime:  goex.ToInt(omap["opentm"]),
	}
}

func (k *Kraken) GetOrderInfos(txids ...string) ([]goex.Order, error) {
	params := url.Values{}
	params.Set("txid", strings.Join(txids, ","))

	var resultmap map[string]interface{}
	err := k.doAuthenticatedRequest("POST", "private/QueryOrders", params, &resultmap)
	if err != nil {
		return nil, err
	}

	var ords []goex.Order
	for txid, v := range resultmap {
		ord := k.toOrder(v)
		ord.OrderID2 = txid
		ords = append(ords, ord)
	}

	return ords, nil
}

func (k *Kraken) GetOneOrder(orderId string, currency goex.CurrencyPair) (*goex.Order, error) {
	orders, err := k.GetOrderInfos(orderId)

	if err != nil {
		return nil, err
	}

	if len(orders) == 0 {
		return nil, errors.New("Could not find the order " + orderId)
	}

	ord := &orders[0]
	ord.Currency = currency
	return ord, nil
}

func (k *Kraken) GetUnfinishOrders(currency goex.CurrencyPair) ([]goex.Order, error) {
	var result struct {
		Open map[string]interface{} `json:"open"`
	}

	err := k.doAuthenticatedRequest("POST", "private/OpenOrders", url.Values{}, &result)
	if err != nil {
		return nil, err
	}

	var orders []goex.Order

	for txid, v := range result.Open {
		ord := k.toOrder(v)
		ord.OrderID2 = txid
		ord.Currency = currency
		orders = append(orders, ord)
	}

	return orders, nil
}

func (k *Kraken) GetOrderHistorys(currency goex.CurrencyPair, currentPage, pageSize int) ([]goex.Order, error) {
	panic("")
}

func (k *Kraken) GetAccount() (*goex.Account, error) {
	params := url.Values{}
	apiuri := "private/Balance"

	var resustmap map[string]interface{}
	err := k.doAuthenticatedRequest("POST", apiuri, params, &resustmap)
	if err != nil {
		return nil, err
	}

	acc := new(goex.Account)
	acc.Exchange = k.GetExchangeName()
	acc.SubAccounts = make(map[goex.Currency]goex.SubAccount)

	for key, v := range resustmap {
		currency := k.convertCurrency(key)
		amount := goex.ToFloat64(v)
		//log.Println(symbol, amount)
		acc.SubAccounts[currency] = goex.SubAccount{Currency: currency, Amount: amount, FrozenAmount: 0, LoanAmount: 0}

		if currency.Symbol == "XBT" {
			acc.SubAccounts[goex.BTC] = goex.SubAccount{Currency: goex.BTC, Amount: amount, FrozenAmount: 0, LoanAmount: 0}
		}
	}

	return acc, nil

}

func (k *Kraken) Withdraw(currencyPair goex.CurrencyPair, address goex.CryptoAddressReader, amount float64, wallet string, adminPassword string) (*goex.Withdraw, error) {
	apiuri := "private/Withdraw"
	params := url.Values{}
	a := strconv.FormatFloat(amount, 'f', -1, 64)
	params.Set("amount", a)
	params.Set("asset", currencyPair.CurrencyA.String())
	params.Set("key", address.Tag())
	var result goex.Withdraw
	err := k.doAuthenticatedRequest("POST", apiuri, params, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (k *Kraken) WithdrawStatus(currency goex.Currency) (*[]WithdrawStatus, error) {
	apiuri := "private/WithdrawStatus"
	params := url.Values{}
	params.Set("asset", currency.String())
	var result []WithdrawStatus
	err := k.doAuthenticatedRequest("POST", apiuri, params, &result)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("resultmap: %v#\n", resultmap[0])
	return &result, nil
}

func (k *Kraken) GetTicker(currency goex.CurrencyPair) (*goex.Ticker, error) {
	var resultmap map[string]interface{}
	err := k.doAuthenticatedRequest("GET", "public/Ticker?pair="+k.convertPair(currency).ToSymbol(""), url.Values{}, &resultmap)
	if err != nil {
		return nil, err
	}

	ticker := new(goex.Ticker)
	for _, t := range resultmap {
		tickermap := t.(map[string]interface{})
		ticker.Last = goex.ToFloat64(tickermap["c"].([]interface{})[0])
		ticker.Buy = goex.ToFloat64(tickermap["b"].([]interface{})[0])
		ticker.Sell = goex.ToFloat64(tickermap["a"].([]interface{})[0])
		ticker.Low = goex.ToFloat64(tickermap["l"].([]interface{})[0])
		ticker.High = goex.ToFloat64(tickermap["h"].([]interface{})[0])
		ticker.Vol = goex.ToFloat64(tickermap["v"].([]interface{})[0])
	}

	return ticker, nil
}

func (k *Kraken) GetDepth(size int, currency goex.CurrencyPair) (*goex.Depth, error) {
	apiuri := fmt.Sprintf("public/Depth?pair=%s&count=%d", k.convertPair(currency).ToSymbol(""), size)
	var resultmap map[string]interface{}
	err := k.doAuthenticatedRequest("GET", apiuri, url.Values{}, &resultmap)
	if err != nil {
		return nil, err
	}

	//log.Println(respmap)
	dep := goex.Depth{}
	for _, d := range resultmap {
		depmap := d.(map[string]interface{})
		asksmap := depmap["asks"].([]interface{})
		bidsmap := depmap["bids"].([]interface{})
		for _, v := range asksmap {
			ask := v.([]interface{})
			dep.AskList = append(dep.AskList, goex.DepthRecord{
				Price:  goex.ToFloat64(ask[0]),
				Amount: goex.ToFloat64(ask[1])},
			)
		}
		for _, v := range bidsmap {
			bid := v.([]interface{})
			dep.BidList = append(dep.BidList, goex.DepthRecord{
				Price:  goex.ToFloat64(bid[0]),
				Amount: goex.ToFloat64(bid[1])},
			)
		}
		break
	}
	return &dep, nil
}

func (k *Kraken) GetKlineRecords(currency goex.CurrencyPair, period, size, since int) ([]goex.Kline, error) {
	panic("")
}

//非个人，整个交易所的交易记录
func (k *Kraken) GetTrades(currencyPair goex.CurrencyPair, since int64) ([]goex.Trade, error) {
	panic("")
}

func (k *Kraken) GetExchangeName() string {
	return "kraken.com"
}

func (k *Kraken) buildParamsSigned(apiuri string, postForm *url.Values) string {
	postForm.Set("nonce", fmt.Sprintf("%d", time.Now().UnixNano()))
	urlPath := API_V0 + apiuri

	secretByte, _ := base64.StdEncoding.DecodeString(k.secretKey)
	encode := []byte(postForm.Get("nonce") + postForm.Encode())

	sha := sha256.New()
	sha.Write(encode)
	shaSum := sha.Sum(nil)

	pathSha := append([]byte(urlPath), shaSum...)

	mac := hmac.New(sha512.New, secretByte)
	mac.Write(pathSha)
	macSum := mac.Sum(nil)

	sign := base64.StdEncoding.EncodeToString(macSum)

	return sign
}

func (k *Kraken) doAuthenticatedRequest(method, apiuri string, params url.Values, ret interface{}) error {
	headers := map[string]string{}

	if "POST" == method {
		signature := k.buildParamsSigned(apiuri, &params)
		headers = map[string]string{
			"API-Key":  k.accessKey,
			"API-Sign": signature,
		}
	}

	resp, err := goex.NewHttpRequest(k.httpClient, method, API_DOMAIN+apiuri, params.Encode(), headers)
	if err != nil {
		return err
	}
	//println(string(resp))
	var base BaseResponse
	base.Result = ret

	err = json.Unmarshal(resp, &base)
	if err != nil {
		return err
	}

	//println(string(resp))

	if len(base.Error) > 0 {
		return errors.New(base.Error[0])
	}

	return nil
}

func (k *Kraken) convertCurrency(currencySymbol string) goex.Currency {
	if len(currencySymbol) >= 4 {
		currencySymbol = strings.Replace(currencySymbol, "X", "", 1)
		currencySymbol = strings.Replace(currencySymbol, "Z", "", 1)
	}
	return goex.NewCurrency(currencySymbol, "")
}

func (k *Kraken) convertPair(pair goex.CurrencyPair) goex.CurrencyPair {
	if "BTC" == pair.CurrencyA.Symbol {
		return goex.NewCurrencyPair(goex.XBT, pair.CurrencyB)
	}

	if "BTC" == pair.CurrencyB.Symbol {
		return goex.NewCurrencyPair(pair.CurrencyA, goex.XBT)
	}

	return pair
}

func (k *Kraken) convertSide(typeS string) goex.TradeSide {
	switch typeS {
	case "sell":
		return goex.SELL
	case "buy":
		return goex.BUY
	}
	return goex.SELL
}

func (k *Kraken) convertOrderStatus(status string) goex.TradeStatus {
	switch status {
	case "open", "pending", "expired":
		return goex.ORDER_UNFINISH
	case "canceled", "closed":
		return goex.ORDER_CANCEL
	case "filled":
		return goex.ORDER_FINISH
	case "partialfilled":
		return goex.ORDER_PART_FINISH
	}
	return goex.ORDER_UNFINISH
}
