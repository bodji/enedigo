# Enedigo

Enedigo is a Golang "SDK" for the Enedis "API"
It's just a wrapper for the Enedis customer portal.

The project follows the awesome work of : 
@outadoc  https://github.com/outadoc/linkindle, for the original revese engineering of Enedis portal
@beufanet https://github.com/beufanet/linkyndle, for the graphing part


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

