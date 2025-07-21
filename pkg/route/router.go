
// Initialize 初始化路由
func Initialize() {
	Router = mux.NewRouter()
	routes.RegisterWebRoutes(Router)
}
