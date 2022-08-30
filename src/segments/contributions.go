package segments

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"oh-my-posh/environment"
	"oh-my-posh/properties"
	"strconv"
	"strings"
	"time"
)

type Contributions struct {
	props properties.Properties
	env   environment.Environment

	TotalContributions int
	CurrentContributions int
	URL string
}

const (
	GITHUB_TOKEN = "github_token"
	ContributionsCacheKeyResponse string = "Contributions_response"
	ContributionsCacheKeyURL string = "Contributions_url"
)

func (d *Contributions) Enabled() bool {
	err := d.setStatus()
	return err == nil
}

func (d *Contributions) Template() string {
	return " {{ .CurrentContributions }}/{{ .TotalContributions }} "
}

func getJsonQuery() []byte {
	jsonData := map[string]string{
		"query": `
			query { 
				viewer {
					contributionsCollection{
						contributionCalendar{
							totalContributions
								weeks{
									contributionDays{
										contributionCount
										date
									}
								}
							}      
					}
				}
			}
		`,
	}
	jsonValue, _ := json.Marshal(jsonData)
	return jsonValue
}

func getTotalAndCurrentContribution(responseData map[string]interface{}) (int, int){
	root := responseData["data"].(map[string]interface{})
	viewer := root["viewer"].(map[string]interface{})
	contributionsCollection := viewer["contributionsCollection"].(map[string]interface{})
	contributionCalendar := contributionsCollection["contributionCalendar"].(map[string]interface{})
	totalContributions := contributionCalendar["totalContributions"].(float64)
	weeks := contributionCalendar["weeks"].([]interface{})
	lastWeek := weeks[len(weeks)-1].(map[string]interface{})
	contributionDays := lastWeek["contributionDays"].([]interface{})
	lastDay := contributionDays[len(contributionDays)-1].(map[string]interface{})
	return int(math.Round(totalContributions)), int(math.Round(lastDay["contributionCount"].(float64)))
}

func fetchData(jsonValue []byte, timeOut int, token string) (int, int) {
	request, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer(jsonValue))
	request.Header.Set("Authorization", "Bearer " + token)
	client := &http.Client{Timeout: time.Duration(timeOut) * time.Second}
	response, err := client.Do(request)
	var responseData map[string]interface{}
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	}
	defer response.Body.Close()
	data, _ := ioutil.ReadAll(response.Body)
	if err := json.Unmarshal(data, &responseData); err != nil {
		panic(err)
	}
	total, current := getTotalAndCurrentContribution(responseData)
	return current, total
}

func (d *Contributions) getResult() (int, int , error) {
	cacheTimeout := d.props.GetInt(properties.CacheTimeout, properties.DefaultCacheTimeout)
	if cacheTimeout > 0 {
		// check if data stored in cache
		val, found := d.env.Cache().Get(ContributionsCacheKeyResponse)
		// we got something from the cache
		if found {
			cachedData := strings.Split(val, "/")
			i, err := strconv.Atoi(cachedData[0])
			if err != nil {
				return 0, 0, err
			}
			j, err := strconv.Atoi(cachedData[1])
			if err != nil {
				return 0, 0, err
			}
			return i, j, nil
		}
	}
	httpTimeout := d.props.GetInt(properties.HTTPTimeout, properties.DefaultHTTPTimeout)
	d.URL = "https://api.Contributions.com/graphql"
	body := getJsonQuery()
	current, total := fetchData(body, httpTimeout, d.props.GetString(GITHUB_TOKEN, ""))
	if cacheTimeout > 0 {
		// persist new forecasts in cache
		d.env.Cache().Set(ContributionsCacheKeyResponse, fmt.Sprint(current)+"/"+fmt.Sprint(total), cacheTimeout)
		d.env.Cache().Set(CacheKeyURL, d.URL, cacheTimeout)
	}
	return current, total, nil
}

func (d *Contributions) setStatus() error {
	current, total, err := d.getResult()
	if err != nil {
		return err
	}
	d.CurrentContributions = current
	d.TotalContributions = total
	return nil
}

func (d *Contributions) Init(props properties.Properties, env environment.Environment) {
	d.props = props
	d.env = env
}