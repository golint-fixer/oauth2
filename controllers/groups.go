package controllers

import (
	"database/sql"
	"net/http"
	"strconv"

	. "github.com/quorumsco/jsonapi"
	"github.com/quorumsco/logs"
	"github.com/quorumsco/oauth2/models"
	"github.com/quorumsco/oauth2/views"
	"github.com/quorumsco/router"
)

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

func RetrieveGroup(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(router.Context(r).Param("id"))
	if err != nil {
		logs.Debug(err)
		Fail(w, r, map[string]interface{}{"id": "not integer"}, http.StatusBadRequest)
		return
	}

	var (
		g          = models.Group{ID: uint(id)}
		db         = getDB(r)
		groupStore = models.GroupStore(db)
	)
	if err = groupStore.First(&g); err != nil {
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
	if err = groupStore.First(g); err != nil {
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

func CreateGroup(w http.ResponseWriter, r *http.Request) {
	var (
		g = new(models.Group)

		err error
	)
	if err = Request(&views.Group{Group: g}, r); err != nil {
		logs.Debug(err)
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
