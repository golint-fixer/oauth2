package main

import (
	"fmt"
	"runtime"

	"gopkg.in/redis.v3"

	"github.com/RangelReale/osin"
	"github.com/codegangsta/cli"
	"github.com/iogo-framework/application"
	"github.com/iogo-framework/cmd"
	"github.com/iogo-framework/logs"
	"github.com/iogo-framework/router"
	"github.com/quorumsco/oauth2/components"
	"github.com/quorumsco/oauth2/controllers"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	cmd := cmd.New()
	cmd.Name = "oauth2"
	cmd.Usage = "Quorums OAuth2 server"
	cmd.Version = "0.0.1"
	cmd.Before = serve
	cmd.Flags = append(cmd.Flags, []cli.Flag{
		cli.StringFlag{Name: "listen-host", Value: "0.0.0.0", Usage: "server listening host", EnvVar: "LISTEN_HOST"},
		cli.IntFlag{Name: "listen-port", Value: 8080, Usage: "server listening port", EnvVar: "LISTEN_PORT"},

		cli.StringFlag{Name: "redis-host", Value: "redis", Usage: "redis host", EnvVar: "REDIS_HOST"},
		cli.IntFlag{Name: "redis-port", Value: 6379, Usage: "redis port", EnvVar: "REDIS_PORT"},

		cli.BoolFlag{Name: "debug, d", Usage: "print debug information", EnvVar: "DEBUG"},
		cli.HelpFlag,
	}...)
	cmd.RunAndExitOnError()
}

func serve(ctx *cli.Context) error {
	var app = application.New()

	if ctx.Bool("debug") {
		logs.Level(logs.DebugLevel)
	}

	redisHostPort := fmt.Sprintf("%s:%d", ctx.String("redis-host"), ctx.Int("redis-port"))
	client := redis.NewClient(&redis.Options{Addr: redisHostPort})
	if _, err := client.Ping().Result(); err != nil {
		return err
	}
	logs.Debug("Connected to Redis at %s", redisHostPort)
	app.Components["Redis"] = client

	cfg := osin.NewServerConfig()
	cfg.AllowedAuthorizeTypes = osin.AllowedAuthorizeType{osin.CODE, osin.TOKEN}
	cfg.AllowedAccessTypes = osin.AllowedAccessType{osin.AUTHORIZATION_CODE,
		osin.REFRESH_TOKEN, osin.PASSWORD}

	server := osin.NewServer(cfg, components.NewRedisStorage(client))
	app.Components["OAuth"] = server
	app.Components["Mux"] = router.New()

	if ctx.Bool("debug") {
		app.Use(router.Logger)
	}

	app.Use(app.Apply)
	app.Get("/authorize", controllers.Authorize)
	app.Post("/token", controllers.Token)
	app.Get("/info", controllers.Info)

	app.Get("/test", controllers.Test)

	return app.Serve(fmt.Sprintf("%s:%d", ctx.String("listen-host"), ctx.Int("listen-port")))
}
