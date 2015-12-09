package stats

import (
	"log"
	"time"

	"github.com/influxdb/influxdb/client/v2"
)

func sendInflux(server, database, username, password string, interval time.Duration) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     server,
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatalf("Error creating client for influx: %s", err.Error())
	}
	for {
		time.Sleep(interval)
		now := time.Now()
		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  database,
			Precision: "s",
		})
		if err != nil {
			log.Printf("Error creating batch for influx: %s", err.Error())
			continue
		}

		for _, pt := range Statistics.GetAll() {
			pt2, err := client.NewPoint(pt.Metric, pt.Tags, map[string]interface{}{"value": pt.Value}, now)
			if err != nil {
				log.Printf("Error creating point for influx: %s", err.Error())
				continue
			}
			bp.AddPoint(pt2)
		}

		if err = c.Write(bp); err != nil {
			log.Printf("Error sending to influx: %s", err.Error())
			continue
		}
		log.Printf("Sent %d points to influx.", len(bp.Points()))

	}
}
