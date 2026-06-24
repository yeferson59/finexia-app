package routes

func (r *Routes) ExchangeRates() {
	er := r.router.Group("/exchange-rates")
	er.Post("/sync", r.handlers.SyncExchangeRates)
}
