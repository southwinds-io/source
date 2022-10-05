# Source

## A lightweight configuration database

Source is a lightweight database designed to store configuration items providing minimal 
installation and database maintenance.

It uses Sqlite as a backend and allows to:

1. store any configuration in json format as an encrypted BLOB
2. identify configurations using natural keys of your choice
3. validate the configuration using predefined json schemas (no need to create schemas, they are inferred from 
   configuration prototypes)
4. optionally attach tags to configuration (tags can have a name only or a name and a value)
6. optionally associate configurations via links

### Launching the service

```bash
# start service in a docker container
docker run \
   --name src \
   --restart=always \
   -d \
   -p 8999:8080 \
   -e ART_PACKAGE_NAME="app/source" \
   -e OX_HTTP_USER="USER-NAME-HERE" \
   -e OX_HTTP_PWD="USER-PASSWORD-HERE" \
   -e SOURCE_DATA_PATH="volume_0" \
   quay.io/artisan/app-run:ubi-minimal
       
# launch Open API in a browser
python -mwebbrowser http://localhost:8999/api/
```

### Using the go client

[See here](src/readme.md).

