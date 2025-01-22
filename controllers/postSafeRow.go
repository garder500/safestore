package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/garder500/safestore/database"
	"github.com/garder500/safestore/utils"
	"gorm.io/gorm"
)

func PostSafeRow(w http.ResponseWriter, r *http.Request, db *gorm.DB, path *string) {

	data := make(map[string]interface{})

	// we gather the body of the request
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var paths []map[string]interface{}
	utils.GeneratePaths(data, *path, &paths)

	err = database.InsertInSafeRow(db, &paths)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// we marshal data to json to return it
	formattedData := map[string]interface{}{
		"success": true,
		"data":    data,
		"path":    *path,
	}
	jsonData, err := json.Marshal(formattedData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(jsonData)
}
