for i in baxx-api baxx-send-email baxx-notification-rules baxx-status baxx-watcher; do
    docker stop $i
    docker rm $i
done
