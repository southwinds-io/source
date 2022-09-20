/*
  Source Configuration Service
  © 2022 Southwinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package service

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
)

var db *DataBase

const homeFolder = ".source"

// ensures the database exists
func init() {
	dbPath := getPath()
	d, err := newDb(dbPath)
	if err != nil {
		log.Fatalf("cannot create database: %s", err)
		panic(err)
	}
	db = d
	log.Printf("using '%s' database path\n", dbPath)
}

func getPath() string {
	dbPath := os.Getenv("SOURCE_DATA_PATH")
	if len(dbPath) == 0 {
		currentUser, err := user.Current()
		if err != nil {
			dbPath = filepath.Join("opt", homeFolder)
		}
		dbPath, _ = filepath.Abs(filepath.Join(currentUser.HomeDir, homeFolder))
	}
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		if err = os.MkdirAll(dbPath, 0755); err != nil {
			panic(err)
		}
	}
	return dbPath
}
