id=judoc-01
docker run \
       --name $id \
       --net=host \
       --restart on-failure \
       -e JUDOC_CLUSTER="95.217.32.97,95.217.32.98" \
       -e JUDOC_BIND="127.0.0.1:9122" \
       -dit jackdoe/judoc:0.1
