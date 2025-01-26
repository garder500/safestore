package database

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SafeRow struct {
	Path             LTree           `gorm:"type:ltree;primaryKey;index:idx_path_gist,type:gist" json:"path"`
	Int              *int32          `gorm:"column:int_value" json:"int_value"`
	Text             *string         `gorm:"column:text_value" json:"text_value"`
	CollectionString pq.StringArray  `gorm:"type:text[];column:collection_string" json:"collection_string"`
	CollectionInt    pq.Int32Array   `gorm:"type:integer[];column:collection_int" json:"collection_int"`
	Timestamp        *time.Time      `gorm:"column:timestamp_value" json:"timestamp_value"`
	Boolean          *bool           `gorm:"column:boolean_value" json:"boolean_value"`
	Numeric          *pgtype.Numeric `gorm:"type:numeric(15,2);column:numeric_value" json:"numeric_value"`
	UUID             *uuid.UUID      `gorm:"type:uuid;column:uuid_value" json:"uuid_value"`
	BinaryData       []byte          `gorm:"type:bytea;column:binary_data" json:"binary_data"`
	GeoPoint         *pgtype.Point   `gorm:"type:point;column:geo_point" json:"geo_point"`
}

func (*SafeRow) TableName() string {
	return "realtime.safe_rows"
}

type LTree string

func (l *LTree) Scan(value interface{}) error {
	if value == nil {
		*l = ""
		return nil
	}
	*l = LTree(value.(string))
	return nil
}

func (l LTree) Value() (driver.Value, error) {
	return string(l), nil
}

func FormatChildrenRecursive(children []*SafeRow, startPath string) (map[string]interface{}, error) {
	results := make(map[string]interface{})

	for _, child := range children {
		childPath := string(child.Path)

		// Convert startPath from URL format to dot notation
		startPathDot := strings.ReplaceAll(startPath, "/", ".")
		if startPathDot != "" && !strings.HasPrefix(childPath, startPathDot) {
			continue
		}

		// Remove the startPath prefix
		relativePath := strings.TrimPrefix(childPath, startPathDot)
		relativePath = strings.TrimPrefix(relativePath, ".")

		pathParts := strings.Split(relativePath, ".")

		value, err := child.GetTheNonNullValue()
		if err != nil {
			return nil, fmt.Errorf("error getting value for path %s: %w", child.Path, err)
		}

		// Navigate or create nested maps
		current := results
		for i, part := range pathParts {
			if part == "" {
				continue
			}
			if i == len(pathParts)-1 {
				// Last part, set the value
				current[part] = value
			} else {
				// Create nested map if it doesn't exist
				if _, exists := current[part]; !exists {
					current[part] = make(map[string]interface{})
				}
				nextMap, ok := current[part].(map[string]interface{})
				if !ok {
					// If it's not a map, create a new one
					nextMap = make(map[string]interface{})
					current[part] = nextMap
				}
				current = nextMap
			}
		}
	}

	return results, nil
}

func (s *SafeRow) ToJson() (string, error) {
	dataInterface := map[string]interface{}{}
	uniqueValue, err := s.GetTheNonNullValue()
	if err != nil {
		return "", err
	}
	dataInterface[s.GetKeyFromPath()] = uniqueValue
	data, err := json.Marshal(dataInterface)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *SafeRow) GetKeyFromPath() string {
	// if the path is xxx.yyyy.zzzz then the key is zzzz
	splittedPath := strings.Split(string(s.Path), ".")
	return splittedPath[len(splittedPath)-1]
}

func (s *SafeRow) GetTheNonNullValue() (interface{}, error) {
	if s.Int != nil {
		return *s.Int, nil
	} else if s.Text != nil {
		return *s.Text, nil
	} else if len(s.CollectionString) > 0 {
		return s.CollectionString, nil
	} else if len(s.CollectionInt) > 0 {
		return s.CollectionInt, nil
	} else if s.Timestamp != nil {
		return *s.Timestamp, nil
	} else if s.Boolean != nil {
		return *s.Boolean, nil
	} else if s.Numeric != nil {
		return s.Numeric, nil
	} else if s.UUID != nil {
		return *s.UUID, nil
	} else if len(s.BinaryData) > 0 {
		return s.BinaryData, nil
	} else if s.GeoPoint != nil {
		return s.GeoPoint, nil
	}
	return nil, nil
}

func InsertInSafeRow(db *gorm.DB, values *[]map[string]interface{}) error {
	rows := make([]SafeRow, 0)

	for _, value := range *values {
		// we need to get the type of the value
		var safeRow SafeRow = SafeRow{
			Path: LTree(value["path"].(string)),
		}
		for key, val := range value {
			if key == "path" {
				continue
			}
			switch val := val.(type) {
			case int:
				intValue := int32(val)
				safeRow.Int = &intValue
			case float64:
				intValue := int32(val)
				safeRow.Int = &intValue
			case string:
				textValue := val
				safeRow.Text = &textValue
			case time.Time:
				timestampValue := val
				safeRow.Timestamp = &timestampValue
			case bool:
				booleanValue := val
				safeRow.Boolean = &booleanValue
			case *pgtype.Numeric:
				safeRow.Numeric = val
			case uuid.UUID:
				uuidValue := val
				safeRow.UUID = &uuidValue
			case []byte:
				binaryData := val
				safeRow.BinaryData = binaryData
			case *pgtype.Point:
				safeRow.GeoPoint = val
			}

			// it's possible that the value is an array of values and we need to check the type of the first element
			// to know the type of the array
			switch val := val.(type) {
			case []interface{}:
				if len(val) == 0 {
					continue
				}
				switch val[0].(type) {
				case string:
					collectionString := make([]string, 0)
					for _, v := range val {
						collectionString = append(collectionString, v.(string))
					}
					safeRow.CollectionString = collectionString
				case int:
					collectionInt := make([]int32, 0)
					for _, v := range val {
						collectionInt = append(collectionInt, int32(v.(int)))
					}
					safeRow.CollectionInt = collectionInt
				case float64:
					collectionInt := make([]int32, 0)
					for _, v := range val {
						collectionInt = append(collectionInt, int32(v.(float64)))
					}
					safeRow.CollectionInt = collectionInt
				}
			}
		}
		rows = append(rows, safeRow)
	}

	// we need to remove bottom rows if they are already in the database
	// we remove any path that start with each path
	for _, row := range rows {
		err := StartWith(string(row.Path), db).Delete(&SafeRow{}).Error
		if err != nil {
			return err
		}
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "path"}},
		DoUpdates: clause.AssignmentColumns([]string{"int_value", "text_value", "collection_string", "collection_int", "timestamp_value", "boolean_value", "numeric_value", "uuid_value", "binary_data", "geo_point"}),
	}).Create(&rows).Error

}

func DeleteInSafeRow(db *gorm.DB, path *string) error {
	rows := make([]*SafeRow, 0)
	if *path == "" {
		// we need to delete all the rows
		err := db.Find(&rows).Error
		if err != nil {
			return err
		}
		for _, row := range rows {
			err := db.Delete(row).Error
			if err != nil {
				return err
			}
		}
	}

	return StartWith(*path, db).Delete(&rows).Error

}
