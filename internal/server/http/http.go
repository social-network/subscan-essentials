package http

import (
	"github.com/go-kratos/kratos/pkg/conf/paladin"
	bm "github.com/go-kratos/kratos/pkg/net/http/blademaster"
	"github.com/social-network/subscan/internal/middleware"
	"github.com/social-network/subscan/internal/service"
	"github.com/social-network/subscan/plugins"
)

var (
	svc *service.Service
)

func New(s *service.Service) (engine *bm.Engine) {
	var (
		hc struct {
			Server *bm.ServerConfig
		}
	)
	if err := paladin.Get("http.toml").UnmarshalTOML(&hc); err != nil {
		if err != paladin.ErrNotExist {
			panic(err)
		}
	}
	svc = s
	engine = bm.DefaultServer(hc.Server)
	engine.HandleMethodNotAllowed = false

	initRouter(engine)

	if err := engine.Start(); err != nil {
		panic(err)
	}
	return engine
}

func initRouter(e *bm.Engine) {
	limiter := bm.NewRateLimiter(nil)
	e.Use(limiter.Limit(), middlewares.CORS())

	e.Ping(ping)
	// internal
	g := e.Group("/api")
	{
		g.GET("system/status", systemStatus)
		g.POST("/now", now)
		s := g.Group("/scan")
		{
			s.POST("metadata", metadata)
			// Block
			s.POST("blocks", blocks)
			s.POST("block", block)

			// Extrinsic
			s.POST("extrinsics", extrinsics)
			s.POST("extrinsic", extrinsic)
			// Event
			s.POST("events", events)

			s.POST("check_hash", checkSearchHash)

			s.POST("runtime/metadata", runtimeMetadata)
			s.POST("runtime/list", runtimeList)
		}
		pluginRouter(g)
	}
}

func pluginRouter(g *bm.RouterGroup) {
	for name, plugin := range plugins.RegisteredPlugins {
		for _, r := range plugin.InitHttp() {
			g.Group("plugin").Group(name).POST(r.Router, func(context *bm.Context) {
				_ = r.Handle(context.Writer, context.Request)
			})
		}
	}
}
