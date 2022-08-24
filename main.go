/*
  Onix Config Manager - Source
  © 2022 southwinds.io - All rights reserved
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package main

import (
	"fmt"
	"github.com/gatblau/onix/oxlib/httpserver"
	"github.com/gorilla/mux"
	"github.com/southwinds-io/source/service"
	"net/http"
)

func main() {
	fmt.Printf(`
+++++++++| ONIX CONFIG MANAGER |+++++++++
|      ___  ___  _   _ _ __ ___ ___     |
|     / __|/ _ \| | | | '__/ __/ _ \    |
|     \__ \ (_) | |_| | | | (_|  __/    |
|     |___/\___/ \__,_|_|  \___\___|    |
|                                       |
+++++++| configuration database |++++++++
%s
`, service.Version)
	server := httpserver.New("SOURCE")
	server.Http = func(router *mux.Router) {
		router.HandleFunc("/ready", service.ReadyHandler).Methods(http.MethodGet)
		// validation
		router.HandleFunc("/type/{key}", service.SetTypeHandler).Methods(http.MethodPut)
		router.HandleFunc("/type/{key}", service.GetTypeHandler).Methods(http.MethodGet)
		router.HandleFunc("/type", service.GetTypesHandler).Methods(http.MethodGet)
		router.HandleFunc("/type", service.DeleteTypeHandler).Methods(http.MethodDelete)
		// configurations
		router.HandleFunc("/item/{key}", service.SetItemHandler).Methods(http.MethodPut)
		router.HandleFunc("/item/{key}", service.GetItemHandler).Methods(http.MethodGet)
		router.HandleFunc("/item", service.GetItemsHandler).Methods(http.MethodGet)
		router.HandleFunc("/item", service.DeleteItemHandler).Methods(http.MethodDelete)
		router.HandleFunc("/item/{key}/children", service.GetChildrenHandler).Methods(http.MethodGet)
		router.HandleFunc("/item/{key}/parents", service.GetParentsHandler).Methods(http.MethodGet)
		router.HandleFunc("/item/tag/{tags}", service.GetTaggedItemsHandler).Methods(http.MethodGet)
		// tagging
		router.HandleFunc("/item/{key}/tag/{name-value}", service.SetTagHandler).Methods(http.MethodPut)
		router.HandleFunc("/item/{key}/tag/{name}", service.DeleteTagHandler).Methods(http.MethodDelete)
		router.HandleFunc("/item/{key}/tag", service.GetTagsHandler).Methods(http.MethodGet)
		router.HandleFunc("/item/{key}/tag/{name}", service.GetTagValueHandler).Methods(http.MethodGet)
		router.HandleFunc("/tag", service.GetAllTagsHandler).Methods(http.MethodGet)
		// linking
		router.HandleFunc("/link/{from-key}/to/{to-key}", service.LinkHandler).Methods(http.MethodPut)
		router.HandleFunc("/unlink/{from-key}/to/{to-key}", service.UnlinkHandler).Methods(http.MethodDelete)
		router.HandleFunc("/link", service.GetLinksHandler).Methods(http.MethodGet)
		router.HandleFunc("/link", service.DeleteLinksHandler).Methods(http.MethodDelete)
	}
	server.Serve()
}
