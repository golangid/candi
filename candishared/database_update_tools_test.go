package candishared

import (
	"database/sql"
	"testing"
	"time"

	"github.com/golangid/candi/candihelper"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestDBUpdateToolsSQL(t *testing.T) {
	type SubModel struct {
		Title       string       `gorm:"column:title" json:"title"`
		Profile     string       `gorm:"column:profile;default:null" json:"profile"`
		ActivatedAt sql.NullTime `gorm:"column:activated_at" json:"activatedAt"`
		CityAddress string       `gorm:"type:text"`
	}
	type Model struct {
		ID        int     `gorm:"column:id;" json:"id"`
		Name      *string `gorm:"column:name;" json:"name"`
		Address   string  `gorm:"column:address" json:"address"`
		No        int
		CreatedAt time.Time
		IgnoreMe  SubModel `gorm:"column:test" json:"ignoreMe" ignoreUpdate:"true"`
		Rel       SubModel `gorm:"foreignKey:ID" json:"rel"`
		Log       []byte
		StrArr    pq.StringArray
		IntArr    pq.Int64Array
		Ch        chan string
		Multi     []SubModel
		Map       map[int]SubModel
		PtrModel  *SubModel
		DeletedAt *time.Time
		NamedArg  *sql.NamedArg
		SubModel
	}
	var updated map[string]any

	updated = DBUpdateTools{KeyExtractorFunc: DBUpdateGORMExtractorKey}.ToMap(
		&SubModel{Title: "test", CityAddress: "Jakarta"},
		DBUpdateSetUpdatedFields("Profile"),
	)
	assert.Equal(t, 1, len(updated))
	assert.Equal(t, nil, updated["profile"])

	updated = DBUpdateTools{KeyExtractorFunc: DBUpdateGORMExtractorKey}.ToMap(
		&Model{ID: 1, Name: candihelper.ToStringPtr("01"), Address: "street", SubModel: SubModel{Title: "test", CityAddress: "Jakarta"}, Rel: SubModel{Title: "rel sub"}},
		DBUpdateSetUpdatedFields("ID", "Name", "Title", "CityAddress"),
	)
	assert.Equal(t, 4, len(updated))
	assert.Equal(t, 1, updated["id"])
	assert.Equal(t, "01", updated["name"])
	assert.Equal(t, "test", updated["title"])
	assert.Equal(t, "Jakarta", updated["city_address"])

	updated = DBUpdateTools{KeyExtractorFunc: DBUpdateGORMExtractorKey}.ToMap(
		Model{ID: 1, Name: candihelper.ToStringPtr("01"), Address: "street", SubModel: SubModel{Title: "test", ActivatedAt: sql.NullTime{Valid: true}}},
		DBUpdateSetIgnoredFields("ID", "Name", "Title"),
	)
	assert.Equal(t, 10, len(updated))
	assert.Equal(t, "street", updated["address"])

	updated = DBUpdateTools{KeyExtractorFunc: DBUpdateGORMExtractorKey}.ToMap(
		Model{
			No:        10,
			Multi:     make([]SubModel, 1),
			Rel:       SubModel{Title: "001"},
			CreatedAt: time.Now(),
			SubModel:  SubModel{ActivatedAt: sql.NullTime{Valid: true, Time: time.Now()}, CityAddress: "Jakarta"},
			PtrModel:  &SubModel{CityAddress: "New"},
			Log:       []byte(`123`),
			StrArr:    pq.StringArray{"1", "2", "3"},
			IntArr:    pq.Int64Array{1, 2, 3},
		},
	)
	assert.Equal(t, 13, len(updated))
	assert.Equal(t, "Jakarta", updated["city_address"])
	assert.Equal(t, 10, updated["no"])
	assert.Equal(t, "{\"1\",\"2\",\"3\"}", updated["str_arr"])
	assert.Equal(t, []byte(`123`), updated["log"])
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

func TestDBUpdateSqlExtractorKey(t *testing.T) {
	type SubModel struct {
		Title       string       `sql:"column:title" json:"title"`
		Profile     string       `sql:"column:profile" json:"profile"`
		ActivatedAt sql.NullTime `sql:"column:activated_at" json:"activatedAt"`
		CityAddress string       `sql:"type:text"`
	}
	type Model struct {
		ID        int     `sql:"column:id;" json:"id"`
		Name      *string `sql:"column:name;" json:"name"`
		Address   string  `sql:"column:alamat" json:"address"`
		No        int
		CreatedAt time.Time
		IgnoreMe  SubModel `sql:"column:test" json:"ignoreMe" ignoreUpdate:"true"`
		Rel       SubModel `sql:"foreignKey:ID" json:"rel"`
		Log       []byte
		StrArr    pq.StringArray
		IntArr    pq.Int64Array
		Ch        chan string
		Multi     []SubModel
		Map       map[int]SubModel
		PtrModel  *SubModel
		DeletedAt *time.Time
		NamedArg  *sql.NamedArg
		SubModel
	}
	var updated map[string]any

	updated = DBUpdateTools{KeyExtractorFunc: DBUpdateSqlExtractorKey}.ToMap(
		&Model{ID: 1, Name: candihelper.ToStringPtr("01"), Address: "street",
			SubModel: SubModel{Title: "test", CityAddress: "Jakarta"}, Rel: SubModel{Title: "rel sub"}},
		DBUpdateSetUpdatedFields("ID", "Name", "Title", "CityAddress", "Address"),
	)

	assert.Equal(t, 5, len(updated))
	assert.Equal(t, 1, updated["id"])
	assert.Equal(t, "01", updated["name"])
	assert.Equal(t, "test", updated["title"])
	assert.Equal(t, "Jakarta", updated["city_address"])

	updated = DBUpdateTools{KeyExtractorFunc: DBUpdateSqlExtractorKey}.ToMap(
		Model{ID: 1, Name: candihelper.ToStringPtr("01"), Address: "street", SubModel: SubModel{Title: "test", ActivatedAt: sql.NullTime{Valid: true}}},
		DBUpdateSetIgnoredFields("ID", "Name", "Title"),
	)
	assert.Equal(t, 10, len(updated))
	assert.Equal(t, "street", updated["alamat"])

	updated = DBUpdateTools{KeyExtractorFunc: DBUpdateSqlExtractorKey}.ToMap(
		Model{
			No:        10,
			Multi:     make([]SubModel, 1),
			Rel:       SubModel{Title: "001"},
			CreatedAt: time.Now(),
			SubModel:  SubModel{ActivatedAt: sql.NullTime{Valid: true, Time: time.Now()}, CityAddress: "Jakarta"},
			PtrModel:  &SubModel{CityAddress: "New"},
			Log:       []byte(`123`),
			StrArr:    pq.StringArray{"1", "2", "3"},
			IntArr:    pq.Int64Array{1, 2, 3},
		},
	)
	assert.Equal(t, 13, len(updated))
	assert.Equal(t, "Jakarta", updated["city_address"])
	assert.Equal(t, 10, updated["no"])
	assert.Equal(t, "{\"1\",\"2\",\"3\"}", updated["str_arr"])
	assert.Equal(t, []byte(`123`), updated["log"])
}

func TestDBUpdateToolsMongo2(t *testing.T) {
	type Model struct {
		Title       string  `bson:"title" json:"title"`
		Profile     *string `bson:"profile" json:"profile,omitempty"`
		CityAddress *string `bson:"city_address,omitempty" json:"city_address"`
	}

	updated := DBUpdateTools{}.ToMap(&Model{})
	candihelper.PrintJSON(updated)
}
