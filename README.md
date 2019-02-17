# Enedigo

Enedigo is a Golang "SDK" for the Enedis "API"
It's just a wrapper for the Enedis customer portal.

The project follows the awesome work of : 
 - @outadoc  https://github.com/outadoc/linkindle, for the original revese engineering of Enedis portal
 - @beufanet https://github.com/beufanet/linkyndle, for the graphing part


## Create Enedis Client

You can simply create a Client with username, password, and optionaly off-peak periods, if you are concerned with peak/offpeak subcription :

    enedisClient, err := enedis.New(&enedis.Config{
	    Login: "user@tld.fr",
	    Password: "",
	    OffpeakPeriods: []*enedis.OffPeakPeriod{
		    &enedis.OffPeakPeriod{From: "01:00", To: "07:00"},
		    &enedis.OffPeakPeriod{From: "12:30", To: "14:30"},
	    }
	})    


    if err !=  nil {
	    log.Fatalf("Fail to instantiate enedis : %s", err)
    }


## Get power measurements

To get power measurements from your linky, you have GetDataPerHour(from time.Time, to time.Time)

    measures, err := enedisClient.GetDataPerHour(time.Now().AddDate(0, 0, -5), time.Now())
    if err !=  nil {
	    log.Fatalf("Fail to get measures from enedis : %s", err)
	}


You'll get an []*enedis.PowerMeasure :

    type  PowerMeasure  struct {
	    Date time.Time
	    Power float64
	    IsOffpeak bool
    }
    
Where :

 - Date is the time of the measure
 - Power is the amount of kWH of the measure
 - IsOffpeak will tell you whether or not the mesure is off-peak (If you have configured off-peak periods obviously)



## Pushing to InfluxDB
I have created a little tool which mimic @beufanet python script in cmd/enedis2influx.go
You just have to tune enedigo.yml file, and run it :

     $ ./enedis2influx --days 3
    2019/02/17 17:38:16 Will get last 3 days from Enedis
    2019/02/17 17:38:16 Creating enedis client....
    2019/02/17 17:38:18 Getting data from enedis...
    2019/02/17 17:38:20 Got measure of 2019-02-12T00:00:00+01:00 : 1.634 | HC:0 | HP:1
    2019/02/17 17:38:20 Got measure of 2019-02-12T00:30:00+01:00 : 1.658 | HC:0 | HP:1
    2019/02/17 17:38:20 Got measure of 2019-02-12T01:00:00+01:00 : 2.828 | HC:0 | HP:1
    2019/02/17 17:38:20 Got measure of 2019-02-12T01:30:00+01:00 : 2.329 | HC:1 | HP:0
    2019/02/17 17:38:20 Got measure of 2019-02-12T02:00:00+01:00 : 1.621 | HC:1 | HP:0
    2019/02/17 17:38:20 Got measure of 2019-02-12T02:30:00+01:00 : 1.639 | HC:1 | HP:0
    2019/02/17 17:38:20 Got measure of 2019-02-12T03:00:00+01:00 : 1.617 | HC:1 | HP:0
    2019/02/17 17:38:20 Got measure of 2019-02-12T03:30:00+01:00 : 1.622 | HC:1 | HP:0
    .....
    2019/02/17 17:38:20 Pushing 240 points to InfluxDB...
