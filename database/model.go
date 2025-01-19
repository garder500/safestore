package database

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
	"gorm.io/gorm"
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

func (s *SafeRow) GetChildren(gormDB *gorm.DB) ([]SafeRow, error) {
	var children []SafeRow
	if err := gormDB.Where("path <@ ? AND path != ?", s.Path, s.Path).Find(&children).Error; err != nil {
		return nil, err
	}
	return children, nil
}

func (s *SafeRow) FormatChildren(gormDB *gorm.DB) (map[string]interface{}, error) {
	children, err := s.GetChildren(gormDB)
	if err != nil {
		return nil, err
	}
	return formatChildrenRecursive(children, s)
}

/*
*

		 The final objective is to get an interface like that :
		 {
	    	"posts": {
	        	"chack": {
	            	"jack": 3
	        	},
	    	}
		}
*/
func formatChildrenRecursive(children []SafeRow, parent *SafeRow) (map[string]interface{}, error) {
	parentKey := parent.GetKeyFromPath()
	result := make(map[string]interface{})

	parentValue, err := parent.GetTheNonNullValue()
	if err != nil {
		return nil, err
	}

	childrenData := make(map[string]interface{})

	for _, child := range children {
		if strings.HasPrefix(string(child.Path), string(parent.Path)+".") && strings.Count(string(child.Path), ".") == strings.Count(string(parent.Path), ".")+1 {
			childKey := child.GetKeyFromPath()
			childData, err := formatChildrenRecursive(children, &child)
			if err != nil {
				return nil, err
			}
			if len(childData) > 0 {
				childrenData[childKey] = childData[childKey]
			} else {
				childValue, err := child.GetTheNonNullValue()
				if err != nil {
					return nil, err
				}
				if childValue != nil {
					childrenData[childKey] = childValue
				}
			}
		}
	}

	if len(childrenData) > 0 {
		result[parentKey] = childrenData
	} else if parentValue != nil {
		result[parentKey] = parentValue
	}

	return result, nil
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
