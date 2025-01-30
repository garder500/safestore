package controllers

import (
	"encoding/json"
	"net/http"
	"safestore/database"
	"safestore/utils"
	"strings"

	"gorm.io/gorm"
)

func PostController(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	// get the collection from the URL
	path := strings.TrimPrefix(r.URL.Path, "/database/")
	urlPaths := strings.Split(path, "/")
	// get the ID from the URL
	id := urlPaths[len(urlPaths)-1]
	// get the collection from the URL
	collection := strings.Join(urlPaths[:len(urlPaths)-1], ".")

	// parse body to get the data
	var data map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&data)

	if err != nil {
		utils.FormatHttpError(w, 500, err.Error(), "Error parsing body")
		return
	}

	// update the current collection or override it

	err = database.UpdateOrCreateInterface(db, collection, id, data)
	if err != nil {
		utils.FormatHttpError(w, 500, err.Error(), "Error updating or creating interface")
		return
	}
	utils.FormatHttpSuccess(w, map[string]interface{}{"id": id, "collection": collection, "data": data})
}
