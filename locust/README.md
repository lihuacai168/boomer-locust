# Quick Start Locust Master

## Build Locust image
```shell
cd ./locust
docker build -t locust-master:latest .
```

## Run Locust Master
```shell
docker run --name locust-master -it -d  -p 8089:8089 -p 5557:5557  locust-master:latest
```

## Open Locust Master `http://localhost:8089`