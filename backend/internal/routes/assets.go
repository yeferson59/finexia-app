package routes

func (r *Routes) Assets() {
	a := r.router.Group("/assets")
	a.Post("/sync", r.handlers.SyncAssetPrices)
}
