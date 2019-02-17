package enedis

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// OffPeakPeriod represents a time range where you can pay Electricity less
// From and To are like this : 07:00 14:00
type OffPeakPeriod struct {
	From string
	To   string

	fromHour   int
	fromMinute int
	toHour     int
	toMinute   int
}

func NewOffPeakPeriod(from string, to string) (opp *OffPeakPeriod, err error) {

	opp = new(OffPeakPeriod)
	opp.From = from
	opp.To = to

	err = opp.Parse()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("fail to parse offpeak period : %s", err))
	}

	return
}

func (opp *OffPeakPeriod) Parse() (err error) {

	// Remove leading 0
	opp.From = strings.TrimPrefix(opp.From, "0")
	opp.To = strings.TrimPrefix(opp.To, "0")

	// Split
	splitsFrom := strings.SplitN(opp.From, ":", 2)
	if len(splitsFrom) > 1 {

		opp.fromHour, err = strconv.Atoi(splitsFrom[0])
		if err != nil {
			return err
		}
		opp.fromMinute, err = strconv.Atoi(splitsFrom[1])
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("weird from format : %s", opp.To)
	}

	splitsTo := strings.SplitN(opp.To, ":", 2)
	if len(splitsFrom) > 1 {

		opp.toHour, err = strconv.Atoi(splitsTo[0])
		if err != nil {
			return err
		}
		opp.toMinute, err = strconv.Atoi(splitsTo[1])
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("weird to format : %s", opp.To)
	}

	return nil
}

func (opp *OffPeakPeriod) IsInPeriod(t time.Time) bool {

	from := time.Date(t.Year(), t.Month(), t.Day(), opp.fromHour, opp.fromMinute, 0, 0, time.Local)
	to := time.Date(t.Year(), t.Month(), t.Day(), opp.toHour, opp.toMinute, 0, 0, time.Local)

	if t.After(from) && t.Before(to) {
		return true
	}

	return false
}
