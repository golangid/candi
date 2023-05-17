package candishared

import (
	"testing"

	"github.com/golangid/candi/candihelper"
	"github.com/stretchr/testify/assert"
)

func TestDBUpdateTools(t *testing.T) {
	type SubModel struct {
		Title   string `gorm:"column:title" json:"title"`
		Profile string `gorm:"column:profile" json:"profile"`
	}

	type Model struct {
		ID       int      `gorm:"column:db_id;" json:"id"`
		Name     *string  `gorm:"column:db_name;" json:"name"`
		Address  string   `gorm:"column:db_address" json:"address"`
		IgnoreMe SubModel `gorm:"column:test" json:"ignoreMe" ignoreUpdate:"true"`
		Rel      SubModel `gorm:"foreignKey:ID" json:"rel"`
		SubModel
	}

	updated := DBUpdateTools{KeyExtractorFunc: DBUpdateGORMExtractorKey}.ToMap(
		&Model{ID: 1, Name: candihelper.ToStringPtr("01"), Address: "street", SubModel: SubModel{Title: "test"}, Rel: SubModel{Title: "rel sub"}},
		DBUpdateSetUpdatedFields("ID", "Name", "Title"),
	)
	assert.Equal(t, 3, len(updated))
	assert.Equal(t, 1, updated["db_id"])
	assert.Equal(t, "01", updated["db_name"])
	assert.Equal(t, "test", updated["title"])

	updated = DBUpdateTools{}.ToMap(
		Model{ID: 1, Name: candihelper.ToStringPtr("01"), Address: "street", SubModel: SubModel{Title: "test"}},
		DBUpdateSetIgnoredFields("ID", "Name", "Title"),
	)
	assert.Equal(t, 3, len(updated))
	assert.Equal(t, "street", updated["address"])

	updated = DBUpdateTools{}.ToMap(
		Model{ID: 1, Name: candihelper.ToStringPtr("01"), Address: "street", IgnoreMe: SubModel{Title: "t"}, SubModel: SubModel{Title: "test"}},
	)
	assert.Equal(t, 6, len(updated))
	assert.Equal(t, "street", updated["address"])
}
