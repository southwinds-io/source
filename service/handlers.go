/*
  Onix Config Manager - Source
  © 2022 southwinds.io - All rights reserved
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package service

import (
	"fmt"
	"github.com/gatblau/onix/oxlib/httpserver"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/southwinds-io/source/cdb"
	_ "github.com/southwinds-io/source/docs"
	"io"
	"log"
	"net/http"
	"strings"
)

// @title Onix Source
// @version 1.0
// @description Ultra lightweight configuration data service
// @contact.name SouthWinds Tech Ltd
// @contact.url https://southwinds.io/
// @contact.email info@southwinds.io

// ReadyHandler
// @Summary Check service readiness
// @Description Check any relevant backends are online and healthy.
// @Tags General
// @Router /ready [get]
// @Accepts json
// @Produce json
// @Failure 500 {string} the service is not ready
// @Success 200 {string} the service is ready
func ReadyHandler(w http.ResponseWriter, r *http.Request) {
	_, err := db.GetItem(uuid.NewString())
	if err != nil && !strings.Contains(err.Error(), "sql: no rows") {
		log.Printf("service not ready: %s\n", err)
		httpserver.Err(w, http.StatusInternalServerError, fmt.Sprintf("service not ready: %s\n", err))
		return
	}
	w.WriteHeader(http.StatusOK)
}

// SetTypeHandler
// @Summary Set the validation for an item type
// @Description Set the json schema to validate an item of the specific type
// @Tags Types
// @Router /type/{key} [put]
// @Param key path string true "the key for the type to set"
// @Param schema body string true "the json schema to apply to the item type"
// @Accepts json
// @Produce json
// @Failure 400 {string} the request is not correct
// @Failure 500 {string} there was an unexpected error processing the request
// @Success 204 {string} the request was successful
func SetTypeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("cannot read request body: %s\n", err)
		httpserver.Err(w, http.StatusBadRequest, fmt.Sprintf("cannot read request body: %s\n", err))
		return
	}
	err = db.SetType(key, string(body[:]))
	if err != nil {
		log.Printf("cannot set type: %s\n", err)
		httpserver.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot set type: %s\n", err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetTypeHandler
// @Summary Get the json schema for a configuration item
// @Description Get the json schema for a configuration item
// @Tags Types
// @Router /type/{key} [get]
// @Param key path string true "the key for the type of configuration item json schema"
// @Produce json
// @Failure 400 {string} the request is not correct
// @Failure 404 {string} json schema not found
// @Failure 500 {string} there was an unexpected error processing the request
// @Success 200 {string} the request was successful
func GetTypeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	schema, err := db.GetSchema(key)
	if err != nil {
		if err == cdb.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Printf("cannot get schema: %s\n", err)
		httpserver.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot get schema: %s\n", err))
		return
	}
	w.Write([]byte(schema))
}

// GetTypesHandler
// @Summary Get all the item types
// @Description Get all the item types
// @Tags Types
// @Router /type [get]
// @Produce json
// @Failure 500 {string} there was an unexpected error processing the request
// @Success 200 {string} the request was successful
func GetTypesHandler(w http.ResponseWriter, r *http.Request) {
	types, err := db.GetTypes()
	if err != nil {
		log.Printf("cannot get types: %s\n", err)
		httpserver.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot get types: %s\n", err))
		return
	}
	httpserver.Write(w, r, types)
}

// DeleteTypeHandler
// @Summary Delete a configuration type
// @Description Delete a configuration type
// @Tags Types
// @Router /type/{key} [delete]
// @Param key path string true "the key for the configuration type to delete"
// @Produce json
// @Failure 400 {string} the request is not correct
// @Failure 500 {string} there was an unexpected error processing the request
// @Success 200 {string} the request was successful
func DeleteTypeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	err := db.DeleteType(key)
	if err != nil {
		log.Printf("cannot delete configuration: %s\n", err)
		httpserver.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot delete configuration: %s\n", err))
		return
	}
	w.WriteHeader(http.StatusOK)
}

// SetItemHandler
// @Summary Set the value of a configuration item
// @Description Set value of a configuration item
// @Tags Items
// @Router /item/{key} [put]
// @Param key path string true "the key for the configuration item to set"
// @Param schema body string true "the json based configuration"
// @Param Source-Type header string false "the key that defines the type of item for validation purposes. If not specified, no validation is performed."
// @Accepts json
// @Produce json
// @Failure 400 {string} the request is not correct
// @Failure 500 {string} there was an unexpected error processing the request
// @Success 204 {string} the request was successful
func SetItemHandler(w http.ResponseWriter, r *http.Request) {
	itemType := r.Header.Get("Source-Type")
	vars := mux.Vars(r)
	key := vars["key"]
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("cannot read request body: %s\n", err)
		httpserver.Err(w, http.StatusBadRequest, fmt.Sprintf("cannot read request body: %s\n", err))
		return
	}
	err = db.SetItem(key, itemType, string(body[:]), len(itemType) > 0)
	if err != nil {
		if err == cdb.ErrInvalidItemType {
			log.Printf("cannot set item: %s\n", err)
			httpserver.Err(w, http.StatusBadRequest, fmt.Sprintf("cannot set item: %s\n", err))
			return
		}
		log.Printf("cannot set item: %s\n", err)
		httpserver.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot set item: %s\n", err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetItemHandler
// @Summary Get the value of a configuration item
// @Description Get value of a configuration item
// @Tags Items
// @Router /item/{key} [get]
// @Param key path string true "the key for the configuration item to get"
// @Produce json
// @Failure 400 {string} the request is not correct
// @Failure 404 {string} configuration not found
// @Failure 500 {string} there was an unexpected error processing the request
// @Success 200 {string} the request was successful
func GetItemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	item, err := db.GetItem(key)
	if err != nil {
		if err == cdb.ErrNotFound {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Printf("cannot get configuration: %s\n", err)
		httpserver.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot get configuration: %s\n", err))
		return
	}
	httpserver.Write(w, r, item)
}

// GetItemsHandler
// @Summary Get all the configurations
// @Description Get all the configurations
// @Tags Items
// @Router /item [get]
// @Produce json
// @Failure 500 {string} there was an unexpected error processing the request
// @Success 200 {string} the request was successful
func GetItemsHandler(w http.ResponseWriter, r *http.Request) {
	items, err := db.GetItems()
	if err != nil {
		log.Printf("cannot get types: %s\n", err)
		httpserver.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot get types: %s\n", err))
		return
	}
	httpserver.Write(w, r, items)
}

// DeleteItemHandler
// @Summary Delete a configuration item
// @Description Delete a configuration item
// @Tags Items
// @Router /item/{key} [delete]
// @Param key path string true "the key for the configuration item to delete"
// @Produce json
// @Failure 400 {string} the request is not correct
// @Failure 500 {string} there was an unexpected error processing the request
// @Success 200 {string} the request was successful
func DeleteItemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	err := db.DeleteItem(key)
	if err != nil {
		log.Printf("cannot delete configuration: %s\n", err)
		httpserver.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot delete configuration: %s\n", err))
		return
	}
	w.WriteHeader(http.StatusOK)
}
