package controllers

import (
	"net/http"
	"safestore/database"
	"safestore/utils"
	"strings"

	"gorm.io/gorm"
)

func GetController(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	// get the collection from the URL
	path := strings.TrimPrefix(r.URL.Path, "/database/")
	urlPaths := strings.Split(path, "/")

	var collection string = ""
	var id string = ""
	if len(urlPaths)%2 == 0 {
		id = urlPaths[len(urlPaths)-1]
		collection = strings.Join(urlPaths[:len(urlPaths)-1], ".")
	} else {
		id = ""
		collection = strings.Join(urlPaths, ".")
	}
	// get the collection data

	if len(urlPaths)%2 == 0 {
		parentRow, err := database.GetInterface(db, collection, id)
		if err != nil {
			utils.FormatHttpError(w, 500, err.Error(), "Error getting parent collection")
			return
		}
		utils.FormatHttpSuccess(w, parentRow)
	} else {
		rows, err := database.GetCollection(db, collection)
		if err != nil {
			utils.FormatHttpError(w, 500, err.Error(), "Error getting collection")
			return
		}
		utils.FormatHttpSuccess(w, rows)
	}
}
