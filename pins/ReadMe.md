1) ADD curl -i -X POST -d 'board=0&title=exampleTitle&description=exampleDescription&image_link=/example/link' localhost:8080/pin/
2) GET curl -X GET  localhost:8080/pins/0
3) DELETE curl -X DELETE  localhost:8080/pins/0


