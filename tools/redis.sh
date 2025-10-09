#!/bin/bash
sudo docker run --name redis -p 6379:6379 -d redis
echo "Redis started on localhost:6379"