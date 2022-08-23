# Source

## A lightweight and embeddable configuration database

Source is an ultra lightweight database designed to store configuration items providing a minimal 
installation and database maintenance.

It uses Sqlite as a backend and allows to:

1. store any configuration in json format
2. identify configurations using natural keys of your choice
3. optionally validates the configuration item using predefined json schemas
4. optionally attach tags to configuration items
5. tags can have a name only or a name and a value
6. optionally associate configuration items via links

## Embedded use

Import the module as follows:

```bash
$ go get github.com/southwinds-io/source/cdb
```

```go
package main 

import (
    "fmt"
    "github.com/southwinds-io/source/cdb"
)

type testV struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}

func main() {
    // create database
    db, err := cdb.New(".")
    if err != nil {
        panic(err)
    }
    // set json schema for type kv
    err = db.SetType("kv",
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
    // set a configuration item using a serializable struct
    err = db.SetItem("my-item", "kv", []testV{
        {
            Key:   "key1",
            Value: "value1",
        },
        {
            Key:   "key2",
            Value: "value2",
        },
    }, true)

    // tag it
    err = db.Tag("my-item", "init")
    if err != nil {
        panic(err)
    }

    // get the item
    i, err := db.GetItem("my-item")
    if err != nil {
        panic(err)
    }
    fmt.Println(i.Value)
}
```

For more example see [test here](cdb/api_test.go).


