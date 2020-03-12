# covidctl
Commandline tool to put COVID-19 statistics into a InfluxDB.

## Start InfluxDB
```
docker run -d -e INFLUXDB_ADMIN_USER=adminuser -e INFLUXDB_ADMIN_PASSWORD=abcdefg -e INFLUXDB_HTTP_AUTH_ENABLED=true -e INFLUXDB_DB=covid -e INFLUXDB_READ_USER=ro -e INFLUXDB_READ_USER_PASSWORD=ropw -p 8086:8086 influxdb
```

## Start Grafana
```
docker run --rm -p 3000:3000 grafana/grafana
```
* login to grafana http://localhost:3000
* add influx datasource
  * URL: `http://172.17.0.3:8086/` (use docker inspect to find IP of docker container)
  * Database: covid
  * HTTP Method: POST

