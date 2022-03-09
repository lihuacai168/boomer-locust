# Fasthttp Boomer Demo

# Quick Start

## Docker Build
```shell
docker build -t boomer:latest .
```
## Docker Run
```shell
# http get
docker run boomer:latest --run-tasks worker --url='http://httpbin.org/get?a=123' --method=GET

# http post
docker run boomer:latest --run-tasks worker --url='http://httpbin.org/post' --method=POST --data='{"a":123}'
```


# Local Dev
```shell
# mac local
go build -o boomer

# linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o boomer
```

## Debug
```shell
# get
./boomer --run-tasks worker --url='http://httpbin.org/get?a=123' --method=GET

# post
./boomer --run-tasks worker --url=http://httpbin.org/post  --method=POST --content-type="application/json"  --raw-data='{"ids": [123,234]}'
```


## As Slave
```shell
./boomer --url='http://httpbin.org/get?a=123' --method=GET
```