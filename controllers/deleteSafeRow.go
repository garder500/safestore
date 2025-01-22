package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/garder500/safestore/database"
	"gorm.io/gorm"
)

func DeleteSafeRow(w http.ResponseWriter, r *http.Request, db *gorm.DB, path *string) {
	// we gather the path from the request
	rows := make([]*database.SafeRow, 0)
	if *path == "" {
		// we need to delete all the rows
		err := db.Find(&rows).Error
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, row := range rows {
			err := db.Delete(row).Error
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		data, err := database.FormatChildrenRecursive(rows, *path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// we marshal data to json to return it
		formattedData := map[string]interface{}{
			"success": true,
			"path":    path,
			"data":    data,
		}
		jsonData, err := json.Marshal(formattedData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(jsonData)
		return
	}
	err := database.StartWith(*path, db).Delete(&rows).Error

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data, err := database.FormatChildrenRecursive(rows, *path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// we marshal data to json to return it
	formattedData := map[string]interface{}{
		"success": true,
		"path":    path,
		"data":    data,
	}
	jsonData, err := json.Marshal(formattedData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(jsonData)
}
