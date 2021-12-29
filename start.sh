#!/bin/bash
trap "rm server;kill 0" EXIT

go build -o server

# # ./server -node=0 &
./server -node=1 &
./server -node=2 &

./server -api=0 &
./server -api=1 &
./server -api=2 &

sleep 5
echo ">>> start test"
curl "http://localhost:9999/api?key=Tom" &
curl "http://localhost:9999/api?key=Tom" &
curl "http://localhost:9999/api?key=Tom" &
# 上面的可能会合并成一条请求
# 下面的请求会独立开
sleep 2
curl "http://localhost:9999/api?key=Tom" &

wait