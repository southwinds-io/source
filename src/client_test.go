/*
  Onix Config Manager - Source
  © 2022 southwinds.io - All rights reserved
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package src

import "testing"

func TestSetType(t *testing.T) {
	c := New("http://127.0.0.1:8080", "", "", nil)
	err := c.SetType("AAA", ClientOptions{})
	if err != nil {
		t.Fatalf(err.Error())
	}
}
