package controllers

import (
	"net/http"

	"github.com/RangelReale/osin"
	"github.com/RangelReale/osin/example"
	"github.com/iogo-framework/application"
	"github.com/iogo-framework/logs"
	"github.com/iogo-framework/router"
)

func OAuthComponent(r *http.Request) *osin.Server {
	return router.GetContext(r).Env["Application"].(*application.Application).Components["OAuth"].(*osin.Server)
}

// Authorize endpoint
func Authorize(w http.ResponseWriter, r *http.Request) {
	server := OAuthComponent(r)
	resp := server.NewResponse()
	defer resp.Close()

	if ar := server.HandleAuthorizeRequest(resp, r); ar != nil {
		if !example.HandleLoginPage(ar, w, r) {
			return
		}
		ar.UserData = struct{ Login string }{Login: "test"}
		ar.Authorized = true
		server.FinishAuthorizeRequest(resp, r, ar)
	}
	if resp.IsError && resp.InternalError != nil {
		logs.Error("%s\n", resp.InternalError)
	}
	osin.OutputJSON(resp, w, r)
}

// Token endpoint
func Token(w http.ResponseWriter, r *http.Request) {
	server := OAuthComponent(r)
	resp := server.NewResponse()
	defer resp.Close()

	if ar := server.HandleAccessRequest(resp, r); ar != nil {
		switch ar.Type {
		case osin.AUTHORIZATION_CODE:
			ar.Authorized = true
		case osin.REFRESH_TOKEN:
			ar.Authorized = true
		case osin.PASSWORD:
			if ar.Username == "test" && ar.Password == "test" {
				ar.Authorized = true
			}
		}
		server.FinishAccessRequest(resp, r, ar)
	}
	if resp.IsError && resp.InternalError != nil {
		logs.Error("%s\n", resp.InternalError)
	}
	osin.OutputJSON(resp, w, r)
}
