package candishared

import (
	"database/sql"
	"testing"
	"time"

	"github.com/golangid/candi/candihelper"
	"github.com/stretchr/testify/assert"
)

func TestDBUpdateToolsSQL(t *testing.T) {
	type SubModel struct {
		Title       string       `gorm:"column:title" json:"title"`
		Profile     string       `gorm:"column:profile" json:"profile"`
		ActivatedAt sql.NullTime `gorm:"column:activated_at" json:"activatedAt"`
		CityAddress string       `gorm:"type:text"`
	}
	type Model struct {
		ID       int     `gorm:"column:db_id;" json:"id"`
		Name     *string `gorm:"column:db_name;" json:"name"`
		Address  string  `gorm:"column:db_address" json:"address"`
		No       int
		IgnoreMe SubModel `gorm:"column:test" json:"ignoreMe" ignoreUpdate:"true"`
		Rel      SubModel `gorm:"foreignKey:ID" json:"rel"`
		Multi    []SubModel
		Map      map[int]SubModel
		PtrModel *SubModel
		SubModel
	}
	var updated map[string]any

	updated = DBUpdateTools{KeyExtractorFunc: DBUpdateGORMExtractorKey}.ToMap(
		&Model{ID: 1, Name: candihelper.ToStringPtr("01"), Address: "street", SubModel: SubModel{Title: "test", CityAddress: "Jakarta"}, Rel: SubModel{Title: "rel sub"}},
		DBUpdateSetUpdatedFields("ID", "Name", "Title", "CityAddress"),
	)
	assert.Equal(t, 4, len(updated))
	assert.Equal(t, 1, updated["db_id"])
	assert.Equal(t, "01", updated["db_name"])
	assert.Equal(t, "test", updated["title"])
	assert.Equal(t, "Jakarta", updated["city_address"])

	updated = DBUpdateTools{KeyExtractorFunc: DBUpdateGORMExtractorKey}.ToMap(
		Model{ID: 1, Name: candihelper.ToStringPtr("01"), Address: "street", SubModel: SubModel{Title: "test", ActivatedAt: sql.NullTime{Valid: true}}},
		DBUpdateSetIgnoredFields("ID", "Name", "Title"),
	)
	assert.Equal(t, 5, len(updated))
	assert.Equal(t, "street", updated["db_address"])

	updated = DBUpdateTools{KeyExtractorFunc: DBUpdateGORMExtractorKey}.ToMap(
		Model{
			No: 10, Rel: SubModel{Title: "001"},
			SubModel: SubModel{ActivatedAt: sql.NullTime{Valid: true, Time: time.Now()}, CityAddress: "Jakarta"},
			PtrModel: &SubModel{CityAddress: "New"},
		},
	)
	assert.Equal(t, 8, len(updated))
	assert.Equal(t, "Jakarta", updated["city_address"])
	assert.Equal(t, 10, updated["no"])
}

func TestDBUpdateToolsMongo(t *testing.T) {
	type SubModel struct {
		Title       string `bson:"title" json:"title"`
		Profile     string `bson:"profile" json:"profile"`
		CityAddress string `bson:"city_address"`
	}
	type Model struct {
		ID       int     `bson:"db_id" json:"id"`
		Name     *string `bson:"db_name" json:"name"`
		Address  string  `bson:"db_address" json:"address"`
		No       int
		IgnoreMe SubModel
		Rel      SubModel   `bson:"rel" json:"rel"`
		Multi    []SubModel `bson:"multi"`
		SubModel
	}

	var updated map[string]any

	updated = DBUpdateTools{KeyExtractorFunc: DBUpdateMongoExtractorKey}.ToMap(
		&Model{ID: 1, Name: candihelper.WrapPtr("01"), Address: "street", SubModel: SubModel{Title: "test", CityAddress: "Jakarta"}, Rel: SubModel{Title: "rel sub"}},
		DBUpdateSetUpdatedFields("ID", "Name", "Title", "CityAddress"),
	)
	assert.Equal(t, 4, len(updated))
	assert.Equal(t, 1, updated["db_id"])
	assert.Equal(t, "01", updated["db_name"])
	assert.Equal(t, "test", updated["title"])
	assert.Equal(t, "Jakarta", updated["city_address"])

	updated = DBUpdateTools{KeyExtractorFunc: DBUpdateMongoExtractorKey}.ToMap(
		&Model{
			ID: 1, Name: candihelper.WrapPtr("01"), Address: "street", No: 100,
			SubModel: SubModel{Title: "test", CityAddress: "Jakarta"},
			Rel:      SubModel{Title: "rel sub"},
			Multi:    make([]SubModel, 0),
		},
	)
	assert.Equal(t, 9, len(updated))
}
