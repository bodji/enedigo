package enedis

import "time"

// PowerMeasure is a measure of consumtion for a given moment
// Power is expressed in kWH (multiply by 1000 before graph)
type PowerMeasure struct {
	Date  time.Time
	Power float64
	IsOffpeak bool
}
