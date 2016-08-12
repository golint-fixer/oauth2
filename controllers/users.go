// Bundle of functions managing the CRUD
package controllers

import (
	"database/sql"
	"net/http"
	"strconv"

	"golang.org/x/crypto/bcrypt"

	. "github.com/quorumsco/jsonapi"
	"github.com/quorumsco/logs"
	"github.com/quorumsco/oauth2/models"
	"github.com/quorumsco/oauth2/views"
	"github.com/quorumsco/router"
)

// Return a string's pointer
func sPtr(s string) *string {
	if s == "" {
		return nil
	} else {
		return &s
	}
}

// Creates a new user
func Register(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		req.ParseForm()

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.FormValue("password")), bcrypt.DefaultCost)
		if err != nil {
			panic(err)
		}

		u := &models.User{
			Firstname: sPtr(req.FormValue("firstname")),
			Surname:   sPtr(req.FormValue("surname")),
			Mail:      sPtr(req.FormValue("mail")),
			Password:  sPtr(string(passwordHash)),
		}

		errs := u.Validate()
		if len(errs) > 0 {
			logs.Error(errs)
			Error(w, req, "Vous avez une ou des erreur(s) dans le formulaire d'inscription. vérifiez votre saisie (formatage du mail par exemple)", http.StatusBadRequest)
			//Fail(w, req, "", http.StatusInternalServerError)
			return
		}

		var store = models.UserStore(getDB(req))
		err = store.Save(u)
		if err != nil {
			logs.Error(err)
			Error(w, req, err.Error(), http.StatusBadRequest)
			return
		}
	}

	templates := getTemplates(req)
	if err := templates["users/register"].ExecuteTemplate(w, "base", nil); err != nil {
		logs.Error(err)
	}
}

// Returns a user
func RetrieveUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(router.Context(r).Param("id"))
	if err != nil {
		logs.Debug(err)
		Fail(w, r, map[string]interface{}{"id": "not integer"}, http.StatusBadRequest)
		return
	}

	var (
		u         = models.User{ID: int64(id)}
		db        = getDB(r)
		userStore = models.UserStore(db)
	)
	if err = userStore.First(&u); err != nil {
		if err == sql.ErrNoRows {
			Fail(w, r, nil, http.StatusNotFound)
			return
		}
		logs.Error(err)
		Error(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	Success(w, r, views.User{User: &u}, http.StatusOK)
}
