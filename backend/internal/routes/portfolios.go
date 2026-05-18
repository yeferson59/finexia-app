package routes

func (r *Routes) Portfolios() {
	portfolios := r.router.Group("/portfolios")

	portfolios.Get("/risks", r.handlers.GetPortfoliosRisks)
	portfolios.Get("/id", r.handlers.GetPortfolios)
	portfolios.Post("", r.handlers.CreatePortfolio)
}
