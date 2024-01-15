package main

import (
	"documentapi/pkg/api"
	"documentapi/pkg/database"
)

type DocumentCommentService struct {
	SQL *database.SQLite
	API *api.API
}

func main() {
	sqlService := &database.SQLite{}
	apiService := &api.API{}

	sqlService.Initialize()
	apiService.Initialize(sqlService)

	d := DocumentCommentService{
		SQL: sqlService,
		API: apiService,
	}

	d.API.StartAPI()
}
