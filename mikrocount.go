package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
)

type entry struct {
	FromIP  net.IP
	ToIP    net.IP
	Bytes   uint
	Packets uint
}

func getEnv(env string, def string) string {
	val, ok := os.LookupEnv(env)
	if !ok {
		val = def
	}
	return val
}

func main() {
	influxDBURL := getEnv("INFLUX_URL", "http://influxdb:8086")
	influxDBUser := getEnv("INFLUX_USER", "")
	influxDBPass := getEnv("INFLUX_PWD", "")
	localCIDR := getEnv("LOCAL_CIDR", "192.168.0.0/16")
	mikrotikAddr := getEnv("MIKROTIK_ADDR", "192.168.0.1")
	timer, _ := strconv.Atoi(getEnv("MIKROCOUNT_TIMER", "15"))

	conf := client.HTTPConfig{Addr: influxDBURL}
	if len(influxDBUser) > 0 {
		conf = client.HTTPConfig{Addr: influxDBURL, Username: influxDBUser, Password: influxDBPass}
	}
	c, _ := client.NewHTTPClient(conf)
	q := client.NewQuery("CREATE DATABASE mikrocount", "", "")
	if response, err := c.Query(q); err == nil && response.Error() == nil {
		log.Println(response.Results)
	}
	defer c.Close()

	_, ipnet, _ := net.ParseCIDR(localCIDR)
	dataChan := make(chan []entry)

	for {
		select {
		case <-time.After(time.Second * time.Duration(timer)):
			go getData(mikrotikAddr, dataChan)
		case e := <-dataChan:
			go recordEntries(e, ipnet, c)
		}
	}
}

func getData(mikrotikAddr string, dataChan chan []entry) {
	var entries []entry

	resp, err := http.Get(fmt.Sprintf("http://%s/accounting/ip.cgi", mikrotikAddr))
	if err != nil {
		log.Printf("Error fetching data from Mikrotik: %s", err)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading data: %s", err)
	}

	lines := strings.Split(string(body), "\n")

	for _, l := range lines {
		if l == "" {
			break
		}
		cols := strings.Split(l, " ")
		bytes, _ := strconv.Atoi(cols[2])
		packets, _ := strconv.Atoi(cols[3])
		e := entry{
			FromIP:  net.ParseIP(cols[0]),
			ToIP:    net.ParseIP(cols[1]),
			Bytes:   uint(bytes),
			Packets: uint(packets),
		}
		entries = append(entries, e)
	}

	dataChan <- entries
}

func recordEntries(entries []entry, ipnet *net.IPNet, c client.Client) {
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "mikrocount",
		Precision: "u",
	})

	for _, e := range entries {
		var ip, direction string
		if ipnet.Contains(e.FromIP) {
			ip = e.FromIP.String()
			direction = "upload"
		} else if ipnet.Contains(e.ToIP) {
			ip = e.ToIP.String()
			direction = "download"
		} else {
			log.Printf("Weirdness! From: %s :: To: %s", e.FromIP.String(), e.ToIP.String())
			return
		}

		tags := map[string]string{
			"ip":        ip,
			"direction": direction,
		}

		fields := map[string]interface{}{
			"bytes":   e.Bytes,
			"packets": e.Packets,
		}
		pt, err := client.NewPoint("usage", tags, fields, time.Now())
		if err != nil {
			log.Println("Error: ", err.Error())
		}
		bp.AddPoint(pt)
	}

	err := c.Write(bp)
	if err != nil {
		log.Println("Error: ", err.Error())
	}
}
