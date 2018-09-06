// Authentification management structure
package main

import (
	"os"
	"runtime"

	"gopkg.in/redis.v3"

	"github.com/RangelReale/osin"
	"github.com/codegangsta/cli"
	"github.com/jinzhu/gorm"
	"github.com/quorumsco/application"
	"github.com/quorumsco/cmd"
	"github.com/quorumsco/databases"
	"github.com/quorumsco/gojimux"
	"github.com/quorumsco/logs"
	"github.com/quorumsco/oauth2/components"
	"github.com/quorumsco/oauth2/controllers"
	"github.com/quorumsco/oauth2/models"
	"github.com/quorumsco/oauth2/views"
	"github.com/quorumsco/router"
	"github.com/quorumsco/settings"
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
		cli.StringFlag{Name: "config, c", Usage: "configuration file", EnvVar: "CONFIG"},
		cli.HelpFlag,
	}...)
	cmd.RunAndExitOnError()
}

func serve(ctx *cli.Context) error {
	var err error

	var config settings.Config
	if ctx.String("config") != "" {
		config, err = settings.Parse(ctx.String("config"))
		if err != nil {
			logs.Error(err)
		}
	}

	if config.Debug() {
		logs.Level(logs.DebugLevel)
	}
	smtpSettings, err := config.Smtp()
	logs.Info(smtpSettings.User)

	dialect, args, err := config.SqlDB()
	if err != nil {
		logs.Critical(err)
		os.Exit(1)
	}
	logs.Debug("database type: %s", dialect)

	var app = application.New()
	if app.Components["DB"], err = databases.InitGORM(dialect, args); err != nil {
		logs.Critical(err)
		os.Exit(1)
	}
	logs.Debug("connected to %s", args)

	if config.Migrate() {
		app.Components["DB"].(*gorm.DB).AutoMigrate(models.Models()...)
		logs.Debug("database migrated successfully")
	}

	redisSettings, err := config.Redis()
	client := redis.NewClient(&redis.Options{Addr: redisSettings.String()})
	if _, err := client.Ping().Result(); err != nil {
		return err
	}
	logs.Debug("Connected to Redis at %s", redisSettings.String())

	app.Components["Redis"] = client

	cfg := osin.NewServerConfig()
	cfg.AllowedAuthorizeTypes = osin.AllowedAuthorizeType{osin.CODE, osin.TOKEN}
	cfg.AllowedAccessTypes = osin.AllowedAccessType{osin.AUTHORIZATION_CODE,
		osin.REFRESH_TOKEN, osin.PASSWORD}
	cfg.AccessExpiration = 3600 * 10
	// cfg.AccessExpiration = 30

	oauthServer := osin.NewServer(cfg, components.NewRedisStorage(client))
	app.Components["OAuth"] = oauthServer

	app.Components["Templates"] = views.Templates()

	app.Components["Smtp"] = smtpSettings

	app.Components["Mux"] = gojimux.New()

	if config.Debug() {
		app.Components["DB"].(*gorm.DB).LogMode(true)
		app.Use(router.Logger)
	}

	app.Use(app.Apply)

	app.Get("/oauth2/authorize", controllers.Authorize)
	app.Post("/oauth2/token", controllers.Token)
	app.Get("/oauth2/info", controllers.Info)
	//app.Get("/oauth2/test_ladon"),controllers.TestLadon)

	app.Post("/users/register", controllers.Register)
	app.Post("/users/newregister", controllers.NewRegister)
	app.Post("/users/registerFromAdmin", controllers.RegisterFromAdmin)
	app.Get("/users/:id", controllers.RetrieveUser)
	app.Post("/users_all/:id", controllers.RetrieveAllUsersByGroup)
	//app.Post("/users_team/:id", controllers.RetrieveAllUsersByTeam)
	app.Patch("/users/update", controllers.Update)
	app.Delete("/users/:id", controllers.Delete)

	//for compatibility v0 MOBILE - IONIC
	app.Post("/users/validPasswordForOldVersion", controllers.ValidPassword)
	app.Post("/users/updatePasswordForOldVersion", controllers.UpdatePassword)

	//for new versions -> webapp and mobile V1 (react native)
	app.Post("/users/savePassword", controllers.NewSavePassword)
	app.Post("/users/updatePassword", controllers.SendMailWithUrlForPasswordChange)

	app.Post("/users/sendrequesttoreferent", controllers.SendRequestToReferent)
	app.Post("/users/validUser", controllers.ValidUser)
	app.Post("/users/existMail", controllers.ExistMail)

	app.Get("/groups", controllers.RetrieveGroupCollection)
	app.Get("/groups/retrieve_mail_referent/:cause", controllers.RetrieveGroupByCode_cause)
	app.Post("/groups", controllers.CreateGroup)
	app.Get("/groups/:id", controllers.RetrieveGroup)
	app.Delete("/groups/:id", controllers.DeleteGroup)
	app.Patch("/groups/:id", controllers.UpdateGroup)

	//app.Get("/teams", controllers.RetrieveTeamCollection)
	//app.Post("/teams", controllers.CreateTeam)
	//app.Get("/teams/:id", controllers.RetrieveTeam)
	//app.Delete("/teams/:id", controllers.DeleteTeam)
	//app.Patch("/teams/:id", controllers.UpdateTeam)
	//app.Get("/teams/retrieve_team_group/:id", controllers.RetrieveTeamByGroupID)

	server, err := config.Server()
	if err != nil {
		logs.Critical(err)
		os.Exit(1)
	}
	return app.Serve(server.String())
}
