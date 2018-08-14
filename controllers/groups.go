// Bundle of functions managing the CRUD
package controllers

import (
	"database/sql"
	//"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/schema"
	. "github.com/quorumsco/jsonapi"
	"github.com/quorumsco/logs"
	"github.com/quorumsco/oauth2/models"
	"github.com/quorumsco/oauth2/views"
	"github.com/quorumsco/router"
)

// RetrieveGroupCollection calls the GroupSQL Find method and returns the results
func RetrieveGroupCollection(w http.ResponseWriter, r *http.Request) {
	var (
		err        error
		groups     []models.Group
		db         = getDB(r)
		groupStore = models.GroupStore(db)
	)
	if groups, err = groupStore.Find(); err != nil {
		logs.Error(err)
		Error(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	Success(w, r, views.Groups{Groups: groups}, http.StatusOK)
}

// RetrieveGroup calls the GroupSQL First method and returns the results
func RetrieveGroup(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(router.Context(r).Param("id"))
	if err != nil {
		logs.Debug(err)
		Fail(w, r, map[string]interface{}{"id": "not integer"}, http.StatusBadRequest)
		return
	}

	var (
		g          = models.Group{}
		db         = getDB(r)
		groupStore = models.GroupStore(db)
	)
	if err = groupStore.First(&g, uint(id)); err != nil {
		if err == sql.ErrNoRows {
			Fail(w, r, nil, http.StatusNotFound)
			return
		}
		logs.Error(err)
		Error(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	Success(w, r, views.Group{Group: &g}, http.StatusOK)
}

// RetrieveGroup calls the GroupSQL First method and returns the results
func RetrieveGroupByCode_cause(w http.ResponseWriter, r *http.Request) {
	Code_cause := router.Context(r).Param("cause")
	var (
		g          = models.Group{}
		db         = getDB(r)
		groupStore = models.GroupStore(db)
	)
	if err := groupStore.FirstByCodeCause(&g, Code_cause); err != nil {

		if (err == sql.ErrNoRows) || (err.Error() == "record not found") {
			logs.Info("Groupe non trouvé")
			Fail(w, r, nil, http.StatusNotFound)
			return
		}
		logs.Error(err)
		Error(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	Success(w, r, views.Group{Group: &g}, http.StatusOK)
}

// UpdateGroup calls the GroupSQL Save method and returns the results
func UpdateGroup(w http.ResponseWriter, r *http.Request) {
	var (
		groupID int
		err     error
	)
	if groupID, err = strconv.Atoi(router.Context(r).Param("id")); err != nil {
		logs.Debug(err)
		Fail(w, r, map[string]interface{}{"id": "not integer"}, http.StatusBadRequest)
		return
	}

	var (
		db         = getDB(r)
		groupStore = models.GroupStore(db)
		g          = &models.Group{ID: uint(groupID)}
	)
	if err = groupStore.First(g, g.ID); err != nil {
		Fail(w, r, map[string]interface{}{"group": err.Error()}, http.StatusBadRequest)
		return
	}

	if err = Request(&views.Group{Group: g}, r); err != nil {
		logs.Debug(err)
		Fail(w, r, map[string]interface{}{"group": err.Error()}, http.StatusBadRequest)
		return
	}

	if err = groupStore.Save(g); err != nil {
		logs.Error(err)
		Error(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	Success(w, r, views.Group{Group: g}, http.StatusOK)
}

// CreateGroup calls the GroupSQL Save method and returns the results
func CreateGroup(w http.ResponseWriter, r *http.Request) {
	var (
		g   = new(models.Group)
		err error
	)

	err = r.ParseForm()
	if err != nil {
		logs.Error(err)
		Fail(w, r, map[string]interface{}{"group": err.Error()}, http.StatusBadRequest)
		return
	}

	decoder := schema.NewDecoder()
	err = decoder.Decode(g, r.PostForm)
	if err != nil {
		logs.Error(err)
		Fail(w, r, map[string]interface{}{"group": err.Error()}, http.StatusBadRequest)
		return
	}

	var (
		db         = getDB(r)
		groupStore = models.GroupStore(db)
	)

	if err = groupStore.Save(g); err != nil {
		logs.Error(err)
		Error(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	Success(w, r, views.Group{Group: g}, http.StatusCreated)
}

// DeleteGroup calls the GroupSQL Delete method and returns the results
func DeleteGroup(w http.ResponseWriter, r *http.Request) {
	var (
		groupID int

		err error
	)
	if groupID, err = strconv.Atoi(router.Context(r).Param("id")); err != nil {
		logs.Debug(err)
		Fail(w, r, map[string]interface{}{"id": "not integer"}, http.StatusBadRequest)
		return
	}

	var (
		db         = getDB(r)
		groupStore = models.GroupStore(db)
		g          = &models.Group{ID: uint(groupID)}
	)
	if err = groupStore.Delete(g); err != nil {
		logs.Debug(err)
		Fail(w, r, map[string]interface{}{"id": "not integer"}, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	Success(w, r, nil, http.StatusNoContent)
}
