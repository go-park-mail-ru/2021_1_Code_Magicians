This repository will contain Golang code for CodeMagicians' next big project: Pinterbest

s3.env should contain your Amazon s3 bucket's info, like following:
AWS_REGION = eu-central-1
AWS_ACCESS_KEY_ID = AKIAYF1gyr41eEXAMPLE                       // Replace with yours
AWS_SECRET_ACCESS_KEY = TLG2ltgulVp/2oWu883484uetsijerEXAMPLE  // Replace with yours
BUCKET_NAME = pinterbestbucket

To get AWS access key, visit https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/setting-up.html

To launch this bad boy:
- $go mod download

- create in root folder (where this file is) file named s3.mod. Populate it according to instruction at the start of this file
- create in root file named cert.pem, key.pem by running the following command:
- $openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout key.pem -out cert.pem

- If you need to use remote (Amazon) Postgresql database, change variables in server_main.go: replace LOCAL prefix to AMAZON prefix
- Also, change .mod file, add your local database credentials to it
- Otherwise, restore database schema from backup file "Postgres DB Backup.sql"  (located in root) and run your local Postgres server

- If HTTPS support is needed, edit .env variable HTTPS_ON = true

- $go run server_main.go

- ???

- PROFIT!
