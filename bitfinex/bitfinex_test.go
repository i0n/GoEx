package bitfinex

import (
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	addresses "github.com/i0n/crypto-addresses"
	goex "github.com/nntaoli-project/GoEx"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

var configFilePath = "../api-keys.yml"

// Manifest struct representing YAML api key manifest
type Manifest struct {
	APIVersion int                     `yaml:"api_version"`
	Exchanges  map[string]ExchangeKeys `yaml:"exchanges"`
}

// ExchangeKeys struct holds keys for each exchanges API
type ExchangeKeys struct {
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
}

func GetManifest(configFilePath *string) *Manifest {
	manifest := &Manifest{}
	configFile, err := ioutil.ReadFile(*configFilePath)
	if err != nil {
		panic(err)
	}
	yaml.Unmarshal(configFile, &manifest)
	return manifest
}

var manifest = GetManifest(&configFilePath)
var bfx = New(http.DefaultClient, manifest.Exchanges["bitfinex.com"].AccessKey, manifest.Exchanges["bitfinex.com"].SecretKey)

func TestBitfinex_GetTicker(t *testing.T) {
	ticker, _ := bfx.GetTicker(goex.ETH_BTC)
	t.Log(ticker)
}

func TestBitfinex_GetDepth_Bid(t *testing.T) {
	dep, _ := bfx.GetDepth(2, goex.ETH_BTC)
	assert.True(t, dep.BidList[0].Price > dep.BidList[1].Price)
	t.Log(dep.BidList)
}

func TestBitfinex_GetDepth_Ask(t *testing.T) {
	dep, _ := bfx.GetDepth(2, goex.ETH_BTC)
	assert.True(t, dep.AskList[0].Price < dep.AskList[1].Price)
	t.Log(dep.AskList)
}

func TestBitfinex_Withdraw(t *testing.T) {
	_, err := bfx.Withdraw(goex.ETC_USD, addresses.All["okcoin.com"][goex.ETC], 0.1, "exchange", "")
	assert.Equal(t, errors.New("Min 250 USD Equivalent"), err)
}

// TODO Write more tests
