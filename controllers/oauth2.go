package controllers

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/RangelReale/osin"
	"github.com/RangelReale/osin/example"
	"github.com/quorumsco/application"
	. "github.com/quorumsco/jsonapi"
	"github.com/quorumsco/logs"
	"github.com/quorumsco/oauth2/models"
	"github.com/quorumsco/router"
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

func Auth(username string, password string, r *http.Request) (int, error) {
	var (
		u         = models.User{Mail: &username, Password: sPtr(password)}
		db        = getDB(r)
		userStore = models.UserStore(db)
	)
	if err := userStore.First(&u); err != nil {
		logs.Error(err)
		return 0, err
	}
	if u.ID == 0 {
		return 0, errors.New("No such user")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*u.Password), []byte(password)); err != nil {
		return 0, errors.New("Wrong password")
	}
	return int(u.GroupID), nil
}

func checkUser(username string, password string, r *http.Request) (int, error) {
	groupID, err := Auth(username, password, r)
	if err != nil || groupID == 0 {
		return 0, err
	}
	var (
		g          = models.Group{ID: 0}
		db         = getDB(r)
		groupStore = models.GroupStore(db)
	)
	err = groupStore.First(&g, uint(groupID))
	if err != nil {
		return 0, err
	}
	if g.ID == 0 {
		return 0, errors.New("No such group")
	}
	return groupID, nil
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
				ar.UserData = "1"
			} else {
				groupID, err := checkUser(ar.Username, ar.Password, r)
				if err != nil {
					resp.IsError = true
					resp.InternalError = err
				} else {
					ar.Authorized = true
				}
				ar.UserData = strconv.Itoa(groupID)
			}
		}
		server.FinishAccessRequest(resp, r, ar)
	}
	if resp.IsError && resp.InternalError != nil {
		logs.Error(resp.InternalError.Error())
		Fail(w, r, map[string]interface{}{"User": resp.InternalError.Error()}, http.StatusBadRequest)
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
