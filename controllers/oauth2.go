package controllers

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/RangelReale/osin"
	"github.com/RangelReale/osin/example"
	"github.com/iogo-framework/application"
	"github.com/iogo-framework/logs"
	"github.com/iogo-framework/router"
)

func OAuthComponent(r *http.Request) *osin.Server {
	return router.Context(r).Env["Application"].(*application.Application).Components["OAuth"].(*osin.Server)
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
		logs.Error(resp.InternalError.Error())
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
				ar.UserData = ar.Username
			}
		}
		server.FinishAccessRequest(resp, r, ar)
	}
	if resp.IsError && resp.InternalError != nil {
		logs.Error(resp.InternalError.Error())
	}
	osin.OutputJSON(resp, w, r)
}

func parseBearerToken(auth string) (string, error) {
	if !strings.HasPrefix(auth, "Bearer ") {
		return "", errors.New("Not a bearer authorization header")
	}
	return strings.TrimPrefix(auth, "Bearer "), nil
}

func Test(w http.ResponseWriter, r *http.Request) {
	var token string
	var err error

	if token, err = parseBearerToken(r.Header.Get("Authorization")); err != nil {
		return
	}

	server := OAuthComponent(r)
	access, err := server.Storage.LoadAccess(token)
	if err != nil {
		return
	}
	io.WriteString(w, "Hello "+access.Client.GetId())
}

func Info(w http.ResponseWriter, r *http.Request) {
	server := OAuthComponent(r)
	resp := server.NewResponse()
	defer resp.Close()

	if ir := server.HandleInfoRequest(resp, r); ir != nil {
		// don't process if is already an error
		if resp.IsError {
			return
		}

		// output data
		resp.Output["client_id"] = ir.AccessData.Client.GetId()
		// resp.Output["access_token"] = ir.AccessData.AccessToken
		resp.Output["token_type"] = server.Config.TokenType
		resp.Output["expires_in"] = ir.AccessData.CreatedAt.Add(time.Duration(ir.AccessData.ExpiresIn)*time.Second).Sub(server.Now()) / time.Second
		if ir.AccessData.RefreshToken != "" {
			resp.Output["refresh_token"] = ir.AccessData.RefreshToken
		}
		if ir.AccessData.Scope != "" {
			resp.Output["scope"] = ir.AccessData.Scope
		}
		if ir.AccessData.UserData != nil {
			resp.Output["owner"] = ir.AccessData.UserData.(string)
		}
	}
	osin.OutputJSON(resp, w, r)
}
