id=judoc-01
docker run \
       --name $id \
       --restart on-failure \
       -p 127.0.0.1:9122:9122 \
       -e JUDOC_CLUSTER="95.217.32.97,95.217.32.98" \
       -e JUDOC_BIND=":9122" \
       -dit jackdoe/judoc:0.1
