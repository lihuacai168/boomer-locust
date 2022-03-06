FROM busybox

WORKDIR /app

COPY boomer_linux .
COPY data.csv .

ENTRYPOINT ["/app/boomer_linux"]
