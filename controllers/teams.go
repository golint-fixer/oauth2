// Bundle of functions managing the CRUD
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

// RetrieveTeamCollection calls the TeamSQL Find method and returns the results
func RetrieveTeamCollection(w http.ResponseWriter, r *http.Request) {
	var (
		err       error
		teams     []models.Team
		db        = getDB(r)
		teamStore = models.TeamStore(db)
	)
	if teams, err = teamStore.Find(); err != nil {
		logs.Error(err)
		Error(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	Success(w, r, views.Teams{Teams: teams}, http.StatusOK)
}

// RetrieveTeam calls the TeamSQL First method and returns the results
func RetrieveTeam(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(router.Context(r).Param("id"))
	if err != nil {
		logs.Debug(err)
		Fail(w, r, map[string]interface{}{"id": "not integer"}, http.StatusBadRequest)
		return
	}

	var (
		g         = models.Team{}
		db        = getDB(r)
		teamStore = models.TeamStore(db)
	)
	if err = teamStore.First(&g, uint(id)); err != nil {
		if err == sql.ErrNoRows {
			Fail(w, r, nil, http.StatusNotFound)
			return
		}
		logs.Error(err)
		Error(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	Success(w, r, views.Team{Team: &g}, http.StatusOK)
}

// RetrieveTeam calls the TeamSQL First method and returns the results
func RetrieveTeamByGroupID(w http.ResponseWriter, r *http.Request) {
	Team_id := router.Context(r).Param("groupID")
	var (
		g         = models.Team{}
		db        = getDB(r)
		teamStore = models.TeamStore(db)
	)
	if err := teamStore.FirstByGroup(&g, Team_id); err != nil {

		if (err == sql.ErrNoRows) || (err.Error() == "record not found") {
			logs.Info("Teame non trouvé")
			Fail(w, r, nil, http.StatusNotFound)
			return
		}
		logs.Error(err)
		Error(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	Success(w, r, views.Team{Team: &g}, http.StatusOK)
}

// UpdateTeam calls the TeamSQL Save method and returns the results
func UpdateTeam(w http.ResponseWriter, r *http.Request) {
	var (
		teamID int
		err    error
	)
	if teamID, err = strconv.Atoi(router.Context(r).Param("id")); err != nil {
		logs.Debug(err)
		Fail(w, r, map[string]interface{}{"id": "not integer"}, http.StatusBadRequest)
		return
	}

	var (
		db        = getDB(r)
		teamStore = models.TeamStore(db)
		g         = &models.Team{ID: uint(teamID)}
	)
	if err = teamStore.First(g, g.ID); err != nil {
		Fail(w, r, map[string]interface{}{"team": err.Error()}, http.StatusBadRequest)
		return
	}

	if err = Request(&views.Team{Team: g}, r); err != nil {
		logs.Debug(err)
		Fail(w, r, map[string]interface{}{"team": err.Error()}, http.StatusBadRequest)
		return
	}

	if err = teamStore.Save(g); err != nil {
		logs.Error(err)
		Error(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	Success(w, r, views.Team{Team: g}, http.StatusOK)
}

// CreateTeam calls the TeamSQL Save method and returns the results
func CreateTeam(w http.ResponseWriter, r *http.Request) {
	var (
		g = new(models.Team)

		err error
	)
	if err = Request(&views.Team{Team: g}, r); err != nil {
		logs.Debug(err)
		Fail(w, r, map[string]interface{}{"team": err.Error()}, http.StatusBadRequest)
		return
	}

	var (
		db        = getDB(r)
		teamStore = models.TeamStore(db)
	)
	if err = teamStore.Save(g); err != nil {
		logs.Error(err)
		Error(w, r, err.Error(), http.StatusInternalServerError)
		return
	}

	Success(w, r, views.Team{Team: g}, http.StatusCreated)
}

// DeleteTeam calls the TeamSQL Delete method and returns the results
func DeleteTeam(w http.ResponseWriter, r *http.Request) {
	var (
		teamID int

		err error
	)
	if teamID, err = strconv.Atoi(router.Context(r).Param("id")); err != nil {
		logs.Debug(err)
		Fail(w, r, map[string]interface{}{"id": "not integer"}, http.StatusBadRequest)
		return
	}

	var (
		db        = getDB(r)
		teamStore = models.TeamStore(db)
		g         = &models.Team{ID: uint(teamID)}
	)
	if err = teamStore.Delete(g); err != nil {
		logs.Debug(err)
		Fail(w, r, map[string]interface{}{"id": "not integer"}, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	Success(w, r, nil, http.StatusNoContent)
}
