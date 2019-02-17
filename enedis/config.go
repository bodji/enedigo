package enedis

type Config struct {
	Login             string
	Password          string
	MaxPowerSubcribed int
	OffpeakPeriods    []*OffPeakPeriod
}
