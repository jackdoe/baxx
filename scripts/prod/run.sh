
#!/bin/sh
export BAXX_DB=mysql
export BAXX_DB_URL='baxx:PASSWORD_HERE@/baxx?charset=utf8&parseTime=True&loc=Local'
/home/baxx/baxx.api -bind 127.0.0.1:9123 -root /mnt/volume_fra1_01/baxx
