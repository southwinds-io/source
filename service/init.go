/*
  Onix Config Manager - Source
  © 2022 southwinds.io - All rights reserved
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package service

import (
	"github.com/southwinds-io/source/cdb"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

var db *cdb.DataBase

const homeFolder = ".source"

// ensures the database exists
func init() {
	dbPath := os.Getenv("SW_SRC_DATA_PATH")
	if len(dbPath) == 0 {
		currentUser, err := user.Current()
		if err != nil {
			dbPath = filepath.Join("opt", homeFolder)
		}
		dbPath = currentUser.HomeDir
	}
	dbPath, _ = filepath.Abs(filepath.Join(dbPath, homeFolder))
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		if err = os.MkdirAll(dbPath, 0755); err != nil {
			panic(err)
		}
	}
	d, err := cdb.New(dbPath)
	if err != nil {
		panic(err)
	}
	db = d
	log.Printf("using '%s' database path\n", dbPath)
}
