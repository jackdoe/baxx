id=judoc-01
docker run \
       --name $id \
       --net=host \
       --restart on-failure \
       -v /root/cert/cadb.pem:/etc/scylla/cadb.pem \
       -e JUDOC_CLUSTER="95.217.32.97,95.217.32.98" \
       -e JUDOC_BIND="127.0.0.1:9122" \
       -e JUDOC_ARGS="-capath /etc/scylla/cadb.pem" \
       -dit jackdoe/judoc:0.6
