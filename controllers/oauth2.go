package controllers

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/RangelReale/osin"
	"github.com/RangelReale/osin/example"
	"github.com/quorumsco/application"
	"github.com/quorumsco/logs"
	"github.com/quorumsco/router"
	"github.com/quorumsco/settings"
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
		ar.UserData = "test" // Get user_id
		ar.Authorized = true
		server.FinishAuthorizeRequest(resp, r, ar)
	}
	if resp.IsError && resp.InternalError != nil {
		logs.Error(resp.InternalError.Error())
	}
	osin.OutputJSON(resp, w, r)
}

func request(method string, urlstr string, body io.ReadCloser) ([]byte, error) {
	logs.Debug(method + " " + urlstr)

	client := &http.Client{}
	req, err := http.NewRequest(method, urlstr, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func Service(method string, r *http.Request, service string, path string) ([]byte, error) {
	s := router.Context(r).Env["Application"].(*application.Application).Components[service].(settings.Server)
	urlstr := fmt.Sprintf("http://%s:%d%s", s.Host, s.Port, path)

	fmt.Println("TEST : ")
	fmt.Println(urlstr)
	return request(method, urlstr, r.Body)
}

func checkUser(ar *osin.AccessRequest, w http.ResponseWriter, r *http.Request) error {
	body, err := Service("GET", r, "Users", "/auth")
	if err != nil {
		logs.Error(err)
		return err
	}

	fmt.Println(body)
	return nil
}

// Token endpoint
func Token(w http.ResponseWriter, r *http.Request) {
	var (
		server = OAuthComponent(r)
		resp   = server.NewResponse()
	)
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
			} else {
				checkUser(ar, w, r)
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
	var (
		token string
		err   error
	)

	if token, err = parseBearerToken(r.Header.Get("Authorization")); err != nil {
		return
	}

	var (
		server = OAuthComponent(r)
		access *osin.AccessData
	)
	if access, err = server.Storage.LoadAccess(token); err != nil {
		return
	}
	io.WriteString(w, "Hello "+access.Client.GetId())
}

func Info(w http.ResponseWriter, r *http.Request) {
	var (
		server = OAuthComponent(r)
		resp   = server.NewResponse()
	)
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
