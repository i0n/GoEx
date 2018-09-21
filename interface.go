package goex

// CryptoAddressReader is an interface for reading crpto address information
type CryptoAddressReader interface {
	Currency() Currency
	Address() string
	Tag() string
	ExchangeName() string
}
