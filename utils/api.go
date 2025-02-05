package utils

import (
	"errors"
	"net/http"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/avast/retry-go"
	"github.com/gocolly/colly"
)

func (*UtilsStruct) GetDataFromAPI(url string) ([]byte, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	var body []byte
	err := retry.Do(
		func() error {
			response, err := client.Get(url)
			if err != nil {
				return err
			}
			defer response.Body.Close()
			if response.StatusCode != 200 {
				log.Errorf("API: %s responded with status code %d", url, response.StatusCode)
				return errors.New("unable to reach API")
			}
			body, err = IOInterface.ReadAll(response.Body)
			if err != nil {
				return err
			}
			return nil
		}, retry.Attempts(2), retry.Delay(time.Second*2))
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (*UtilsStruct) GetDataFromJSON(jsonObject map[string]interface{}, selector string) (interface{}, error) {
	if selector[0] == '[' {
		selector = "$" + selector
	} else {
		selector = "$." + selector
	}
	return jsonpath.Get(selector, jsonObject)
}

func (*UtilsStruct) GetDataFromXHTML(url string, selector string) (string, error) {
	c := colly.NewCollector()
	var priceData string
	c.OnXML(selector, func(e *colly.XMLElement) {
		priceData = e.Text
	})
	err := c.Visit(url)
	if err != nil {
		return "", err
	}
	return priceData, nil
}
