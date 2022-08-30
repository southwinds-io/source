/*
  Source Configuration Service
  © 2022 Southwinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package service

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	h "southwinds.dev/http"
	"southwinds.dev/source/client"
	_ "southwinds.dev/source/docs"
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
// @Tags Health
// @Router /ready [get]
// @Accepts json
// @Produce json
// @Failure 500 {string} the service is not ready
// @Success 200 {string} the service is ready
func ReadyHandler(w http.ResponseWriter, r *http.Request) {
	_, err := db.getItem(uuid.NewString())
	if err != nil && err != ErrNotFound {
		log.Printf("service not ready: %s\n", err)
		h.Err(w, http.StatusInternalServerError, fmt.Sprintf("service not ready: %s\n", err))
		return
	}
	w.WriteHeader(http.StatusOK)
}

// SetTypeHandler
// @Summary Set the validation for an item type
// @Description Set the json schema to validate an item of the specific type
// @Tags Validation
// @Router /type [put]
// @Param schema body src.TT true "the json schema to apply to the item type and an example prototype"
// @Accepts json
// @Produce json
// @Failure 400 {string} the request is not correct
// @Failure 500 {string} there was an unexpected error processing the request
// @Success 204 {string} the request was successful
func SetTypeHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("cannot read request body: %s\n", err)
		h.Err(w, http.StatusBadRequest, fmt.Sprintf("cannot read request body: %s\n", err))
		return
	}
	t := new(src.TT)
	err = json.Unmarshal(body, t)
	if err != nil {
		log.Printf("cannot unmarshal request body: %s\n", err)
		h.Err(w, http.StatusBadRequest, fmt.Sprintf("cannot unmarshal request body: %s\n", err))
		return
	}
	err = db.setTypeFromString(t.Key, t.Schema, t.Proto)
	if err != nil {
		log.Printf("cannot set type: %s\n", err)
		h.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot set type: %s\n", err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetTypeHandler
// @Summary Get the json schema for a configuration item
// @Description Get the json schema for a configuration item
// @Tags Validation
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
	typeInfo, err := db.getTypeInfo(key)
	if err != nil {
		if err == ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Printf("cannot get schema: %s\n", err)
		h.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot get schema: %s\n", err))
		return
	}
	h.Write(w, r, typeInfo)
}

// GetTypesHandler
// @Summary Get all the item types
// @Description Get all the item types
// @Tags Validation
// @Router /type [get]
// @Produce json
// @Failure 500 {string} there was an unexpected error processing the request
// @Success 200 {string} the request was successful
func GetTypesHandler(w http.ResponseWriter, r *http.Request) {
	types, err := db.getTypes()
	if err != nil {
		log.Printf("cannot get types: %s\n", err)
		h.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot get types: %s\n", err))
		return
	}
	h.Write(w, r, types)
}

// DeleteTypeHandler
// @Summary Delete a configuration type
// @Description Delete a configuration type
// @Tags Validation
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
		h.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot delete configuration: %s\n", err))
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
		h.Err(w, http.StatusBadRequest, fmt.Sprintf("cannot read request body: %s\n", err))
		return
	}
	err, isValidationError := db.SetItem(key, itemType, string(body[:]))
	if err != nil {
		if err == ErrInvalidItemType || isValidationError {
			log.Printf("cannot set item: %s\n", err)
			h.Err(w, http.StatusBadRequest, fmt.Sprintf("cannot set item: %s\n", err))
			return
		}
		log.Printf("cannot set item: %s\n", err)
		h.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot set item: %s\n", err))
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
	item, err := db.getItem(key)
	if err != nil {
		if err == ErrNotFound {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Printf("cannot get configuration: %s\n", err)
		h.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot get configuration: %s\n", err))
		return
	}
	h.Write(w, r, item)
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
	items, err := db.getItems()
	if err != nil {
		log.Printf("cannot get types: %s\n", err)
		h.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot get types: %s\n", err))
		return
	}
	h.Write(w, r, items)
}

// GetTaggedItemsHandler
// @Summary Get all the configurations that have the specified tags
// @Description Get all the configurations that have the specified tags
// @Tags Items
// @Router /item/tag/{tags} [get]
// @Param tags path string true "a pipe separated list of tags (e.g. tag1|tag2|tag3) where tag is the tag name, not the value"
// @Produce json
// @Failure 500 {string} there was an unexpected error processing the request
// @Success 200 {string} the request was successful
func GetTaggedItemsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tags := vars["tags"]
	tagList := strings.Split(tags, "|")
	items, err := db.getTaggedItems(tagList...)
	if err != nil {
		log.Printf("cannot get tagged items: %s\n", err)
		h.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot get tagged items: %s\n", err))
		return
	}
	h.Write(w, r, items)
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
		h.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot delete configuration: %s\n", err))
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GetChildrenHandler
// @Summary Get the children linked to a configuration
// @Description Get the children linked to a configuration
// @Tags Items
// @Router /item/{key}/children [get]
// @Param key path string true "the key for the item having the children"
// @Produce json
// @Failure 500 {string} there was an unexpected error processing the request
// @Success 200 {string} the request was successful
func GetChildrenHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	children, err := db.getChildren(key)
	if err != nil {
		log.Printf("cannot get types: %s\n", err)
		h.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot get types: %s\n", err))
		return
	}
	h.Write(w, r, children)
}

// GetParentsHandler
// @Summary Get the parents linked to a configuration
// @Description Get the parents linked to a configuration
// @Tags Items
// @Router /item/{key}/parents [get]
// @Param key path string true "the key for the item having the children"
// @Produce json
// @Failure 500 {string} there was an unexpected error processing the request
// @Success 200 {string} the request was successful
func GetParentsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	children, err := db.getParents(key)
	if err != nil {
		log.Printf("cannot get types: %s\n", err)
		h.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot get types: %s\n", err))
		return
	}
	h.Write(w, r, children)
}

// SetTagHandler
// @Summary Tag an item
// @Description Tag the item identified by its key with a name
// @Tags Tagging
// @Router /item/{key}/tag/{name-value} [put]
// @Param key path string true "the key for the item to tag"
// @Param name-value path string true "the name / value of the tag in the format '{name}|{value}'. It is possible to have tags with no value, in which case the expression should be just '{name}'"
// @Accepts json
// @Produce json
// @Failure 400 {string} the request is not correct
// @Failure 500 {string} there was an unexpected error processing the request
// @Success 204 {string} the request was successful
func SetTagHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	nv := vars["name-value"]
	if len(nv) == 0 {
		log.Printf("missing tag name-value\n")
		h.Err(w, http.StatusBadRequest, "missing tag name-value\n")
		return
	}
	parts := strings.Split(nv, "|")
	var name, value string
	name = parts[0]
	if len(parts) > 1 {
		value = parts[1]
	}
	err := db.tagValue(key, name, value)
	if err != nil {
		log.Printf("cannot tag configuration: %s\n", err)
		h.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot tag configuration: %s\n", err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeleteTagHandler
// @Summary Untag a configuration
// @Description Delete the Tag associated with an item
// @Tags Tagging
// @Router /item/{key}/tag/{name} [delete]
// @Param key path string true "the key for the item to untag"
// @Param name path string true "the name of the tag to delete"
// @Accepts json
// @Produce json
// @Failure 500 {string} there was an unexpected error processing the request
// @Success 204 {string} the request was successful
func DeleteTagHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	name := vars["name"]
	err := db.untag(key, name)
	if err != nil {
		log.Printf("cannot tag configuration: %s\n", err)
		h.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot tag configuration: %s\n", err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetTagsHandler
// @Summary Get all tags for a configuration
// @Description Get all tags for a configuration
// @Tags Tagging
// @Router /item/{key}/tag [get]
// @Param key path string true "the key for the item whose tags are to be retrieved"
// @Produce json
// @Failure 500 {string} there was an unexpected error processing the request
// @Success 200 {string} the request was successful
func GetTagsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	tags, err := db.getTags(key)
	if err != nil {
		log.Printf("cannot get tags: %s\n", err)
		h.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot get tags: %s\n", err))
		return
	}
	h.Write(w, r, tags)
}

// GetAllTagsHandler
// @Summary Get all tags for a all configurations
// @Description Get all tags for a all configurations
// @Tags Tagging
// @Router /tag [get]
// @Produce json
// @Failure 500 {string} there was an unexpected error processing the request
// @Success 200 {string} the request was successful
func GetAllTagsHandler(w http.ResponseWriter, r *http.Request) {
	tags, err := db.getAllTags()
	if err != nil {
		log.Printf("cannot get tags: %s\n", err)
		h.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot get tags: %s\n", err))
		return
	}
	h.Write(w, r, tags)
}

// GetTagValueHandler
// @Summary Get a tag value
// @Description Get the value of a tag for a configuration
// @Tags Tagging
// @Router /item/{key}/tag/{name} [get]
// @Param key path string true "the key for the item having the tag"
// @Param name path string true "the name of the tag having the value that has to be retrieved"
// @Produce json
// @Failure 500 {string} there was an unexpected error processing the request
// @Success 200 {string} the request was successful
func GetTagValueHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	name := vars["name"]
	tags, err := db.getTags(key)
	if err != nil {
		log.Printf("cannot get types: %s\n", err)
		h.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot get types: %s\n", err))
		return
	}
	for _, tag := range tags {
		if strings.EqualFold(tag.Name, name) {
			w.Write([]byte(tag.Value))
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

// LinkHandler
// @Summary Link two configurations
// @Description Link two configurations
// @Tags Linking
// @Router /link/{from-key}/to/{to-key} [put]
// @Param from-key path string true "the key for the first configuration to link"
// @Param to-key path string true "the key for the second configuration to link"
// @Accepts json
// @Produce json
// @Failure 500 {string} there was an unexpected error processing the request
// @Success 204 {string} the request was successful
func LinkHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	from := vars["from-key"]
	to := vars["to-key"]
	err := db.Link(from, to)
	if err != nil {
		log.Printf("cannot link configurations: %s\n", err)
		h.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot link configurations: %s\n", err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UnlinkHandler
// @Summary Unlink two configurations
// @Description Unlink two configurations
// @Tags Linking
// @Router /link/{from-key}/to/{to-key} [delete]
// @Param from-key path string true "the key for the first configuration to unlink"
// @Param to-key path string true "the key for the second configuration to unlink"
// @Accepts json
// @Produce json
// @Failure 500 {string} there was an unexpected error processing the request
// @Success 204 {string} the request was successful
func UnlinkHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	from := vars["from-key"]
	to := vars["to-key"]
	err := db.unLink(from, to)
	if err != nil {
		log.Printf("cannot unlink configurations: %s\n", err)
		h.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot unlink configurations: %s\n", err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetLinksHandler
// @Summary Get all configuration links
// @Description Get all configuration links
// @Tags Linking
// @Router /link [get]
// @Accepts json
// @Produce json
// @Failure 500 {string} there was an unexpected error processing the request
// @Success 200 {string} the request was successful
func GetLinksHandler(w http.ResponseWriter, r *http.Request) {
	links, err := db.getLinks()
	if err != nil {
		log.Printf("cannot retireve configuration links: %s\n", err)
		h.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot retireve configuration links: %s\n", err))
		return
	}
	h.Write(w, r, links)
}

// DeleteLinksHandler
// @Summary Delete all configuration links
// @Description Delete all configuration links
// @Tags Linking
// @Router /link [delete]
// @Accepts json
// @Produce json
// @Failure 500 {string} there was an unexpected error processing the request
// @Success 200 {string} the request was successful
func DeleteLinksHandler(w http.ResponseWriter, r *http.Request) {
	err := db.deleteLinks()
	if err != nil {
		log.Printf("cannot retireve configuration links: %s\n", err)
		h.Err(w, http.StatusInternalServerError, fmt.Sprintf("cannot retireve configuration links: %s\n", err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
