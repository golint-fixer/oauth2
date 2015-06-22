package views

import (
	"encoding/json"
	"html/template"
)

type Request struct {
	Data json.RawMessage `json:"data"`
}

type Success struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

type Fail struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

type Error struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func Templates(funcMap *template.FuncMap) map[string]*template.Template {
	var T = make(map[string]*template.Template)

	return T
}
