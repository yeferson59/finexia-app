package routes

func (r *Routes) Assets() {
	a := r.router.Group("/assets")
	a.Post("", r.middlewares.RequireAdmin(), r.handlers.CreateAsset)
	a.Post("/sync", r.middlewares.RequireAdmin(), r.handlers.SyncAssetPrices)
	a.Post("/:id/sync", r.middlewares.RequireAdmin(), r.handlers.SyncSingleAsset)
}
