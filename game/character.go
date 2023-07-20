package game

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type Character struct {
	ID             int
	Name           string
	CreateTS       time.Time
	UpdateTS       time.Time
	DeleteTS       sql.NullTime
	db             *sql.DB
	IsDeleted      bool
	ClassIDs       []int
	CharacterClass []*CharacterClass
	Attributes     []*CharacterAtribute
}

type CharacterAtribute struct {
	Attribute *Attribute
	Value     *AttributeValue
}

type CharacterClass struct {
	Class *Class
	Lvl   int
}

func NewCharacter(db *sql.DB, name string, classIDs []int) *Character {
	return &Character{
		Name:     name,
		ClassIDs: classIDs,
		db:       db,
	}
}

func (c *Character) Save() error {
	tx, err := c.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	query := `
		INSERT INTO game.dim_character ("name")
		VALUES ($1)
		RETURNING id, create_ts, update_ts, delete_ts
	`
	err = tx.QueryRow(query, c.Name).Scan(&c.ID, &c.CreateTS, &c.UpdateTS, &c.DeleteTS)
	if err != nil {
		tx.Rollback()
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return fmt.Errorf("Персонаж с таким именем уже существует.")
		}
		return fmt.Errorf("failed to save character: %v", err)
	}

	// Связывание персонажа с классами в таблице bridge_character_class
	for _, classID := range c.ClassIDs {
		query = `
			INSERT INTO game.bridge_character_class (character_id, class_id, lvl)
			VALUES ($1, $2, 0)
		`
		_, err = tx.Exec(query, c.ID, classID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to save character class: %v", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func (c *Character) Delete() error {
	query := `
		UPDATE game.dim_character
		SET is_deleted = true, update_ts = current_timestamp
		WHERE id = 1;
	`
	err := c.db.QueryRow(query, c.Name).Scan(&c.ID, &c.CreateTS, &c.UpdateTS, &c.DeleteTS)
	if err != nil {
		return fmt.Errorf("failed to delete character: %v", err)
	}
	c.IsDeleted = true // Устанавливаем флаг IsDeleted в true
	return nil
}

func (c *Character) Load() error {
	query := `
		SELECT "name", create_ts, update_ts, delete_ts, is_deleted
		FROM game.dim_character
		WHERE id = $1
	`
	err := c.db.QueryRow(query, c.ID).Scan(&c.Name, &c.CreateTS, &c.UpdateTS, &c.DeleteTS, &c.IsDeleted)
	if err != nil {
		return fmt.Errorf("failed to load character: %v", err)
	}
	return nil
}

func (c *Character) GetCharacterClasses() ([]*CharacterClass, error) {
	query := `
		SELECT cc.class_id, cc.lvl
		FROM game.bridge_character_class cc
		WHERE cc.character_id = $1
	`
	rows, err := c.db.Query(query, c.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to scan character class: %v", err)
	}
	defer rows.Close()

	var characterClasses []*CharacterClass
	for rows.Next() {
		var lvl, classID int
		err := rows.Scan(&classID, &lvl)
		if err != nil {
			return nil, fmt.Errorf("failed to scan character class: %v", err)
		}

		class, err := GetClassById(c.db, classID)
		if err != nil {
			return nil, fmt.Errorf("failed to get class: %v", err)
		}
		characterClasses = append(characterClasses,
			&CharacterClass{
				Class: class,
				Lvl:   lvl,
			})
	}
	c.CharacterClass = characterClasses
	return characterClasses, nil
}

func GetAllUserCharacters(db *sql.DB, userID int64) ([]*Character, error) {
	query := `
		SELECT character_id
		FROM game.bridge_tg_user_character
		WHERE user_id = $1
		ORDER BY is_main_character DESC, update_ts ASC
	`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user characters: %v", err)
	}
	defer rows.Close()

	var characters []*Character
	if rows != nil {
		isMainCharacter := true // Флаг для определения главного персонажа
		for rows.Next() {
			var characterID int
			err := rows.Scan(&characterID)
			if err != nil {
				return nil, fmt.Errorf("failed to scan user character: %v", err)
			}

			character, err := GetCharacterByID(db, characterID)
			if err != nil {
				return nil, fmt.Errorf("failed to load user character: %v", err)
			}

			if isMainCharacter {
				isMainCharacter = false
			}

			characters = append(characters, character)
		}

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error in rows: %v", err)
		}
	}

	return characters, nil
}

func GetActiveCharacter(db *sql.DB, userID int64) (*Character, error) {
	query := `
		SELECT c.id
		FROM game.dim_character c
		JOIN game.bridge_tg_user_character b ON c.id = b.character_id
		WHERE b.user_id = $1 
		AND b.is_main_character = true 
		AND c.is_deleted = false
		AND b.is_deleted = false
	`
	var characterID int

	err := db.QueryRow(query, userID).Scan(&characterID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Нет активного персонажа
		}
		return nil, fmt.Errorf("failed to get active character: %v", err)
	}
	log.Println("Start search char by id")
	character, err := GetCharacterByID(db, characterID)
	if err != nil {
		return nil, nil // No actual character
	}

	return character, nil
}

func GetCharacterByID(db *sql.DB, characterID int) (*Character, error) {
	query := `
		SELECT "name", create_ts, update_ts, delete_ts, is_deleted
		FROM game.dim_character
		WHERE id = $1
	`
	var character Character
	err := db.QueryRow(query, characterID).Scan(&character.Name, &character.CreateTS, &character.UpdateTS, &character.DeleteTS, &character.IsDeleted)
	if err != nil {
		return nil, fmt.Errorf("failed to load character: %v", err)
	}

	character.db = db
	_, err = character.GetCharacterClasses()
	if err != nil {
		return nil, fmt.Errorf("Error during get character classes: ", err)
	}

	character.ID = characterID
	_, err = character.GetAttributeValues()
	if err != nil {
		return nil, fmt.Errorf("Error during get character attributes: ", err)
	}
	return &character, nil
}

func GetCharacterByName(db *sql.DB, characterName string) (*Character, error) {
	query := `
		SELECT id, "name", create_ts, update_ts, delete_ts, is_deleted
		FROM game.dim_character
		WHERE "name" = $1
		AND is_deleted = False
	`
	var character Character
	err := db.QueryRow(query, characterName).Scan(&character.ID, &character.Name, &character.CreateTS, &character.UpdateTS, &character.DeleteTS, &character.IsDeleted)
	if err != nil {
		return nil, fmt.Errorf("failed to load character: %v", err)
	}
	character.db = db

	_, err = character.GetCharacterClasses()
	if err != nil {
		return nil, fmt.Errorf("Error during get character classes: ", err)
	}

	_, err = character.GetAttributeValues()
	if err != nil {
		return nil, fmt.Errorf("Error during get character attributes: ", err)
	}
	return &character, nil
}

func (c *Character) SetLocation(locationID int) error {
	log.Printf("Start set location to character %v, new location is %v", c.Name, locationID)
	tx, err := c.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	// Удаление предыдущей записи о местоположении персонажа, если есть
	_, err = tx.Exec("DELETE FROM game.bridge_character_location WHERE character_id = $1", c.ID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete previous character location: %v", err)
	}

	// Вставка новой записи о местоположении персонажа
	_, err = tx.Exec("INSERT INTO game.bridge_character_location (character_id, location_id) VALUES ($1, $2)", c.ID, locationID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to set character location: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func (c *Character) GetCurrentLocation() (*Location, error) {
	query := `
		SELECT l.id
		FROM game.dim_location l
		JOIN game.bridge_character_location b ON l.id = b.location_id
		WHERE b.character_id = $1
	`
	var locationID int
	err := c.db.QueryRow(query, c.ID).Scan(&locationID)
	location, err := GetLocationByID(c.db, locationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Нет текущей локации для персонажа
		}
		return nil, fmt.Errorf("failed to get current location: %v", err)
	}

	return location, nil
}

func (c *Character) GetAttributeValues() ([]*CharacterAtribute, error) {
	query := `
		SELECT 
			attribute_id, 
			numeric_value, 
			string_value, 
			formula_value, 
			bool_value
		FROM game.bridge_character_attribute_value
		WHERE character_id = $1
	`
	rows, err := c.db.Query(query, c.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character attributes: %v", err)
	}
	defer rows.Close()

	var attributes []*CharacterAtribute
	for rows.Next() {
		var attId int
		attVal := &AttributeValue{FormulaValue: &Formula{}}

		err := rows.Scan(
			&attId,
			&attVal.NumericValue,
			&attVal.StringValue,
			&attVal.FormulaValue.Expression,
			&attVal.BoolValue,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan character attribute: %v", err)
		}
		attr, err := GetAttributeByID(c.db, attId)
		if err != nil {
			return nil, err
		}
		attributes = append(attributes, &CharacterAtribute{Attribute: attr, Value: attVal})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error in character attributes rows: %v", err)
	}
	c.Attributes = attributes
	return attributes, nil
}

func (c *Character) GetAttributeValueByName(name string) *AttributeValue {
	var value *AttributeValue = nil
	for _, attr := range c.Attributes {
		if attr.Attribute.Name == name {
			value = attr.Value
			break
		}
	}

	return value
}

func (c *Character) getSummaryLvl() (int, error) {
	var summaryLvl int
	query := `
		SELECT 
			SUM(lvl) lvl
		FROM game.bridge_character_class
		WHERE character_id = $1
		AND is_deleted = False
	`
	rows, err := c.db.Query(query, c.ID)
	if err != nil {
		return summaryLvl, fmt.Errorf("failed to get character summary lvl: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&summaryLvl)
		if err != nil {
			return summaryLvl, fmt.Errorf("failed to scan character summary lvl: %v", err)
		}
	}
	return summaryLvl, nil
}

func CheckCharacterCanLvlUp(character *Character) (bool, error) {
	// Get character LVl
	summaryLvl, err := character.getSummaryLvl()
	if err != nil {
		return false, err
	}

	// Get character EXP
	currentXpAtt := character.GetAttributeValueByName("exp")

	var currentXp int = 0

	if currentXpAtt == nil {
		currentXp = 0
	} else {
		currentXp = currentXpAtt.NumericValue
	}

	// Calc exp to rich next LVL
	nextLvl := summaryLvl + 1
	nextLvlInfo, err := GetLvlInfo(character.db, nextLvl)
	if err != nil {
		return false, err
	}

	// Max lvl reached
	if nextLvlInfo == nil {
		return false, nil
	}

	nextLvlExp := nextLvlInfo.Exp

	if currentXp >= nextLvlExp {
		return true, nil
	} else {
		return false, nil
	}
}

func (c *Character) SetAttributeValue(attribute *Attribute, value *AttributeValue) error {
	query := `
		INSERT INTO game.bridge_character_attribute_value (character_id, attribute_id, numeric_value, string_value, formula_value, bool_value)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (character_id, attribute_id) DO UPDATE
		SET numeric_value = EXCLUDED.numeric_value,
			string_value = EXCLUDED.string_value,
			formula_value = EXCLUDED.formula_value,
			bool_value = EXCLUDED.bool_value
	`
	// Avaiod nill in str
	var formulaValue string
	if value.FormulaValue == nil {
		formulaValue = ""
	} else {
		formulaValue = value.FormulaValue.Expression
	}
	_, err := c.db.Exec(query, c.ID, attribute.ID, value.NumericValue, value.StringValue, formulaValue, value.BoolValue)
	if err != nil {
		return fmt.Errorf("failed to set attribute value: %v", err)
	}
	return nil
}

func (c *Character) addExp(addExp int) error {
	// init exp type
	attribute, err := GetAttributeByName(c.db, "exp")
	if err != nil {
		return err
	}

	// Get current exp val
	currentValue := c.GetAttributeValueByName("exp")
	if err != nil {
		return fmt.Errorf("failed to get current exp value: %v", err)
	}

	var currentExp int
	if currentValue == nil {
		currentExp = 0
	} else {
		currentExp = currentValue.NumericValue
	}

	newValue := currentExp + addExp
	value := &AttributeValue{
		NumericValue: newValue,
	}

	err = c.SetAttributeValue(attribute, value)
	if err != nil {
		return fmt.Errorf("failed to add exp: %v", err)
	}

	return nil
}

func (c *Character) LvlUp(class *Class) (bool, error) {
	// Check character can lvl up
	canLvl, err := CheckCharacterCanLvlUp(c)
	if err != nil {
		return false, err
	}

	if !canLvl {
		return false, nil
	}

	// Start lvling
	query := `
		INSERT INTO game.bridge_character_class (character_id, class_id, lvl)
		VALUES ($1, $2, 1)
		ON CONFLICT (character_id, class_id) DO UPDATE
		SET lvl = bridge_character_class.lvl + 1
		WHERE bridge_character_class.character_id = $1
		AND bridge_character_class.class_id = $2
	`
	_, err = c.db.Exec(query, c.ID, class.ID)
	if err != nil {
		return false, fmt.Errorf("failed to level up: %v", err)
	}

	// Get CON value
	_, err = c.GetAttributeValues()
	if err != nil {
		return false, err
	}
	conAtt := c.GetAttributeValueByName("con")
	conBonus, err := GetAttributeBonus(conAtt)
	if err != nil {
		return true, err
	}

	// Roll for start HP
	hitDiceFormula, err := class.GetAttributeFormulaByName("hp dice")
	if err != nil {
		return true, err
	}
	maxHp, err := CalculateFormula(hitDiceFormula)
	if err != nil {
		return true, err
	}

	maxHp += conBonus // Add bonus from CON
	hpAtt, err := GetAttributeByName(c.db, "hp")
	if err != nil {
		return false, err
	}
	err = c.SetAttributeValue(hpAtt, &AttributeValue{NumericValue: maxHp})
	if err != nil {
		return false, err
	}

	return false, nil
}
