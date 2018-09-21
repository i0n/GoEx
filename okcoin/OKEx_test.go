package okcoin

import (
	"net/http"
	"testing"

	. "github.com/nntaoli-project/GoEx"
	"github.com/stretchr/testify/assert"
)

var (
	okex = NewOKEx(http.DefaultClient, "", "")
)

func TestOKEx_GetFutureDepth(t *testing.T) {
	dep, err := okex.GetFutureDepth(BTC_USD, QUARTER_CONTRACT, 1)
	assert.Nil(t, err)
	t.Log(dep)
}
