package game

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewClass(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer db.Close()

	expectedID := 1
	expectedName := "Warrior"
	expectedCreateTS := time.Now()
	expectedUpdateTS := time.Now()
	var expectedDeleteTS sql.NullTime

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO game.dim_class ("name")`)).
		WithArgs(expectedName).
		WillReturnRows(sqlmock.NewRows([]string{"id", "create_ts", "update_ts", "delete_ts"}).
			AddRow(expectedID, expectedCreateTS, expectedUpdateTS, expectedDeleteTS))

	class, err := NewClass(db, expectedName)
	if err != nil {
		t.Fatalf("failed to create class: %v", err)
	}

	if class.ID != expectedID {
		t.Errorf("unexpected class ID, expected %d, got %d", expectedID, class.ID)
	}
	if class.Name != expectedName {
		t.Errorf("unexpected class name, expected %s, got %s", expectedName, class.Name)
	}
	if class.CreateTS != expectedCreateTS {
		t.Errorf("unexpected class create timestamp, expected %v, got %v", expectedCreateTS, class.CreateTS)
	}
	if class.UpdateTS != expectedUpdateTS {
		t.Errorf("unexpected class update timestamp, expected %v, got %v", expectedUpdateTS, class.UpdateTS)
	}
	if class.DeleteTS != expectedDeleteTS {
		t.Errorf("unexpected class delete timestamp, expected %v, got %v", expectedDeleteTS, class.DeleteTS)
	}
}

func TestGetAllClasses(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer db.Close()

	expectedRows := sqlmock.NewRows([]string{"id", "name", "create_ts", "update_ts", "delete_ts"}).
		AddRow(1, "Warrior", time.Now(), time.Now(), nil).
		AddRow(2, "Mage", time.Now(), time.Now(), nil).
		AddRow(3, "Rogue", time.Now(), time.Now(), nil)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, \"name\", create_ts, update_ts, delete_ts")).
		WillReturnRows(expectedRows)

	classes, err := GetAllClasses(db)
	if err != nil {
		t.Fatalf("failed to get classes: %v", err)
	}

	if len(classes) != 3 {
		t.Errorf("unexpected number of classes, expected 3, got %d", len(classes))
	}
	if classes[0].ID != 1 {
		t.Errorf("unexpected class ID, expected 1, got %d", classes[0].ID)
	}
	if classes[1].Name != "Mage" {
		t.Errorf("unexpected class name, expected Mage, got %s", classes[1].Name)
	}
	if classes[2].UpdateTS.IsZero() {
		t.Errorf("unexpected zero update timestamp for class: %v", classes[2].UpdateTS)
	}
}
