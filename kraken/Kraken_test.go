package kraken

import (
  "errors"
	"io/ioutil"
	"gopkg.in/yaml.v2"
  "bitbucket.org/i0n/compounda/addresses"
	"github.com/i0n/GoEx"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

var configFilePath = "../api-keys.yml"

// Manifest struct representing YAML api key manifest
type Manifest struct {
	APIVersion  int                       `yaml:"api_version"`
	Exchanges   map[string]ExchangeKeys   `yaml:"exchanges"`
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
var k = New(http.DefaultClient, manifest.Exchanges["kraken.com"].AccessKey, manifest.Exchanges["kraken.com"].SecretKey)

var BCH_XBT = goex.NewCurrencyPair(goex.BCH, goex.XBT)

func TestKraken_GetDepth_Bid(t *testing.T) {
	dep, err := k.GetDepth(2, goex.BTC_USD)
	assert.Nil(t, err)
  assert.True(t, dep.BidList[0].Price > dep.BidList[1].Price)
	t.Log(dep)
}

func TestKraken_GetDepth_Ask(t *testing.T) {
	dep, err := k.GetDepth(2, goex.BTC_USD)
	assert.Nil(t, err)
  assert.True(t, dep.AskList[0].Price < dep.AskList[1].Price)
	t.Log(dep)
}

func TestKraken_GetTicker(t *testing.T) {
	ticker, err := k.GetTicker(goex.ETC_BTC)
	assert.Nil(t, err)
	t.Log(ticker)
}

func TestKraken_GetAccount(t *testing.T) {
	acc, err := k.GetAccount()
	assert.Nil(t, err)
	t.Log(acc)
}

// Test for error...
func TestKraken_LimitSell(t *testing.T) {
	ord, err := k.LimitSell("1000000", "690000", goex.BTC_USD)
	assert.Equal(t, errors.New("EOrder:Insufficient funds"), err)
	t.Log(ord)
}

// Test for error...
func TestKraken_LimitBuy(t *testing.T) {
	ord, err := k.LimitBuy("1000000", "61", goex.NewCurrencyPair(goex.XBT, goex.USD))
	assert.Equal(t, errors.New("EOrder:Insufficient funds"), err)
	t.Log(ord)
}

func TestKraken_GetUnfinishOrders(t *testing.T) {
	ords, err := k.GetUnfinishOrders(goex.NewCurrencyPair(goex.XBT, goex.USD))
	assert.Nil(t, err)
	t.Log(ords)
}

// Test for error...
func TestKraken_CancelOrder(t *testing.T) {
	r, err := k.CancelOrder("O6EAJC-YAC3C-XDEEXQ", goex.NewCurrencyPair(goex.XBT, goex.USD))
	assert.Equal(t, errors.New("EOrder:Unknown order"), err)
	t.Log(r)
}

// Test for error...
func TestKraken_GetOneOrder(t *testing.T) {
	ord, err := k.GetOneOrder("ODCRMQ-RDEID-CY334C", goex.BTC_USD)
	assert.Equal(t, errors.New("Could not find the order ODCRMQ-RDEID-CY334C"), err)
	t.Log(ord)
}

func TestKraken_Withdraw(t *testing.T) {
  _, err := k.Withdraw(goex.ETC_USD, addresses.All["okcoin.com"][goex.ETC], 0.1, "", "")
	assert.Equal(t, errors.New("EFunding:Invalid amount"), err)
}

// TODO Write more tests
