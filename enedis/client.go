package enedis

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

const (
	ENEDIS_LOGIN_URI string = "https://espace-client-connexion.enedis.fr/auth/UI/Login"
	ENEDIS_DATA_URI  string = "https://espace-client-particuliers.enedis.fr/group/espace-particuliers/suivi-de-consommation"

	ENEDIS_DATA_PERIOD_HOUR  string = "urlCdcHeure"
	ENEDIS_DATA_PERIOD_DAY   string = "urlCdcJour"
	ENEDIS_DATA_PERIOD_MONTH string = "urlCdcMois"
	ENEDIS_DATA_PERIOD_YEAR  string = "urlCdcAn"
)

// Client is the Enedis client
// It holds the http client and all the cookies needed for the Session
type Client struct {
	config     *Config
	httpClient *http.Client
	cookieJar  *cookiejar.Jar
	logged     bool
}

type DataReturn struct {
	State struct {
		Value string `json:"value"`
	} `json:"etat"`
	Graphe struct {
		Data []struct {
			Order int64   `json:"ordre"`
			Value float64 `json:"valeur"`
		}
	}
}

func New(config *Config) (c *Client, err error) {

	// Checks
	if config.Login == "" || config.Password == "" {
		return nil, errors.New("enedis: no user or password specified")
	}

	// Parse off peak periods
	if len(config.OffpeakPeriods) > 0 {
		for _, opp := range config.OffpeakPeriods {
			err = opp.Parse()
			if err != nil {
				return nil, fmt.Errorf("enedis: fail to parse off peak period: %s", err)
			}
		}
	}

	// New client
	c = new(Client)
	c.config = config
	c.cookieJar, err = cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	c.httpClient = &http.Client{
		Jar: c.cookieJar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Login
	err = c.loginOnEnedis(config.Login, config.Password)
	if err != nil {
		return nil, err
	}

	// Make a simple get request on data API to
	// retrive other useful cookies (JSESSIONID, etc...)
	resp, err := c.httpClient.Get(ENEDIS_DATA_URI)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return c, nil
}

func (c *Client) GetDataPerHour(from time.Time, to time.Time) (powerMeasures []*PowerMeasure, err error) {
	return c.GetData(ENEDIS_DATA_PERIOD_HOUR, from, to)
}
func (c *Client) GetDataPerDay(from time.Time, to time.Time) (powerMeasures []*PowerMeasure, err error) {
	return c.GetData(ENEDIS_DATA_PERIOD_DAY, from, to)
}
func (c *Client) GetDataPerMonth(from time.Time, to time.Time) (powerMeasures []*PowerMeasure, err error) {
	return c.GetData(ENEDIS_DATA_PERIOD_MONTH, from, to)
}
func (c *Client) GetDataPerYear(from time.Time, to time.Time) (powerMeasures []*PowerMeasure, err error) {
	return c.GetData(ENEDIS_DATA_PERIOD_YEAR, from, to)
}
func (c *Client) GetData(resolution string, from time.Time, to time.Time) (powerMeasures []*PowerMeasure, err error) {

	// Init empty return
	powerMeasures = make([]*PowerMeasure, 0)

	// Fixed ID
	fixedId := "lincspartdisplaycdc_WAR_lincspartcdcportlet"

	// Form data are from and to dates
	form := url.Values{}
	form.Add(fmt.Sprintf("_%s_dateDebut", fixedId), from.Format("02/01/2006"))
	form.Add(fmt.Sprintf("_%s_dateFin", fixedId), to.Format("02/01/2006"))

	// Set hours from 00:00:00 to 23:59:59
	from = BeginningOfDay(from)
	to = EndOfDay(to)

	// URL params
	u, err := url.Parse(ENEDIS_DATA_URI)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("p_p_id", fixedId)
	q.Set("p_p_lifecycle", "2")
	q.Set("p_p_state", "normal")
	q.Set("p_p_mode", "view")
	q.Set("p_p_resource_id", resolution)
	q.Set("p_p_cacheability", "cacheLevelPage")
	q.Set("p_p_col_id", "column-1")
	q.Set("p_p_col_pos", "1")
	q.Set("p_p_col_count", "3")
	u.RawQuery = q.Encode()

	// Exec request
	resp, err := c.httpClient.Post(u.String(), "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response
	j, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var enedisDataReturn *DataReturn
	err = json.Unmarshal(j, &enedisDataReturn)
	if err != nil {
		return nil, err
	}

	// Make pretty response
	for _, measure := range enedisDataReturn.Graphe.Data {

		// Create new measure
		pm := new(PowerMeasure)

		// We compute measure date from order
		// We add a milisecond for peak/off-peak easier computation
		pm.Date = from.Add(time.Second*time.Duration((measure.Order-1)*30*60) + time.Millisecond)
		pm.Power = measure.Value / 2 // We divide by two because order is 30min

		// Test if off peak
		for _, opp := range c.config.OffpeakPeriods {
			if opp.IsInPeriod(pm.Date) {
				pm.IsOffpeak = true
			}
		}

		// Append it to return
		powerMeasures = append(powerMeasures, pm)
	}

	return powerMeasures, nil
}

// loginOnEnedis
//   - Fill a new form with rights fields
//	 - Make login POST request
//   - Check that we are loggued
//
func (c *Client) loginOnEnedis(login string, password string) (err error) {

	// The login form data
	form := url.Values{}
	form.Add("IDToken1", login)
	form.Add("IDToken2", password)
	form.Add("SunQueryParamsString", base64.StdEncoding.EncodeToString([]byte("realm=particuliers")))
	form.Add("encoded", "true")
	form.Add("gx_charset", "UTF-8")

	// Make login request on Enedis
	resp, err := c.httpClient.Post(ENEDIS_LOGIN_URI, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return errors.New(fmt.Sprintf("enedis: %s", resp.Status))
	}

	// Check cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "iPlanetDirectoryPro" {
			if cookie.Value == "LOGOUT" {
				return errors.New("enedis: credentials invalid; iPlanetDirectoryPro == LOGOUT")
			} else {
				c.logged = true
			}
		}
	}

	if !c.logged {
		return errors.New("enedis: credentials invalid; no iPlanetDirectoryPro cookie present")
	}

	return
}
