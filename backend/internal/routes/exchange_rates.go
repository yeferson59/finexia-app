package routes

func (r *Routes) ExchangeRates() {
	er := r.router.Group("/exchange-rates")
	er.Post("/sync", r.middlewares.RequireAdmin(), r.handlers.SyncExchangeRates)
}
