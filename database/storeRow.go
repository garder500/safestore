package database

import (
	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"
)

type StoreRow struct {
	gorm.Model
	Collection   LTree             `gorm:"type:ltree,unique_index:idx_store_row_collection_id_path,column:path" json:"collection"`
	CollectionId string            `gorm:"unique_index:idx_store_row_collection_id_path" json:"collection_id"`
	Data         pgtype.JSONBCodec `gorm:"type:jsonb" json:"data"`
}

func (*StoreRow) TableName() string {
	return "store.store_rows"
}

func GetCollection(db *gorm.DB, collection string) ([]StoreRow, error) {
	var rows []StoreRow
	err := StartWith(collection, db).Find(&rows).Error
	return rows, err
}
