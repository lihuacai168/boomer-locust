FROM locustio/locust:1.6.0

WORKDIR /app

RUN pip install prometheus-client

COPY prometheus_exporter.py /app/prometheus_exporter.py


CMD ["--master", "-f", "/app/prometheus_exporter.py"]

EXPOSE 8089
EXPOSE 5557