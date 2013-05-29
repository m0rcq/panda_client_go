======================================================

panda_client_go: interface to Panda's API in Google Go

date: 2013-05-29

author: jack argirov (m0rcq@argirov.co.uk)

licence: MIT

======================================================

Implementation of pandastream.com's API in Go, handles:

* API authentication
* GET/POST/PUT/DELETE against the RESTful resources as Videos/Encodings/Profiles/Clouds/Notifications

More information about the actual API: <http://www.pandastream.com/docs/api>

Includes simple command line client in examples/client.go showing one way of using it. 

Written as a learning exercise in Go - any improvements you might see worth adding, please let me know!

73 de M0RCQ

Install
-------

go get github.com/m0rcq/panda_client_go


Testing
-------
go test github.com/m0rcq/panda_client_go

(test against signature generation included)


Build cmd line tool
-------------------
go get github.com/bitly/go-simplejson

go install github.com/m0rcq/panda_client_go/examples/panda_client_cmd


Run cmd line tool
-----------------

$GOBIN/panda_client_cmd --resource videos --cmd info
(get all video resources)

$GOBIN/panda_client_cmd --resource videos --cmd upload --video /path/to/video/file
(upload video using filepath)

$GOBIN/panda_client_cmd --resource videos --cmd delete --id 1234567890
(delete video by id)

$GOBIN/panda_client_cmd --resource encodings --cmd info
(get all encodings)

(or provide CloudId/AccessKey/SecretKey/ApiHost via command line as: --cloudId / --accessKey / --secretKey --apiHost)

Public functions
----------------

- Init(AccessKey string, SecretKey string, CloudId string, ApiHost string, ApiPort int)

- ApiURL() string
- Get(path string, data map[string]string) (string, error)
- Post(path string, data map[string]string) (string, error)
- Put(path string, data map[string]string) (string, error)
- Delete(path string, data map[string]string) (string, error)

Use in your application
-----------------------

> import "github.com/m0rcq/panda_client_go"

> var client panda.PandaApiInterface = &panda.PandaApi{}                       

> client.Init(accessKey, secretKey, cloudId, apiHost, 80) 

> client.Get("/videos.json", nil)
