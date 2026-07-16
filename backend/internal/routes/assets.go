package routes

func (r *Routes) Assets() {
	a := r.router.Group("/assets")
	a.Post("", r.auth.RequireAdmin(), r.handlers.CreateAsset)
	a.Post("/import", r.auth.RequireAdmin(), r.handlers.ImportAssets)
	a.Post("/sync", r.auth.RequireAdmin(), r.handlers.SyncAssetPrices)
	a.Post("/:id/sync", r.auth.RequireAdmin(), r.handlers.SyncSingleAsset)
}
