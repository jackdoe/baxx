
#!/bin/sh

/home/baxx/baxx.api -bind 127.0.0.1:9123 -root /mnt/volume_fra1_01/baxx \
                    -db-url 'baxx:PASSWORD_HERE@tcp(database.host:3306)/baxx?charset=utf8&parseTime=True&loc=Local' \
                    -db-type mysql \
                    -sendgrid SENDGRID_KEY
