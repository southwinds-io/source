/*
  Source Configuration Service
  © 2022 Southwinds Tech Ltd - www.southwinds.io
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package service

import (
	"fmt"
	"testing"
)

type testV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func TestLifecycle(t *testing.T) {
	// create db
	d, err := newDb(".")
	if err != nil {
		t.Fatalf(err.Error())
	}
	// set the type by inferring the json schema from the passed in struct
	err = d.setTypeFromStruct("kv", []testV{
		{
			Key:   "my-key",
			Value: "my-value",
		},
	})
	if err != nil {
		t.Fatalf(err.Error())
	}
	// set item using kv schema
	err, _ = d.SetItem("test", "kv", `[
    {
      "key": "name1",
      "value": "value1"
    },
    {
      "key": "name2",
      "value": "value2"
    }
]`)
	if err != nil {
		t.Fatalf(err.Error())
	}
	// add tag name & value to item
	err = d.tagValue("test", "dev", "xyz")
	if err != nil {
		t.Fatalf(err.Error())
	}
	// add tag name only to item
	err = d.tag("test", "init")
	if err != nil {
		t.Fatalf(err.Error())
	}
	// set another item using a serializable struct
	err, _ = d.SetItem("test2", "kv", []testV{
		{
			Key:   "key1",
			Value: "value1",
		},
		{
			Key:   "key2",
			Value: "value2",
		},
		{
			Key:   "key3",
			Value: "value3",
		},
	})
	// tag it
	err = d.tag("test2", "init")
	if err != nil {
		t.Fatalf(err.Error())
	}
	// link the two items
	err = d.Link("test", "test2")
	// get items with tags
	tagged, err := d.getTaggedItems("dev")
	if err != nil {
		t.Fatalf(err.Error())
	}
	for _, i := range tagged {
		fmt.Println(i.Value)
	}
	// get tags for an item
	tags, err := d.getTags("test")
	if err != nil {
		t.Fatalf(err.Error())
	}
	for _, tag := range tags {
		fmt.Printf("%s=%s\n", tag.Name, tag.Value)
	}
	// get a specific item
	i, err := d.getItem("test2")
	if err != nil {
		t.Fatalf(err.Error())
	}
	fmt.Println(i.Updated)
	// delete type
	_ = d.DeleteType("kv")
	// delete item
	_ = d.DeleteItem("test")
	_ = d.DeleteItem("test2")
}
