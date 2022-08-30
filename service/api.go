/*
  Source Configuration Service
  © 2022 Southwinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/invopop/jsonschema"
	schemaValidation "github.com/qri-io/jsonschema"
	"io"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
	"southwinds.dev/source/client"
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

// newDb create a new configuration database on the specified path
func newDb(path string) (*DataBase, error) {
	var err error
	m := new(DataBase)
	if m.db, err = getDb(path); err != nil {
		return nil, err
	}
	return m, nil
}

// setTypeFromString set the json schema for an item type using a json string representation of the schema
func (d *DataBase) setTypeFromString(key string, schema, proto []byte) error {
	stmt := `INSERT INTO type(key, schema, proto) VALUES(?, ?, ?) ON CONFLICT(key) DO UPDATE SET schema = excluded.schema, proto = excluded.proto;`
	statement, err := d.db.Prepare(stmt)
	if err != nil {
		return err
	}
	_, err = statement.Exec(key, schema, proto)
	return err
}

// setTypeFromStruct set the json schema for the item type by inferring it from the passed in object
func (d *DataBase) setTypeFromStruct(key string, obj interface{}) error {
	// reflects the json schema from the specified object
	schemaObj := jsonschema.Reflect(obj)
	// marshal the object to json bytes
	schema, err := json.Marshal(schemaObj)
	if err != nil {
		return err
	}
	proto, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return d.setTypeFromString(key, schema, proto)
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
func (d *DataBase) SetItem(key, iType string, value interface{}) (error, bool) {
	if value == nil {
		return fmt.Errorf("value not provided"), false
	}
	if sv, ok := value.(string); ok {
		return d.setItemString(key, iType, sv)
	}
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return err, false
	}
	return d.setItemString(key, iType, string(valueBytes[:]))
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

// unLink remove an association between two items
func (d *DataBase) unLink(from, to string) error {
	stmt := `DELETE FROM link WHERE from_key=? AND to_key=?; `
	statement, err := d.db.Prepare(stmt)
	if err != nil {
		return err
	}
	_, err = statement.Exec(from, to)
	return err
}

// tag an item with  a name only (value is empty)
func (d *DataBase) tag(key, name string) error {
	return d.tagValue(key, name, "")
}

// tagValue tag an item with a name and a value
func (d *DataBase) tagValue(key, name, value string) error {
	stmt := `INSERT INTO tag(item_key, name, value) VALUES(?, ?, ?) ON CONFLICT(item_key, name) DO UPDATE SET value = excluded.value;`
	statement, err := d.db.Prepare(stmt)
	if err != nil {
		return err
	}
	_, err = statement.Exec(key, name, value)
	return err
}

// untag a configuration
func (d *DataBase) untag(key string, name string) error {
	stmt := `DELETE FROM tag WHERE item_key=? AND name=?; `
	statement, err := d.db.Prepare(stmt)
	if err != nil {
		return err
	}
	_, err = statement.Exec(key, name)
	return err
}

func (d *DataBase) deleteLinks() interface{} {
	stmt := `DELETE FROM link;`
	statement, err := d.db.Prepare(stmt)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	return err
}

// getItem get an item by key
func (d *DataBase) getItem(key string) (*src.I, error) {
	row := d.db.QueryRow(`SELECT type, value, updated FROM item WHERE key=?;`, key)
	var (
		itype   string
		value   []byte
		updated sql.NullInt64
	)
	err := row.Scan(&itype, &value, &updated)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, ErrNotFound
		}
		return nil, err
	}
	vv, decErr := decrypt(value)
	if decErr != nil {
		return nil, err
	}
	return &src.I{
		Key:     key,
		Type:    itype,
		Value:   vv,
		Updated: time.Unix(0, updated.Int64).UTC(),
	}, nil
}

// getTaggedItems get the items with the specified tag names
func (d *DataBase) getTaggedItems(tags ...string) ([]src.I, error) {
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
		key, iType string
		value      []byte
		updated    sql.NullInt64
	)
	var items []src.I
	for row.Next() {
		err = row.Scan(&key, &iType, &value, &updated)
		if err != nil {
			return nil, err
		}
		vv, decErr := decrypt(value)
		if decErr != nil {
			return nil, err
		}
		items = append(items, src.I{
			Key:     key,
			Type:    iType,
			Value:   vv,
			Updated: time.Unix(0, updated.Int64).UTC(),
		})
	}
	return items, nil
}

// getTags get the tags (name & value) of an item
func (d *DataBase) getTags(key string) ([]src.T, error) {
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
	var tags []src.T
	for row.Next() {
		err = row.Scan(&name, &value)
		if err != nil {
			return nil, err
		}
		tags = append(tags, src.T{
			Name:  name,
			Value: value,
		})
	}
	return tags, nil
}

func (d *DataBase) getAllTags() ([]src.T, error) {
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
	var tags []src.T
	for row.Next() {
		err = row.Scan(&itemKey, &name, &value)
		if err != nil {
			return nil, err
		}
		tags = append(tags, src.T{
			ItemKey: itemKey,
			Name:    name,
			Value:   value,
		})
	}
	return tags, nil
}

// getChildren get the child items linked to a specified item
func (d *DataBase) getChildren(parentKey string) ([]src.I, error) {
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
		key, iType string
		value      []byte
		updated    sql.NullInt64
	)
	var items []src.I
	for row.Next() {
		err = row.Scan(&key, &iType, &value, &updated)
		if err != nil {
			return nil, err
		}
		vv, decErr := decrypt(value)
		if decErr != nil {
			return nil, err
		}
		items = append(items, src.I{
			Key:     key,
			Type:    iType,
			Value:   vv,
			Updated: time.Unix(0, updated.Int64).UTC(),
		})
	}
	return items, nil
}

// getParents get the parent items linked to a specified item
func (d *DataBase) getParents(childKey string) ([]src.I, error) {
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
		key, iType string
		value      []byte
		updated    sql.NullInt64
	)
	var items []src.I
	for row.Next() {
		err = row.Scan(&key, &iType, &value, &updated)
		if err != nil {
			return nil, err
		}
		vv, decErr := decrypt(value)
		if decErr != nil {
			return nil, err
		}
		items = append(items, src.I{
			Key:     key,
			Type:    iType,
			Value:   vv,
			Updated: time.Unix(0, updated.Int64).UTC(),
		})
	}
	return items, nil
}

func (d *DataBase) getItems() ([]src.I, error) {
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
		key, iType string
		value      []byte
		updated    sql.NullInt64
		items      []src.I
	)
	for row.Next() {
		err = row.Scan(&key, &iType, &value, &updated)
		if err != nil {
			if strings.Contains(err.Error(), "no rows") {
				return nil, ErrNotFound
			}
			return nil, err
		}
		vv, decErr := decrypt(value)
		if decErr != nil {
			return nil, decErr
		}
		items = append(items, src.I{
			Key:     key,
			Type:    iType,
			Value:   vv,
			Updated: time.Unix(0, updated.Int64).UTC(),
		})
	}
	return items, nil
}

func (d *DataBase) getLinks() ([]src.L, error) {
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
		links    []src.L
	)
	for row.Next() {
		err = row.Scan(&from, &to)
		if err != nil {
			if strings.Contains(err.Error(), "no rows") {
				return nil, ErrNotFound
			}
			return nil, err
		}
		links = append(links, src.L{
			From: from,
			To:   to,
		})
	}
	return links, nil
}

func (d *DataBase) getTypes() ([]src.TT, error) {
	row, err := d.db.Query(`SELECT key, schema, proto FROM type;`)
	if err != nil {
		return nil, err
	}
	defer func(row *sql.Rows) {
		err = row.Close()
		if err != nil {
			fmt.Printf("cannot close query row: %s\n", err)
		}
	}(row)
	var key string
	var schema, proto []byte
	var types []src.TT
	for row.Next() {
		err = row.Scan(&key, &schema, &proto)
		if err != nil {
			if strings.Contains(err.Error(), "no rows") {
				return nil, ErrNotFound
			}
			return nil, err
		}
		types = append(types, src.TT{
			Key:    key,
			Schema: schema,
			Proto:  proto,
		})
	}
	return types, nil
}

func (d *DataBase) getTypeInfo(key string) (*src.TT, error) {
	row := d.db.QueryRow(`SELECT schema, proto FROM type WHERE key = ?;`, key)
	var schema, proto []byte
	err := row.Scan(&schema, &proto)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &src.TT{Schema: schema, Proto: proto}, nil
}

func (d *DataBase) setItemString(key, iType, value string) (error, bool) {
	// get the schema for the type
	typeInfo, err := d.getTypeInfo(iType)
	if err != nil {
		return err, false
	}
	// validates only if a schema has been defined
	if typeInfo == nil {
		return ErrInvalidItemType, false
	}

	ctx := context.Background()
	rs := &schemaValidation.Schema{}
	if err = json.Unmarshal(typeInfo.Schema, rs); err != nil {
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

	stmt := `INSERT INTO item(key, type, value, updated) VALUES(?, ?, ?, ?) ON CONFLICT(key) DO UPDATE SET type = excluded.type, value = excluded.value, updated = excluded.updated;`
	statement, err := d.db.Prepare(stmt)
	if err != nil {
		return err, false
	}
	vv, encErr := encrypt([]byte(value))
	if encErr != nil {
		return err, false
	}
	_, err = statement.Exec(key, iType, vv, time.Now().UTC().UnixNano())
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
        "value"      BLOB NOT NULL,
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
        "schema"     BLOB NOT NULL,
        "proto"      BLOB NOT NULL
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

func encrypt(input []byte) ([]byte, error) {
	key := sha256.Sum256([]byte(K))
	block, _ := aes.NewCipher(key[:])
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, input, nil), nil
}

func decrypt(cipherBytes []byte) ([]byte, error) {
	key := sha256.Sum256([]byte(K))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := cipherBytes[:gcm.NonceSize()]
	cipherBytes = cipherBytes[gcm.NonceSize():]
	bytes, err := gcm.Open(nil, nonce, cipherBytes, nil)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

const K = "VCh4IbXtYuFYnNa4L0xC1F49gZKZaNgrqJPMvNmac5T1W3zH0e"
