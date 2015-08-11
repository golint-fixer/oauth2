package controllers

import (
	"bytes"
	"encoding/json"
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
	. "github.com/quorumsco/jsonapi"
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

func request(method string, urlstr string, r *http.Request) ([]byte, error) {
	logs.Debug(method + " " + urlstr)

	client := &http.Client{}

	var tmp = make(map[string]string)
	tmp["username"] = router.Context(r).Env["Username"].(string)
	tmp["password"] = router.Context(r).Env["Password"].(string)
	jsonBody, err := json.Marshal(tmp)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, urlstr, bytes.NewBuffer(jsonBody))

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

	return request(method, urlstr, r)
}

func checkUser(w http.ResponseWriter, r *http.Request) (uint, error) {
	body, err := Service("POST", r, "Users", "/users/auth")
	if err != nil {
		logs.Error(err)
		return 0, err
	}
	infos := make(map[string]interface{})
	if err := json.Unmarshal(body, &infos); err != nil {
		logs.Error(err)
		Fail(w, r, map[string]string{"Authentification": "Error"}, http.StatusBadRequest)
		return 0, err
	}
	groupID := infos["group_id"]
	if groupID == nil {
		return 0, nil
	}
	return groupID.(uint), nil
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
				router.Context(r).Env["Username"] = ar.Username
				router.Context(r).Env["Password"] = ar.Password
				groupID, err := checkUser(w, r)
				if err != nil || groupID == 0 {
					resp.IsError = true
					if err == nil {
						resp.InternalError = errors.New("Wrong username or password")
					} else {
						resp.InternalError = err
					}
				}
				ar.UserData = groupID
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
