package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/influxdata/influxdb-client-go"
	client "github.com/influxdata/influxdb1-client"
)

type Data struct {
	Country  string
	Province string
	Count    int
	Date     time.Time
}

func main() {
	var (
		targetURL = flag.String("url", "http://localhost:8086/", "influx db connection url")
		influxV2  = flag.Bool("v2", false, "use influx db v2")
		user      = flag.String("user", "", "influx username")
		password  = flag.String("password", "", "influx password")
	)

	flag.Parse()

	influxURL, err := url.Parse(*targetURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	data, err := readStats()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *influxV2 {
		err = saveToInflux2DB(influxURL, *user, *password, data)
	} else {
		err = saveToInfluxDB(influxURL, *user, *password, data)
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func readStats() ([]Data, error) {
	resp, err := http.Get("https://raw.githubusercontent.com/CSSEGISandData/COVID-19/master/csse_covid_19_data/csse_covid_19_time_series/time_series_19-covid-Confirmed.csv")
	if err != nil {
		return nil, err
	}

	records, err := csv.NewReader(resp.Body).ReadAll()
	if err != nil {
		return nil, err
	}

	data := []Data{}

	for i, dateStr := range records[0][4:] {
		date, err := time.Parse("1/2/06", dateStr)
		if err != nil {
			return nil, err
		}
		for _, record := range records[1:] {
			var c int
			if record[i+4] == "" {
				continue
			}
			c, err := strconv.Atoi(record[i+4])
			if err != nil {
				return nil, err
			}
			d := Data{
				Date:     date,
				Country:  record[1],
				Province: record[0],
				Count:    c,
			}
			data = append(data, d)
		}
	}
	return data, nil
}

func saveToInflux2DB(influxURL *url.URL, user, password string, data []Data) error {
	var (
		client *influxdb.Client
		err    error
	)
	if user != "" {
		client, err = influxdb.New(influxURL.String(), "")
	} else {
		client, err = influxdb.New(influxURL.String(), "", influxdb.WithUserAndPass(user, password))
	}
	if err != nil {
		return err
	}
	err = client.Ping(context.TODO())
	if err != nil {
		return err
	}

	metrics := []influxdb.Metric{}
	for _, d := range data {
		metric := influxdb.NewRowMetric(
			map[string]interface{}{"count": d.Count},
			"covid-19",
			map[string]string{"country": d.Country, "province": d.Province},
			d.Date)
		metrics = append(metrics, metric)

	}
	i, err := client.Write(context.TODO(), "covid-bucket", "covid-org", metrics...)
	if err != nil {
		return err
	}
	fmt.Printf("%d written\n", i)
	return err
}

func saveToInfluxDB(influxURL *url.URL, user, password string, data []Data) error {
	config := client.NewConfig()
	config.URL = *influxURL
	if user != "" {
		config.Username = user
		config.Password = password
	}
	c, err := client.NewClient(config)
	if err != nil {
		return err
	}
	_, _, err = c.Ping()
	if err != nil {
		return err
	}

	batchPoints := client.BatchPoints{
		Database: "covid",
	}
	for _, d := range data {
		point := client.Point{
			Measurement: "cases",
			Tags:        map[string]string{"country": d.Country, "province": d.Province},
			Time:        d.Date,
			Fields:      map[string]interface{}{"count": d.Count},
		}
		batchPoints.Points = append(batchPoints.Points, point)

	}
	_, err = c.Write(batchPoints)
	return err
}
