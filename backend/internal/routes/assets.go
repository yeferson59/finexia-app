package routes

func (r *Routes) Assets() {
	a := r.router.Group("/assets")
	a.Post("", r.middlewares.RequireAdmin(), r.handlers.CreateAsset)
	a.Post("/sync", r.middlewares.RequireAdmin(), r.handlers.SyncAssetPrices)
}
