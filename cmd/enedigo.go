package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/bodji/enedigo/enedis"
	influx "github.com/influxdata/influxdb1-client/v2"
	"github.com/spf13/viper"
)

var (
	conf EnedigoConfig
)

type EnedigoConfig struct {
	Enedis struct {
		User           string                  `mapstructure:"user"`
		Password       string                  `mapstructure:"password"`
		MaxPower       int                     `mapstructure:"maxPower"`
		OffpeakPeriods []*enedis.OffPeakPeriod `mapstructure:"offpeakPeriods"`
	} `mapstructure:"enedis"`
	Influx struct {
		Url      string `mapstructure:"url"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Database string `mapstructure:"database"`
		Measure  string `mapstructure:"measure"`
	} `mapstructure:"influx"`
}

func main() {

	// Read config file
	viper.SetDefault("days", 5)
	viper.SetConfigName("enedigo")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file: %s", err.Error())
	}
	err = viper.Unmarshal(&conf)
	if err != nil {
		log.Fatalf("Error reading config file: %s", err.Error())
	}

	// Some flags
	days := flag.Int("days", 1, "Number of days to get from Enedis")
	flag.Parse()

	// Instantiate Influx client
	influxClient, err := influx.NewHTTPClient(influx.HTTPConfig{
		Addr:     conf.Influx.Url,
		Username: conf.Influx.User,
		Password: conf.Influx.Password,
	})
	if err != nil {
		log.Fatalf("Error creating InfluxDB Client: ", err.Error())
	}

	// Retrieve data from Enedis
	log.Printf("Will get last %d days from Enedis", *days)
	log.Printf("Creating enedis client....")

	enedisClient, err := enedis.New(&enedis.Config{
		Login:          conf.Enedis.User,
		Password:       conf.Enedis.Password,
		OffpeakPeriods: conf.Enedis.OffpeakPeriods,
	})
	if err != nil {
		log.Fatalf("Fail to instantiate enedis : %s", err)
	}

	log.Printf("Getting data from enedis...")
	measures, err := enedisClient.GetDataPerHour(time.Now().AddDate(0, 0, -5), time.Now())
	if err != nil {
		log.Fatalf("Fail to get measures from enedis : %s", err)
	}

	// Create a new point batch
	bp, _ := influx.NewBatchPoints(influx.BatchPointsConfig{
		Database:  conf.Influx.Database,
		Precision: "s",
	})

	// Make some points
	for _, measure := range measures {

		creuses := "0"
		pleines := "1"

		if measure.IsOffpeak {
			creuses = "1"
			pleines = "0"
		}

		// Create a point and add to batch
		tags := map[string]string{
			"heures_creuses":  creuses,
			"heures_pleines":  pleines,
			"heures_normales": "0",
		}
		fields := map[string]interface{}{
			"value": measure.Power * 1000,
			"max":   conf.Enedis.MaxPower * 1000,
		}

		pt, err := influx.NewPoint(conf.Influx.Measure, tags, fields, measure.Date)
		if err != nil {
			fmt.Println("Error: ", err.Error())
		}

		bp.AddPoint(pt)

		log.Printf("Got measure of %s : %.3f | HC:%s | HP:%s", measure.Date.Format(time.RFC3339), measure.Power, creuses, pleines)
	}

	// Write the batch
	log.Printf("Pushing %d points to InfluxDB...", len(bp.Points()))
	err = influxClient.Write(bp)
	if err != nil {
		log.Fatalf("Fail to write points to influx : %s", err)
	}
}