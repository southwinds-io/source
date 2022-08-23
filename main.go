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
	"github.com/southwinds-io/source/service"
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
	server.Serve()
}
