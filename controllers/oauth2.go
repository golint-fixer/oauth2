// Bundle of functions managing the CRUD
package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/RangelReale/osin"
	"github.com/RangelReale/osin/example"
	manager "github.com/ory-am/ladon/manager/memory"
	"github.com/ory/ladon"
	"github.com/quorumsco/application"
	. "github.com/quorumsco/jsonapi"
	"github.com/quorumsco/logs"
	"github.com/quorumsco/oauth2/models"
	"github.com/quorumsco/router"
)

// OAuthComponents returns the OAuth client defined in the main
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

// Returns the user infos from the database using the credentials
func getUserInfos(username string, password string, origin string, r *http.Request) (models.UserInfos, error, int) {
	var (
		u         = models.User{Mail: &username, Password: sPtr(password)}
		db        = getDB(r)
		userStore = models.UserStore(db)
	)
	if err := userStore.First(&u); err != nil {
		logs.Error(err)
		return models.UserInfos{}, err, http.StatusInternalServerError
	}
	if u.ID == 0 {
		return models.UserInfos{}, errors.New("no such user"), 404
	}
	//codeTMP := u.Validationcode

	if err := bcrypt.CompareHashAndPassword([]byte(*u.Password), []byte(password)); err != nil {
		return models.UserInfos{}, errors.New("wrong password"), http.StatusBadRequest
	}

	if u.Validationcode != nil {
		codeTMP := *u.Validationcode
		if codeTMP != "" {
			return models.UserInfos{}, errors.New("no email validation"), http.StatusPreconditionFailed
		}
	}
	// if u.GroupID == 0 {
	// 	return "0", errors.New("User not activate")
	// }

	role := u.Role
	code := u.Validationcode

	if u.Validationcode != nil && *code != "" {
		if u.GroupID == 0 {
			return models.UserInfos{}, errors.New("User don't have a cause"), http.StatusUpgradeRequired
		} else {
			return models.UserInfos{}, errors.New("Very strange... - contact support"), http.StatusUnauthorized
		}
	} else if u.GroupID == 0 {
		return models.UserInfos{}, errors.New("User don't have a cause"), http.StatusUpgradeRequired
	}

	//----- CONTROLE ACCES WEBAPP -----------//

	logs.Debug("request origin:" + origin)
	if origin == "https://test.quorumapps.com" || origin == "test.quorumapps.com" || origin == "https://cloud.quorumapps.com" || origin == "cloud.quorumapps.com" || origin == "http://localhost:8101" {
		if u.Role == nil || *role == "" {
			logs.Debug("can't access")
			//logs.Debug(*u.Role)
			return models.UserInfos{}, errors.New("User don't have the permission to access the webapp"), http.StatusForbidden
		} else if *u.Role != "admin" {
			logs.Debug("can't access")
			//logs.Debug(*u.Role)
			return models.UserInfos{}, errors.New("User don't have the permission to access the webapp"), http.StatusForbidden
		}
		logs.Debug("can access, role:")
		logs.Debug(*u.Role)
	}

	//----- END CONTROLE ACCES WEBAPP -----------//

	userInfos := models.UserInfos{
		models.UserLight{u.ID},
		u.GroupID,
		*u.Role,
	}

	return userInfos, nil, http.StatusFound
}

// Token endpoint
func Token(w http.ResponseWriter, r *http.Request) {
	logs.Debug(r.FormValue("Origin"))
	var (
		server     = OAuthComponent(r)
		resp       = server.NewResponse()
		codeErreur = http.StatusBadRequest
	)
	defer resp.Close()
	if ar := server.HandleAccessRequest(resp, r); ar != nil {
		switch ar.Type {
		case osin.AUTHORIZATION_CODE:
			ar.Authorized = true
		case osin.REFRESH_TOKEN:
			ar.Authorized = true
		case osin.PASSWORD:
			userInfos, err, code := getUserInfos(ar.Username, ar.Password, r.FormValue("Origin"), r)
			if err != nil {
				resp.IsError = true
				resp.InternalError = err
				resp.ErrorStatusCode = code
				codeErreur = code
			} else {
				ar.Authorized = true
			}

			payload, err := json.Marshal(userInfos)
			if err != nil {
				logs.Error("Cannot marshal user datas to Json: ", err)
				resp.IsError = true
				resp.InternalError = errors.New(http.StatusText(http.StatusInternalServerError))
				resp.ErrorStatusCode = http.StatusInternalServerError
			}

			ar.UserData = string(payload)

		}
		server.FinishAccessRequest(resp, r, ar)
	}
	if resp.IsError && resp.InternalError != nil {
		Fail(w, r, map[string]interface{}{"message": resp.InternalError.Error(), "code": codeErreur}, codeErreur)
		return
	}
	osin.OutputJSON(resp, w, r)
}

// Extracts the token from the header and returns it
func parseBearerToken(auth string) (string, error) {
	if !strings.HasPrefix(auth, "Bearer ") {
		return "", errors.New("Not a bearer authorization header")
	}
	return strings.TrimPrefix(auth, "Bearer "), nil
}

// Info return the token's information via http
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
		//--------------------------------------LADON ------------------------------------
		var pol = &ladon.DefaultPolicy{
			ID:          "1",
			Description: "This policy allows max to update any resource",
			Subjects:    []string{"max"},
			Actions:     []string{"delete"},
			Resources:   []string{"<.*>"},
			Effect:      ladon.AllowAccess,
			Conditions: ladon.Conditions{
				"clientIP": &ladon.CIDRCondition{
					//CIDR: "1.1.1.1/32",
					//CIDR: "0.0.0.0/0",
					CIDR: "0.0.0.0/1",
					//CIDR: "127.0.0.1/32",
				},
			},
		}
		// db := redis.NewClient(&redis.Options{
		//     Addr:     "localhost:6379",
		// })
		//
		// if err := db.Ping().Err(); err != nil {
		//     logs.Error("Could not connect to database")
		// }

		warden := &ladon.Ladon{
			//Manager: ladon.NewMemoryManager(),
			Manager: manager.NewMemoryManager(),

			//Manager: ladon.NewRedisManager(db, "redis_key_prefix:"),
		}
		err := warden.Manager.Create(pol)
		if err != nil {
			logs.Error("err Create(pol):")
			logs.Error(err)
			return
		}

		err2 := warden.IsAllowed(&ladon.Request{
			Subject:  "max",
			Action:   "delete",
			Resource: "myrn:some.domain.com:resource:123",
			Context: ladon.Context{
				"clientIP": "127.0.0.1",
			},
		})
		if err2 != nil {
			logs.Error("Access denied")
			logs.Error(err2)
			return
		}

		//--------------------------------------FIN LADON ------------------------------------
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
			var payload models.UserInfos

			if err := json.Unmarshal([]byte(ir.AccessData.UserData.(string)), &payload); err == nil {
				resp.Output["owner"] = payload
			}
		}
		server.FinishInfoRequest(resp, r, ir)
	}
	//Right here retry with the session.
	osin.OutputJSON(resp, w, r)
}
