package factory

import (
	"leaseCar/payment-service/internal/adapters"
	"leaseCar/payment-service/internal/strategies"
	cfg "leaseCar/utils/config"
)

type PaymentFactory struct {
	conf *cfg.Config
}

func NewPaymentFactory(conf *cfg.Config) *PaymentFactory { return &PaymentFactory{conf: conf} }

func (f *PaymentFactory) GetStrategy(provider string) strategies.PaymentStrategy {
	switch provider {
	case "stripe":
		apiKey := ""
		if f.conf != nil {
			// try to read if available
		}
		return strategies.NewStripeStrategy(apiKey)
	case "bank_api":
		url := ""
		apiKey := ""
		if f.conf != nil {
			url = f.conf.MeiliSearch.URL // placeholder - config must be extended per provider
		}
		adapter := adapters.NewBankAdapter(url, apiKey)
		return strategies.NewBankStrategy(adapter)
	default:
		return nil
	}
}
