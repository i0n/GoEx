package okcoin

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	. "github.com/nntaoli-project/GoEx"
)

const (
	EXCHANGE_NAME_COM = "okcoin.com"
)

type OKCoinCOM_API struct {
	OKCoinCN_API
}

func NewCOM(client *http.Client, api_key, secret_key string) *OKCoinCOM_API {
	return &OKCoinCOM_API{OKCoinCN_API{client, api_key, secret_key, "https://www.okcoin.com/api/v1/"}}
}

func (ctx *OKCoinCOM_API) GetAccount() (*Account, error) {
	postData := url.Values{}
	err := ctx.buildPostForm(&postData)
	if err != nil {
		return nil, err
	}

	body, err := HttpPostForm(ctx.client, ctx.api_base_url+url_userinfo, postData)
	if err != nil {
		return nil, err
	}

	//	println(string(body))

	var respMap map[string]interface{}

	err = json.Unmarshal(body, &respMap)
	if err != nil {
		return nil, err
	}

	if !respMap["result"].(bool) {
		errcode := strconv.FormatFloat(respMap["error_code"].(float64), 'f', 0, 64)
		return nil, errors.New(errcode)
	}

	info := respMap["info"].(map[string]interface{})
	funds := info["funds"].(map[string]interface{})
	asset := funds["asset"].(map[string]interface{})
	free := funds["free"].(map[string]interface{})
	freezed := funds["freezed"].(map[string]interface{})

	account := new(Account)
	account.Exchange = ctx.GetExchangeName()
	account.Asset, _ = strconv.ParseFloat(asset["total"].(string), 64)
	account.NetAsset, _ = strconv.ParseFloat(asset["net"].(string), 64)

	var btcSubAccount SubAccount
	var ltcSubAccount SubAccount
	var ethSubAccount SubAccount
	var etcSubAccount SubAccount
	var bchSubAccount SubAccount
	var usdSubAccount SubAccount

	btcSubAccount.Currency = BTC
	btcSubAccount.Amount, _ = strconv.ParseFloat(free["btc"].(string), 64)
	btcSubAccount.LoanAmount = 0
	btcSubAccount.FrozenAmount, _ = strconv.ParseFloat(freezed["btc"].(string), 64)

	bchSubAccount.Currency = BCH
	bchSubAccount.Amount, _ = strconv.ParseFloat(free["bch"].(string), 64)
	bchSubAccount.LoanAmount = 0
	bchSubAccount.FrozenAmount, _ = strconv.ParseFloat(freezed["bch"].(string), 64)

	ltcSubAccount.Currency = LTC
	ltcSubAccount.Amount, _ = strconv.ParseFloat(free["ltc"].(string), 64)
	ltcSubAccount.LoanAmount = 0
	ltcSubAccount.FrozenAmount, _ = strconv.ParseFloat(freezed["ltc"].(string), 64)

	etcSubAccount.Currency = ETC
	etcSubAccount.Amount, _ = strconv.ParseFloat(free["etc"].(string), 64)
	etcSubAccount.LoanAmount = 0
	etcSubAccount.FrozenAmount, _ = strconv.ParseFloat(freezed["etc"].(string), 64)

	ethSubAccount.Currency = ETH
	ethSubAccount.Amount, _ = strconv.ParseFloat(free["eth"].(string), 64)
	ethSubAccount.LoanAmount = 0
	ethSubAccount.FrozenAmount, _ = strconv.ParseFloat(freezed["eth"].(string), 64)

	usdSubAccount.Currency = USD
	usdSubAccount.Amount, _ = strconv.ParseFloat(free["usd"].(string), 64)
	usdSubAccount.LoanAmount = 0
	usdSubAccount.FrozenAmount, _ = strconv.ParseFloat(freezed["usd"].(string), 64)

	account.SubAccounts = make(map[Currency]SubAccount, 6)
	account.SubAccounts[BTC] = btcSubAccount
	account.SubAccounts[LTC] = ltcSubAccount
	account.SubAccounts[ETH] = ethSubAccount
	account.SubAccounts[ETC] = etcSubAccount
	account.SubAccounts[BCH] = bchSubAccount
	account.SubAccounts[USD] = usdSubAccount

	return account, nil
}

func (ctx *OKCoinCOM_API) GetExchangeName() string {
	return EXCHANGE_NAME_COM
}
