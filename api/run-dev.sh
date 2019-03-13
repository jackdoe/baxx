export BAXX_POSTGRES='host=localhost user=baxx dbname=baxx password=baxx'
export BAXX_SENDGRID_KEY=''
export BAXX_S3_SECRET='bbbbbbbb'
export BAXX_S3_ACCESS_KEY='aaa'
export BAXX_S3_ENDPOINT='localhost:9000'
export BAXX_S3_DISABLE_SSL="true"
go build -o api && ./api -debug -create-tables
