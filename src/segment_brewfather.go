package main

import (
	"encoding/json"
	"encoding/base64"
	"errors"
	"fmt"
	//"time"
)

// segment struct, makes templating easier
type brewfather struct {
	props properties
	env   environmentInfo

	BatchReading
}

const (
	
	BFUserID	Property = "user_id"
	BFAPIKey	Property = "api_key"

	BatchId	Property = "batch_id"

	BFCacheTimeout Property = "cache_timeout"
)

// Returned from https://api.brewfather.app/v1/batches/:batch_id/readings
type BatchReading struct {
    Comment		string 	`json:"comment"`
    Gravity		float32 `json:"sg"`
    DeviceType	string	`json:"type"`
	DeviceId	string	`json:"id"`
	Temperature	float32	`json:"temp"` // celsius
	Timepoint	int64	`json:"timepoint"` // << check what these are...
    Time		int64	`json:"time"`
}

func (bf *brewfather) enabled() bool {
	data, err := bf.getResult()
	if err != nil {
		return false
	}
	bf.BatchReading = *data
	//bf.TrendIcon = bf.getTrendIcon()

	return true
}

func (bf *brewfather) getTrendIcon() string {
	// switch bf.Direction {
	// case "DoubleUp":
	// 	return bf.props.getString(DoubleUpIcon, "↑↑")
	// case "SingleUp":
	// 	return bf.props.getString(SingleUpIcon, "↑")
	// case "FortyFiveUp":
	// 	return bf.props.getString(FortyFiveUpIcon, "↗")
	// case "Flat":
	// 	return bf.props.getString(FlatIcon, "→")
	// case "FortyFiveDown":
	// 	return bf.props.getString(FortyFiveDownIcon, "↘")
	// case "SingleDown":
	// 	return bf.props.getString(SingleDownIcon, "↓")
	// case "DoubleDown":
	// 	return bf.props.getString(DoubleDownIcon, "↓↓")
	// default:
		return ""
	//}
}

func (bf *brewfather) string() string {
	segmentTemplate := bf.props.getString(SegmentTemplate, "{{.Gravity}}{{.Temperature}}")
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  bf,
		Env:      bf.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}

	return text
}

func (bf *brewfather) getResult() (*BatchReading, error) {
	parseSingleElement := func(data []byte) (*BatchReading, error) {
		var result []*BatchReading
		err := json.Unmarshal(data, &result)
		if err != nil {
			return nil, err
		}
		if len(result) == 0 {
			return nil, errors.New("no elements in the array")
		}
		return result[0], nil
	}
	getCacheValue := func(key string) (*BatchReading, error) {
		val, found := bf.env.cache().get(key)
		// we got something from the cache
		if found {
			if data, err := parseSingleElement([]byte(val)); err == nil {
				return data, nil
			}
		}
		return nil, errors.New("no data in cache")
	}

	userId := bf.props.getString(BFUserID, "")
	apiKey := bf.props.getString(BFAPIKey, "")

	if len(userId) == 0 {
		return nil, errors.New("missing Brewfather user id (user_id)")
	}

	if len (apiKey)== 0 {
		return nil, errors.New("missing Brewfather api key (api_key)")
	}

	authString := fmt.Sprintf("%s:%s", userId, apiKey)
	authStringb64 := base64.StdEncoding.EncodeToString([]byte(authString))
	authHeader := fmt.Sprintf("Basic %s", authStringb64)
	batchId := "tbc9nltgWRam8M8yB8QpMGEDvwQteV"
	url := fmt.Sprintf("https://api.brewfather.app/v1/batches/%s/readings", batchId)
	
	httpTimeout := bf.props.getInt(HTTPTimeout, DefaultHTTPTimeout)
	// natural and understood bf timeout is 5, anything else is unusual
	cacheTimeout := bf.props.getInt(BFCacheTimeout, 5)

	if cacheTimeout > 0 {
		if data, err := getCacheValue(url); err == nil {
			return data, nil
		}
	}

	body, err := bf.env.doGetWithAuth(url, httpTimeout, authHeader)
	if err != nil {
		return nil, err
	}
	var arr []*BatchReading
	err = json.Unmarshal(body, &arr)
	if err != nil {
		return nil, err
	}

	data, err := parseSingleElement(body)
	if err != nil {
		return nil, err
	}

	if cacheTimeout > 0 {
		// persist new sugars in cache
		bf.env.cache().set(url, string(body), cacheTimeout)
	}
	return data, nil
}

func (bf *brewfather) init(props properties, env environmentInfo) {
	bf.props = props
	bf.env = env
}
