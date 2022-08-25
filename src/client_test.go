/*
  Onix Config Manager - Source
  © 2022 southwinds.io - All rights reserved
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package src

import "testing"

func TestAll(t *testing.T) {
	c := New("http://127.0.0.1:8080", "", "", nil)
	// define a json schema for a configuration
	// note you do not need to create the schema, it is inferred from an empty struct in this case I am using
	// ClientOptions{}
	err := c.SetType("AAA", ClientOptions{})
	if err != nil {
		t.Fatalf(err.Error())
	}
	// set a configuration: note the actual value is any object you want, in this case I am using ClientOptions{}
	err = c.SetItem("OPT_1", "AAA", ClientOptions{
		InsecureSkipVerify: false,
		Timeout:            60,
	})
	if err != nil {
		t.Fatalf(err.Error())
	}
	// tag the item with a name and also a value
	err = c.TagItem("OPT_1", "status", "dev")
	if err != nil {
		t.Fatalf(err.Error())
	}
	// set another item
	err = c.SetItem("OPT_2", "AAA", ClientOptions{
		InsecureSkipVerify: true,
		Timeout:            120,
	})
	if err != nil {
		t.Fatalf(err.Error())
	}
	// associate the two items
	err = c.Link("OPT_1", "OPT_2")
	if err != nil {
		t.Fatalf(err.Error())
	}
}
