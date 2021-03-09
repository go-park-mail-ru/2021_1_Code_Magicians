1) ADD curl -i -X POST -H 'Content-Type: application/json' -d '{"boardID": 0, "title": "exampletitle", "imageLink": "example/link", "description": "exampleDescription"}' localhost:8080/pin/
2) GET curl -X GET  localhost:8080/pins/0
3) DELETE curl -X DELETE  localhost:8080/pins/0


