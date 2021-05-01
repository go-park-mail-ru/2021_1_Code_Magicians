This repository will contain Golang code for CodeMagicians' next big project: Pinterbest

Instructions:
$command means "run that command in console, from project's main folder"
- $go mod download
- $touch s3.env  // This file will contain info needed to connect to Amazon S3 bucket
- Add following lines to s3.env, that was just created:
AWS_REGION = eu-central-1
AWS_ACCESS_KEY_ID = AKIAYF1gyr41eEXAMPLE                       // Replace with yours
AWS_SECRET_ACCESS_KEY = TLG2ltgulVp/2oWu883484uetsijerEXAMPLE  // Replace with yours
BUCKET_NAME = pinterbestbucket

To get AWS access key id and AWS secret access key, visit https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/setting-up.html#w2aab7b9b3b5

- If you need to use remote (Amazon) Postgresql database, change variable DB_PREFIX in .env file in project's main folder to AMAZON instead of LOCAL

- Otherwise, restore database schema from backup file "Postgres DB Backup.sql"  (located in root) and run your local Postgres server
- And in that case, change .env file, add your local database credentials to it

- If HTTPS support is needed, edit .env variable HTTPS_ON to true and run the following command
- $openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout key.pem -out cert.pem  // Generates ssl keys needed for HTTPS to work

Finally, to start your server, run:
- $go run server_main.go


- $docker build -t postgresdb:0.0.1 -f postgres/dockerfile ./postgres/
- $docker tag postgresdb:0.0.1 postgresdb:latest
- $docker run postgresdb -p 5432:5432
