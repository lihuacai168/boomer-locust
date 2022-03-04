FROM busybox

COPY boomer_linux /usr/local/bin/boomer

ENTRYPOINT ["/usr/local/bin/boomer"]
