/*
  Onix Config Manager - Source
  © 2022 southwinds.io - All rights reserved
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package cdb

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
	d, err := New(".")
	if err != nil {
		t.Fatalf(err.Error())
	}
	// set json schema for type kv
	err = d.SetType("kv",
		`{
    "$schema": "https://json-schema.org/draft/2019-09/schema",
    "$id": "http://example.com/example.json",
    "type": "array",
    "default": [],
    "title": "Root Schema",
    "items": {
        "type": "object",
        "title": "A Schema",
        "required": [
            "key",
            "value"
        ],
        "properties": {
            "key": {
                "type": "string",
                "title": "The key Schema",
                "examples": [
                    "name1",
                    "name2"
                ]
            },
            "value": {
                "type": "string",
                "title": "The value Schema",
                "examples": [
                    "value1",
                    "value2"
                ]
            }
        }
    }
}`)
	if err != nil {
		t.Fatalf(err.Error())
	}
	// set item using kv schema
	err = d.SetItem("test", "kv", `[
    {
      "key": "name1",
      "value": "value1"
    },
    {
      "key": "name2",
      "value": "value2"
    }
]`, true)
	if err != nil {
		t.Fatalf(err.Error())
	}
	// add tag name & value to item
	err = d.TagValue("test", "dev", "xyz")
	if err != nil {
		t.Fatalf(err.Error())
	}
	// add tag name only to item
	err = d.Tag("test", "init")
	if err != nil {
		t.Fatalf(err.Error())
	}
	// set another item using a serializable struct
	err = d.SetItem("test2", "kv", []testV{
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
	}, true)
	// tag it
	err = d.Tag("test2", "init")
	if err != nil {
		t.Fatalf(err.Error())
	}
	// link the two items
	err = d.Link("test", "test2")
	// get items with tags
	tagged, err := d.GetTaggedItems("dev")
	if err != nil {
		t.Fatalf(err.Error())
	}
	for _, i := range tagged {
		fmt.Println(i.Value)
	}
	// get tags for an item
	tags, err := d.GetTags("test")
	if err != nil {
		t.Fatalf(err.Error())
	}
	for _, tag := range tags {
		fmt.Printf("%s=%s\n", tag.Name, tag.Value)
	}
	// get a specific item
	i, err := d.GetItem("test2")
	if err != nil {
		t.Fatalf(err.Error())
	}
	fmt.Println(i.Updated)
	// delete type
	d.DeleteType("kv")
	// delete item
	d.DeleteItem("test")
	d.DeleteItem("test2")
}
