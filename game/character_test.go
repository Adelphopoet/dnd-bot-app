package game

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCharacterSave(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer db.Close()

	expectedID := 1
	expectedName := "John"
	expectedCreateTS := time.Now()
	expectedUpdateTS := time.Now()
	var expectedDeleteTS sql.NullTime

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO game.dim_character ("name")`)).
		WithArgs(expectedName).
		WillReturnRows(sqlmock.NewRows([]string{"id", "create_ts", "update_ts", "delete_ts"}).
			AddRow(expectedID, expectedCreateTS, expectedUpdateTS, expectedDeleteTS))

	// Вместо этого теста для bridge_character_class
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO game.bridge_character_class (character_id, class_id)`)).
		WithArgs(expectedID, 1).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO game.bridge_character_class (character_id, class_id)`)).
		WithArgs(expectedID, 2).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectCommit()

	character := NewCharacter(db, expectedName, []int{1, 2})
	err = character.Save()
	if err != nil {
		t.Fatalf("failed to save character: %v", err)
	}

	if character.ID != expectedID {
		t.Errorf("unexpected character ID, expected %d, got %d", expectedID, character.ID)
	}
	if character.Name != expectedName {
		t.Errorf("unexpected character name, expected %s, got %s", expectedName, character.Name)
	}
	if character.CreateTS != expectedCreateTS {
		t.Errorf("unexpected character create timestamp, expected %v, got %v", expectedCreateTS, character.CreateTS)
	}
	if character.UpdateTS != expectedUpdateTS {
		t.Errorf("unexpected character update timestamp, expected %v, got %v", expectedUpdateTS, character.UpdateTS)
	}
	if character.DeleteTS != expectedDeleteTS {
		t.Errorf("unexpected character delete timestamp, expected %v, got %v", expectedDeleteTS, character.DeleteTS)
	}
}

func TestCharacterDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer db.Close()

	expectedName := "John"

	mock.ExpectQuery(regexp.QuoteMeta(`UPDATE game.dim_character`)).
		WithArgs(expectedName).
		WillReturnRows(sqlmock.NewRows([]string{"id", "create_ts", "update_ts", "delete_ts"}).
			AddRow(1, time.Now(), time.Now(), nil))

	character := &Character{Name: expectedName, db: db}
	err = character.Delete()
	if err != nil {
		t.Fatalf("failed to delete character: %v", err)
	}

	if !character.IsDeleted {
		t.Errorf("unexpected character IsDeleted value, expected true, got false")
	}
}

func TestCharacterLoad(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer db.Close()

	expectedID := 1
	expectedName := "John"
	expectedCreateTS := time.Now()
	expectedUpdateTS := time.Now()
	var expectedDeleteTS sql.NullTime
	expectedIsDeleted := false

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "name", create_ts, update_ts, delete_ts, is_deleted FROM game.dim_character WHERE id = $1`)).
		WithArgs(expectedID).
		WillReturnRows(sqlmock.NewRows([]string{"name", "create_ts", "update_ts", "delete_ts", "is_deleted"}).
			AddRow(expectedName, expectedCreateTS, expectedUpdateTS, expectedDeleteTS, expectedIsDeleted))

	character := &Character{ID: expectedID, db: db}
	err = character.Load()
	if err != nil {
		t.Fatalf("failed to load character: %v", err)
	}

	if character.Name != expectedName {
		t.Errorf("unexpected character name, expected %s, got %s", expectedName, character.Name)
	}
	if character.CreateTS != expectedCreateTS {
		t.Errorf("unexpected character create timestamp, expected %v, got %v", expectedCreateTS, character.CreateTS)
	}
	if character.UpdateTS != expectedUpdateTS {
		t.Errorf("unexpected character update timestamp, expected %v, got %v", expectedUpdateTS, character.UpdateTS)
	}
	if character.DeleteTS != expectedDeleteTS {
		t.Errorf("unexpected character delete timestamp, expected %v, got %v", expectedDeleteTS, character.DeleteTS)
	}
	if character.IsDeleted != expectedIsDeleted {
		t.Errorf("unexpected character IsDeleted value, expected %t, got %t", expectedIsDeleted, character.IsDeleted)
	}
}
