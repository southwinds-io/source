/*
  Source Configuration Service
  © 2022 Southwinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	h "southwinds.dev/http"
	"southwinds.dev/source/client"
	"southwinds.dev/source/service"
)

func main() {
	fmt.Printf(`
++++++++++++++++++++++++++++++++++++++++++
|      ___  ___  _   _ _ __ ___ ___      |
|     / __|/ _ \| | | | '__/ __/ _ \     |
|     \__ \ (_) | |_| | | | (_|  __/     |
|     |___/\___/ \__,_|_|  \___\___|     |
|                                        |
++++++++| configuration service |+++++++++
%s
`, src.Version)
	server := h.New("SOURCE", src.Version)
	server.Http = func(router *mux.Router) {
		// enables basic authentication
		router.Use(server.AuthenticationMiddleware)
		router.HandleFunc("/ready", service.ReadyHandler).Methods(http.MethodGet)
		// validation
		router.HandleFunc("/type", service.SetTypeHandler).Methods(http.MethodPut)
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
		router.HandleFunc("/item/type/{type}", service.GetItemsByTypeHandler).Methods(http.MethodGet)
		router.HandleFunc("/item/oldest/type/{type}", service.GetOldestByTypeHandler).Methods(http.MethodGet)

		// tagging
		router.HandleFunc("/item/{key}/tag/{name-value}", service.SetTagHandler).Methods(http.MethodPut)
		router.HandleFunc("/item/{key}/tag/{name}", service.DeleteTagHandler).Methods(http.MethodDelete)
		router.HandleFunc("/item/{key}/tag", service.GetTagsHandler).Methods(http.MethodGet)
		router.HandleFunc("/item/{key}/tag/{name}", service.GetTagValueHandler).Methods(http.MethodGet)
		router.HandleFunc("/tag", service.GetAllTagsHandler).Methods(http.MethodGet)
		// linking
		router.HandleFunc("/link/{from-key}/to/{to-key}", service.LinkHandler).Methods(http.MethodPut)
		router.HandleFunc("/link/{from-key}/to/{to-key}", service.UnlinkHandler).Methods(http.MethodDelete)
		router.HandleFunc("/link", service.GetLinksHandler).Methods(http.MethodGet)
		router.HandleFunc("/link", service.DeleteLinksHandler).Methods(http.MethodDelete)
	}
	server.Serve()
}
