package database

import (
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type StoreRow struct {
	Collection   LTree  `gorm:"column:path;type:ltree;index:idx_full_path,unique" json:"collection"`
	CollectionId string `gorm:"index:idx_full_path,unique" json:"collection_id"`
	Data         []byte `gorm:"column:data;type:jsonb" json:"data"`
}

func (*StoreRow) TableName() string {
	return "store.store_rows"
}

func GetCollection(db *gorm.DB, collection string) (map[string]interface{}, error) {
	var rows []StoreRow
	err := db.Where("path = ?", collection).Find(&rows).Error
	// return only decoded data

	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{}, 0)
	for _, row := range rows {
		// data is in base64, decode it
		var decodedData map[string]interface{}
		err = json.Unmarshal(row.Data, &decodedData)
		if err != nil {
			return nil, err
		}

		data[row.CollectionId] = decodedData
	}

	return data, nil
}

func MergeInterface(a, b map[string]interface{}) map[string]interface{} {
	for k, v := range b {
		// if the value is a map, merge it recursively
		if _, ok := v.(map[string]interface{}); ok {
			if _, ok := a[k]; !ok {
				a[k] = make(map[string]interface{})
			}
			a[k] = MergeInterface(a[k].(map[string]interface{}), v.(map[string]interface{}))
		} else {
			a[k] = v
		}
	}
	return a
}

func UpdateInterface(db *gorm.DB, collection string, id string, data map[string]interface{}) error {
	// encode JSON data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// update the row
	err = db.Model(&StoreRow{}).Where("path = ?", collection).Where("collection_id = ?", id).Update("data", jsonData).Error
	return err
}

func CreateInterface(db *gorm.DB, collection string, id string, data map[string]interface{}) error {
	// encode JSON data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// create the row
	err = db.Create(&StoreRow{
		Collection:   LTree(collection),
		CollectionId: id,
		Data:         jsonData,
	}).Error
	return err
}

func DeleteInterface(db *gorm.DB, collection string, id string) error {
	// delete the row
	err := db.Where("path = ?", collection).Where("collection_id = ?", id).Delete(&StoreRow{}).Error
	return err
}

func GetInterface(db *gorm.DB, collection string, id string) (map[string]interface{}, error) {
	// get the row
	var row StoreRow
	err := db.Where("path = ?", collection).Where("collection_id = ?", id).First(&row).Error
	if err != nil {
		return nil, err
	}

	// decode JSON data
	var data map[string]interface{}
	err = json.Unmarshal(row.Data, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func GetChildCollections(db *gorm.DB, collection string) ([]string, error) {
	// get the rows
	var rows []StoreRow
	err := db.Where("path ~ ?", collection+".*").Find(&rows).Error
	if err != nil {
		return nil, err
	}

	// extract the child collections
	collections := make([]string, 0)
	for _, row := range rows {
		collections = append(collections, string(row.Collection))
	}

	return collections, nil
}

func UpdateOrCreateInterface(db *gorm.DB, collection string, id string, data map[string]interface{}) error {
	// Update the row if it exists

	jsonData, err := json.Marshal(data)

	if err != nil {
		return err
	}

	return db.Model(&StoreRow{}).Where("path = ?", collection).Where("collection_id = ?", id).FirstOrCreate(&StoreRow{
		Collection:   LTree(collection),
		CollectionId: id,
		Data:         jsonData,
	}).Error
}

type FilterSearch struct {
	Path       string `json:"path"`
	SearchType string `json:"searchType"`
	Value      string `json:"value"`
}

func SearchUsingJsonBPath(db *gorm.DB, collection string, filters []FilterSearch) ([]StoreRow, error) {
	var rows []StoreRow
	var err error
	var finalQuery *gorm.DB = db.Where("path = ?", collection)
	for _, filter := range filters {
		switch filter.SearchType {
		case "contains":
			finalQuery = db.Where("jsonb_path_exists(data, $1)", fmt.Sprintf("$[*] ? (@ like_regex \"%s\")", strings.ReplaceAll(filter.Value, "*", ".*")))
		case "equals":
			finalQuery = db.Where("jsonb_path_exists(data, $1)", fmt.Sprintf("$[*] ? (@ == \"%s\")", filter.Value))
		case "notEquals":
			finalQuery = db.Not("jsonb_path_exists(data, $1)", fmt.Sprintf("$[*] ? (@ == \"%s\")", filter.Value))
		case "startWith":
			finalQuery = db.Where("jsonb_path_exists(data, $1)", fmt.Sprintf("$[*] ? (@ like_regex \"%s.*\")", filter.Value))
		case "endWith":
			finalQuery = db.Where("jsonb_path_exists(data, $1)", fmt.Sprintf("$[*] ? (@ like_regex \".*%s\")", filter.Value))
		case "startAndEndWith":
			finalQuery = db.Where("jsonb_path_exists(data, $1)", fmt.Sprintf("$[*] ? (@ like_regex \"%s.*%s\")", filter.Path, filter.Value))
		}
	}

	err = finalQuery.Find(&rows).Error
	return rows, err
}
