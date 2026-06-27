package routes

func (r *Routes) Assets() {
	a := r.router.Group("/assets")
	a.Post("/sync", r.middlewares.RequireAdmin(), r.handlers.SyncAssetPrices)
}
