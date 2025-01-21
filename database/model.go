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
		if strings.HasPrefix(relativePath, ".") {
			relativePath = relativePath[1:]
		}

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
