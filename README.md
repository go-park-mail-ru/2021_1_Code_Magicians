This repository will contain Golang code for CodeMagicians' next big project: Pinterbest

$some_command means "run that command in console"

Instructions with docker:

- edit server/.env file: change variable LOCAL_DB_HOST to postgres
- install docker
- $docker-compose -f docker-compose.yaml up
- To stop server, press Ctrl+C

NOTE: right now to start server properly you need to run docker-compose, then stop it and run it again.

If you need to clear container (for example), run:
- $docker container prune -f
- $docker volume prune -f


Instructions without docker:

- $cd server
- $go mod download
- Edit s3.env file, adding AWS access key id and secret acces key.

To get AWS access key id and AWS secret access key, visit https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/setting-up.html#w2aab7b9b3b5

- If you need to use remote (Amazon) Postgresql database, change variable DB_PREFIX in .env file to AMAZON instead of LOCAL

- Otherwise, restore database schema from backup file "Postgres DB Backup.sql"  (located in root) and run your local Postgres server
    - Also, change .env file, replace variables with prefx LOCAL with your local database's host, user, password, etc

- If HTTPS support is needed, edit .env variable HTTPS_ON to true and copy your certificate as cert.pem, key as key.pem, adding them to server directory

- If CSRF support is needed, edit .env variable CSRF_ON to true

Finally, to start your server, run:
- $go run server_main.go
