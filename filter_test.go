package golf

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type testModel struct {
	ID       int    `json:"id"`
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
}

func (m *testModel) Field() map[string][]Filter {
	return map[string][]Filter{
		"ID":     {Equal},
		"UserID": {Equal, Gte},
	}
}

func NewDB(db *sql.DB) (*gorm.DB, error) {
	return gorm.Open(
		postgres.New(postgres.Config{
			PreferSimpleProtocol: true,
			Conn:                 db,
		}),
		&gorm.Config{
			Logger: logger.Default,
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true,
			},
			AllowGlobalUpdate: false,
		})
}

func TestGolf_Do(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	gormDB, err := NewDB(db)
	assert.NoError(t, err)
	gol := NewGolf(gormDB)
	m := &testModel{}
	cases := []struct {
		Name   string
		query  map[string][]string
		golf   *Golf
		Except string
	}{
		{
			query: map[string][]string{
				"eq_id": {"1"},
			},
			golf:   NewGolf(gormDB),
			Except: regexp.QuoteMeta(`SELECT * FROM "test_model"`),
		},
	}
	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			assert.NoError(t, gol.Build(m, tt.query).Error)
			var testMod []testModel
			// fixme SQL is not completely written because where conditional disorder
			mock.ExpectQuery(tt.Except).
				WillReturnRows(
					sqlmock.NewRows([]string{"id"}).AddRow("1"))
			assert.NoError(t, gol.Find(&testMod).Error)
			assert.NoError(t, mock.ExpectationsWereMet())
			assert.Equal(t, len(testMod), 1)
		})
	}

}

func TestCheckFilter(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	gormDB, err := NewDB(db)
	assert.NoError(t, err)
	gol := NewGolf(gormDB)
	query := map[string][]string{
		"eq_id":           {"1"},
		"eq_created_user": {"1"},
	}
	lowerQ := map[string][]Filter{
		"id":           {Equal},
		"created_user": {Equal},
	}
	realQ, err := gol.checkAndBuildQuery(lowerQ, query)
	assert.NoError(t, err)
	t.Log(realQ)
}
