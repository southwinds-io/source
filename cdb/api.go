/*
  Onix Config Manager - Source
  © 2022 southwinds.io - All rights reserved
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package cdb

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/invopop/jsonschema"
	schemaValidation "github.com/qri-io/jsonschema"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	sqlDriver = "sqlite"
)

var (
	ErrNotFound         = errors.New("noy found")
	ErrInvalidItemType  = errors.New("invalid item type")
	ErrInvalidItemValue = errors.New("invalid item value, schema verification failed")
)

// DataBase the definition of the configuration database
type DataBase struct {
	db *sql.DB
}

// I the definition of an item
type I struct {
	Key     string      `json:"key"`
	Type    string      `json:"type"`
	Value   interface{} `json:"value"`
	Updated time.Time   `json:"updated"`
}

// L the definition of a configuration link
type L struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// T the definition of an item tag
type T struct {
	ItemKey string `json:"item_key,omitempty"`
	Name    string `json:"name"`
	Value   string `json:"value"`
}

// TT the definition of an item type
type TT struct {
	Key    string `json:"key"`
	Schema string `json:"schema"`
}

// New create a new configuration database on the specified path
func New(path string) (*DataBase, error) {
	var err error
	m := new(DataBase)
	if m.db, err = getDb(path); err != nil {
		return nil, err
	}
	return m, nil
}

// SetTypeFromString set the json schema for an item type using a json string representation of the schema
func (d *DataBase) SetTypeFromString(key, schema string) error {
	stmt := `INSERT INTO type(key, schema) VALUES(?, ?) ON CONFLICT(key) DO UPDATE SET schema = excluded.schema;`
	statement, err := d.db.Prepare(stmt)
	if err != nil {
		return err
	}
	_, err = statement.Exec(key, schema)
	return err
}

// SetTypeFromStruct set the json schema for the item type by inferring it from the passed in object
func (d *DataBase) SetTypeFromStruct(key string, obj interface{}) error {
	// reflects the json schema from the specified object
	schemaObj := jsonschema.Reflect(obj)
	// marshal the object to json bytes
	schema, err := json.Marshal(schemaObj)
	if err != nil {
		return err
	}
	return d.SetTypeFromString(key, string(schema[:]))
}

// DeleteType delete a json schema for an item type
// Can be done with existing items, as a result items of the missing type are not validated
func (d *DataBase) DeleteType(key string) error {
	stmt := `DELETE FROM type WHERE key=?; `
	statement, err := d.db.Prepare(stmt)
	if err != nil {
		return err
	}
	_, err = statement.Exec(key)
	return err
}

// SetItem set the value of an item
func (d *DataBase) SetItem(key, iType string, value interface{}, validate bool) (error, bool) {
	if value == nil {
		return fmt.Errorf("value not provided"), false
	}
	if sv, ok := value.(string); ok {
		return d.setItemString(key, iType, sv, validate)
	}
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return err, false
	}
	return d.setItemString(key, iType, string(valueBytes[:]), validate)
}

// DeleteItem delete the specified item
func (d *DataBase) DeleteItem(key string) error {
	// delete the item
	stmt := `DELETE FROM item WHERE key=?; `
	statement, err := d.db.Prepare(stmt)
	if err != nil {
		return err
	}
	_, err = statement.Exec(key)
	if err != nil {
		return err
	}
	// delete any associations
	statement, err = d.db.Prepare("DELETE FROM link WHERE from_key=? OR to_key=?;")
	if err != nil {
		return err
	}
	_, err = statement.Exec(key, key)
	if err != nil {
		return err
	}
	// delete any tags
	statement, err = d.db.Prepare("DELETE FROM tag WHERE item_key=?;")
	if err != nil {
		return err
	}
	_, err = statement.Exec(key)
	return err
}

// Link add an association between two items
func (d *DataBase) Link(from, to string) error {
	stmt := `INSERT INTO link(from_key, to_key) VALUES(?, ?); `
	statement, err := d.db.Prepare(stmt)
	if err != nil {
		return err
	}
	_, err = statement.Exec(from, to)
	return err
}

// UnLink remove an association between two items
func (d *DataBase) UnLink(from, to string) error {
	stmt := `DELETE FROM link WHERE from_key=? AND to_key=?; `
	statement, err := d.db.Prepare(stmt)
	if err != nil {
		return err
	}
	_, err = statement.Exec(from, to)
	return err
}

// Tag an item with  a name only (value is empty)
func (d *DataBase) Tag(key, name string) error {
	return d.TagValue(key, name, "")
}

// TagValue tag an item with a name and a value
func (d *DataBase) TagValue(key, name, value string) error {
	stmt := `INSERT INTO tag(item_key, name, value) VALUES(?, ?, ?) ON CONFLICT(item_key, name) DO UPDATE SET value = excluded.value;`
	statement, err := d.db.Prepare(stmt)
	if err != nil {
		return err
	}
	_, err = statement.Exec(key, name, value)
	return err
}

// Untag a configuration
func (d *DataBase) Untag(key string, name string) error {
	stmt := `DELETE FROM tag WHERE item_key=? AND name=?; `
	statement, err := d.db.Prepare(stmt)
	if err != nil {
		return err
	}
	_, err = statement.Exec(key, name)
	return err
}

func (d *DataBase) DeleteLinks() interface{} {
	stmt := `DELETE FROM link;`
	statement, err := d.db.Prepare(stmt)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	return err
}

// GetItem get an item by key
func (d *DataBase) GetItem(key string) (*I, error) {
	row := d.db.QueryRow(`SELECT type, value, updated FROM item WHERE key=?;`, key)
	var (
		itype, value string
		updated      sql.NullInt64
	)
	err := row.Scan(&itype, &value, &updated)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &I{
		Key:     key,
		Type:    itype,
		Value:   value,
		Updated: time.Unix(0, updated.Int64).UTC(),
	}, nil
}

// GetTaggedItems get the items with the specified tag names
func (d *DataBase) GetTaggedItems(tags ...string) ([]I, error) {
	stmt := "SELECT DISTINCT i.key, i.type, i.value, i.updated FROM item i INNER JOIN tag t ON i.key = t.item_key WHERE t.name" + toInSqlTags(tags)
	row, err := d.db.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer func(row *sql.Rows) {
		err = row.Close()
		if err != nil {
			fmt.Printf("cannot close query row: %s\n", err)
		}
	}(row)
	var (
		key, itype, value string
		updated           sql.NullInt64
	)
	var items []I
	for row.Next() {
		err = row.Scan(&key, &itype, &value, &updated)
		if err != nil {
			return nil, err
		}
		items = append(items, I{
			Key:     key,
			Type:    itype,
			Value:   value,
			Updated: time.Unix(0, updated.Int64).UTC(),
		})
	}
	return items, nil
}

// GetTags get the tags (name & value) of an item
func (d *DataBase) GetTags(key string) ([]T, error) {
	stmt := fmt.Sprintf("SELECT name, value FROM tag WHERE item_key=?;")
	row, err := d.db.Query(stmt, key)
	if err != nil {
		return nil, err
	}
	defer func(row *sql.Rows) {
		err = row.Close()
		if err != nil {
			fmt.Printf("cannot close query row: %s\n", err)
		}
	}(row)
	var name, value string
	var tags []T
	for row.Next() {
		err = row.Scan(&name, &value)
		if err != nil {
			return nil, err
		}
		tags = append(tags, T{
			Name:  name,
			Value: value,
		})
	}
	return tags, nil
}

func (d *DataBase) GetAllTags() ([]T, error) {
	stmt := fmt.Sprintf("SELECT item_key, name, value FROM tag;")
	row, err := d.db.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer func(row *sql.Rows) {
		err = row.Close()
		if err != nil {
			fmt.Printf("cannot close query row: %s\n", err)
		}
	}(row)
	var itemKey, name, value string
	var tags []T
	for row.Next() {
		err = row.Scan(&itemKey, &name, &value)
		if err != nil {
			return nil, err
		}
		tags = append(tags, T{
			ItemKey: itemKey,
			Name:    name,
			Value:   value,
		})
	}
	return tags, nil
}

// GetChildren get the child items linked to a specified item
func (d *DataBase) GetChildren(parentKey string) ([]I, error) {
	row, err := d.db.Query("SELECT i.key, i.type, i.value, i.updated FROM link l INNER JOIN item i ON l.to_key = i.key WHERE l.from_key=?;", parentKey)
	if err != nil {
		return nil, err
	}
	defer func(row *sql.Rows) {
		err = row.Close()
		if err != nil {
			fmt.Printf("cannot close query row: %s\n", err)
		}
	}(row)
	var (
		key, itype, value string
		updated           sql.NullInt64
	)
	var items []I
	for row.Next() {
		err = row.Scan(&key, &itype, &value, &updated)
		if err != nil {
			return nil, err
		}
		items = append(items, I{
			Key:     key,
			Type:    itype,
			Value:   value,
			Updated: time.Unix(0, updated.Int64).UTC(),
		})
	}
	return items, nil
}

// GetParents get the parent items linked to a specified item
func (d *DataBase) GetParents(childKey string) ([]I, error) {
	row, err := d.db.Query("SELECT i.key, i.type, i.value, i.updated FROM link l INNER JOIN item i ON l.from_key = i.key where l.to_key=?;", childKey)
	if err != nil {
		return nil, err
	}
	defer func(row *sql.Rows) {
		err = row.Close()
		if err != nil {
			fmt.Printf("cannot close query row: %s\n", err)
		}
	}(row)
	var (
		key, itype, value string
		updated           sql.NullInt64
	)
	var items []I
	for row.Next() {
		err = row.Scan(&key, &itype, &value, &updated)
		if err != nil {
			return nil, err
		}
		items = append(items, I{
			Key:     key,
			Type:    itype,
			Value:   value,
			Updated: time.Unix(0, updated.Int64).UTC(),
		})
	}
	return items, nil
}

func (d *DataBase) GetItems() ([]I, error) {
	row, err := d.db.Query(`SELECT key, type, value, updated FROM item;`)
	if err != nil {
		return nil, err
	}
	defer func(row *sql.Rows) {
		err = row.Close()
		if err != nil {
			fmt.Printf("cannot close query row: %s\n", err)
		}
	}(row)
	var (
		key, iType, value string
		updated           sql.NullInt64
		items             []I
	)
	for row.Next() {
		err = row.Scan(&key, &iType, &value, &updated)
		if err != nil {
			if strings.Contains(err.Error(), "no rows") {
				return nil, ErrNotFound
			}
			return nil, err
		}
		items = append(items, I{
			Key:     key,
			Type:    iType,
			Value:   value,
			Updated: time.Unix(0, updated.Int64).UTC(),
		})
	}
	return items, nil
}

func (d *DataBase) GetLinks() ([]L, error) {
	row, err := d.db.Query(`SELECT from_key, to_key FROM link;`)
	if err != nil {
		return nil, err
	}
	defer func(row *sql.Rows) {
		err = row.Close()
		if err != nil {
			fmt.Printf("cannot close query row: %s\n", err)
		}
	}(row)
	var (
		from, to string
		links    []L
	)
	for row.Next() {
		err = row.Scan(&from, &to)
		if err != nil {
			if strings.Contains(err.Error(), "no rows") {
				return nil, ErrNotFound
			}
			return nil, err
		}
		links = append(links, L{
			From: from,
			To:   to,
		})
	}
	return links, nil
}

func (d *DataBase) GetTypes() ([]TT, error) {
	row, err := d.db.Query(`SELECT key, schema FROM type;`)
	if err != nil {
		return nil, err
	}
	defer func(row *sql.Rows) {
		err = row.Close()
		if err != nil {
			fmt.Printf("cannot close query row: %s\n", err)
		}
	}(row)
	var key, schema string
	var types []TT
	for row.Next() {
		err = row.Scan(&key, &schema)
		if err != nil {
			if strings.Contains(err.Error(), "no rows") {
				return nil, ErrNotFound
			}
			return nil, err
		}
		types = append(types, TT{
			Key:    key,
			Schema: schema,
		})
	}
	return types, nil
}

func (d *DataBase) GetSchema(key string) (string, error) {
	row := d.db.QueryRow(`SELECT schema FROM type WHERE key = ?;`, key)
	var result string
	err := row.Scan(&result)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return "", ErrNotFound
		}
		return "", err
	}
	return result, nil
}

func (d *DataBase) setItemString(key, iType, value string, validate bool) (error, bool) {
	// if a type is provided
	if validate {
		// get the schema for the type
		s, err := d.GetSchema(iType)
		if err != nil {
			return err, false
		}
		// validates only if a schema has been defined
		if len(s) > 0 {
			ctx := context.Background()
			var schemaData = []byte(s)
			rs := &schemaValidation.Schema{}
			if err = json.Unmarshal(schemaData, rs); err != nil {
				return fmt.Errorf("unmarshal schema: %s", err), false
			}
			// validate the value using the stored schema
			errs, err := rs.ValidateBytes(ctx, []byte(value))
			if err != nil {
				return err, true
			}
			if len(errs) > 0 {
				return errs[0], true
			}
		} else if len(iType) > 0 {
			return ErrInvalidItemType, false
		}
	}
	stmt := `INSERT INTO item(key, type, value, updated) VALUES(?, ?, ?, ?) ON CONFLICT(key) DO UPDATE SET type = excluded.type, value = excluded.value, updated = excluded.updated;`
	statement, err := d.db.Prepare(stmt)
	if err != nil {
		return err, false
	}
	_, err = statement.Exec(key, iType, value, time.Now().UTC().UnixNano())
	return err, false
}

func getDb(path string) (db *sql.DB, err error) {
	path, err = filepath.Abs(path)
	path = filepath.Join(path, ".cfg.db")
	// if the index db does not exist
	if _, err = os.Stat(path); os.IsNotExist(err) {
		db, err = createDb(path)
		if err != nil {
			return db, err
		}
	} else {
		db, err = sql.Open(sqlDriver, path)
		if err != nil {
			return nil, err
		}
	}
	return db, nil
}

func createDb(path string) (*sql.DB, error) {
	file, err := os.Create(path) // Create SQLite file
	if err != nil {
		return nil, fmt.Errorf("cannot create index database: %s\n", err)
	}
	err = file.Close()
	if err != nil {
		return nil, fmt.Errorf("cannot close index database: %s\n", err)
	}
	db, err := sql.Open(sqlDriver, path)
	if err != nil {
		return nil, err
	}
	if err = createSchema(db); err != nil {
		return db, err
	}
	return db, nil
}

func createSchema(db *sql.DB) error {
	// stores configuration items
	if err := exec(db, `CREATE TABLE item (
        "key"        VARCHAR(100) NOT NULL PRIMARY KEY,
        "type"       VARCHAR(100) NOT NULL,
        "value"      TEXT NOT NULL,
		"updated"    INTEGER NOT NULL
	    );`); err != nil {
		return err
	}
	// stores tags for configuration items
	if err := exec(db, `CREATE TABLE tag (
        "item_key"        VARCHAR(100) NOT NULL,
        "name"            VARCHAR(100) NOT NULL,
        "value"           VARCHAR(100),
        PRIMARY KEY ("item_key", "name")
	    );`); err != nil {
		return err
	}
	// stores associations between configuration items
	if err := exec(db, `CREATE TABLE link (
        "from_key"        VARCHAR(100) NOT NULL,
        "to_key"          VARCHAR(100) NOT NULL,
        PRIMARY KEY ("from_key", "to_key")
	    );`); err != nil {
		return err
	}
	// stores json schemas for validation
	if err := exec(db, `CREATE TABLE type (
        "key"        VARCHAR(100) NOT NULL PRIMARY KEY,
        "schema"     TEXT NOT NULL
	    );`); err != nil {
		return err
	}
	return nil
}

func exec(db *sql.DB, stmt string) error {
	statement, err := db.Prepare(stmt)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	if err != nil {
		return err
	}
	return nil
}

func toInSqlTags(tag []string) string {
	var out strings.Builder
	for i, t := range tag {
		out.WriteString(fmt.Sprintf("'%s'", t))
		if i < len(tag)-1 {
			out.WriteString(",")
		}
	}
	return fmt.Sprintf(" IN (%s);", out.String())
}
