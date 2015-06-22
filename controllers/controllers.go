package controllers

import (
	"encoding/json"
	"html/template"
	"net/http"

	"../components/database"
	"../components/logs"
	"../views"
	"github.com/silverwyrda/iogo"
)

func getEnv(r *http.Request) (*database.DB, map[string]*template.Template) {
	return iogo.GetContext(r).Env["DB"].(*database.DB), iogo.GetContext(r).Env["Templates"].(map[string]*template.Template)
}

func Success(w http.ResponseWriter, req *http.Request, data interface{}, status int) {
	if data == nil {
		w.WriteHeader(status)
		return
	}

	b, err := json.Marshal(views.Success{Status: "success", Data: data})
	if err != nil {
		logs.Error(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(b)
}

func Fail(w http.ResponseWriter, req *http.Request, data interface{}, status int) {
	b, err := json.Marshal(views.Fail{Status: "fail", Data: data})
	if err != nil {
		logs.Error(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(b)
}

func Error(w http.ResponseWriter, req *http.Request, message string, status int) {
	b, err := json.Marshal(views.Error{Status: "success", Message: message})
	if err != nil {
		logs.Error(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(b)
}

func Request(data interface{}, req *http.Request) error {
	var r = new(views.Request)
	var err = json.NewDecoder(req.Body).Decode(&r)
	if err != nil {
		return err
	}

	err = json.Unmarshal(r.Data, data)
	if err != nil {
		return err
	}
	return nil
}
