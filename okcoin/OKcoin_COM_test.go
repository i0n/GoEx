package okcoin

import (
  "errors"
	"io/ioutil"
	"net/http"
	"testing"
  "bitbucket.org/i0n/compounda/addresses"
	"github.com/i0n/GoEx"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

var configFilePath = "../api-keys.yml"

// Manifest struct representing YAML api key manifest
type Manifest struct {
	APIVersion  int                       `yaml:"api_version"`
	Exchanges   map[string]ExchangeKeys   `yaml:"exchanges"`
}

// ExchangeKeys struct holds keys for each exchanges API
type ExchangeKeys struct {
  AccessKey     string `yaml:"access_key"`
  SecretKey     string `yaml:"secret_key"`
  AdminPassword string `yaml:"admin_password"`
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

var okcom = NewCOM(http.DefaultClient, manifest.Exchanges["okcoin.com"].AccessKey, manifest.Exchanges["okcoin.com"].SecretKey)

func TestOKCoinCOM_API_GetTicker(t *testing.T) {
	ticker, err := okcom.GetTicker(goex.BTC_USD)
	assert.Nil(t, err)
	t.Log(ticker)
}

func TestOKCoinCOM_API_GetDepth_Bid(t *testing.T) {
	dep, err := okcom.GetDepth(2, goex.BTC_USD)
	assert.Nil(t, err)
  assert.True(t, dep.BidList[0].Price > dep.BidList[1].Price)
	t.Log(dep)
}

func TestOKCoinCOM_API_GetDepth_Ask(t *testing.T) {
	dep, err := okcom.GetDepth(2, goex.BTC_USD)
	assert.Nil(t, err)
  assert.True(t, dep.AskList[0].Price < dep.AskList[1].Price)
	t.Log(dep)
}

func TestOKCoinCOM_API_Withdraw(t *testing.T) {
  _, err := okcom.Withdraw(goex.ETC_USD, addresses.All["kraken.com"][goex.ETC], 0.1, "", manifest.Exchanges["okcoin.com"].AdminPassword)
	assert.Equal(t, errors.New("10035"), err)
}
