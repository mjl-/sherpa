package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"bitbucket.org/mjl/sherpa"
)

// Client lets you call functions from an existing Sherpa API.
// If the API was initialized with a non-nil function list, some fields will be nil (as indicated).
type Client struct {
	BaseURL       string   `json:"baseurl"`       // BaseURL the API is served from, e.g. https://sherpa.irias.nl/example/
	Functions     []string `json:"functions"`     // Function names exported by the API
	ID            string   `json:"id"`            // Short ID of the API. May be nil.
	Title         string   `json:"title"`         // Human-readable name of the API. May be nil.
	Version       string   `json:"version"`       // Version of the API, should be in the form "major.minor.patch". May be nil.
	SherpaVersion int      `json:"sherpaVersion"` // Version of the Sherpa specification this API implements. May be nil.
}

// New makes a new sherpa Client, for the given URL.
// If "functions" is nil, the API at the URL is contacted for a function list.
func New(url string, functions []string) (*Client, error) {
	c := &Client{BaseURL: url, Functions: functions}

	if functions != nil {
		return c, nil
	}

	resp, err := http.Get(url + "sherpa.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case 200:
		err = json.NewDecoder(resp.Body).Decode(c)
		if err != nil {
			return nil, err
		}
		if c.SherpaVersion != sherpa.SherpaVersion {
			return nil, fmt.Errorf("remote API uses unsupported sherpa version %d", c.SherpaVersion)
		}
		return c, nil
	case 404:
		return nil, fmt.Errorf("no API found at URL %s", url)
	default:
		return nil, fmt.Errorf("unexpected HTTP response %s for URL %s", resp.Status, url)
	}
}

// Call an API function by name.
//
// If error is not null, it's of type Error.
// If result is null, no attempt is made to parse the "result" part of the sherpa response.
func (c *Client) Call(result interface{}, functionName string, params ...interface{}) error {
	req := map[string]interface{}{
		"params": params,
	}
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(req)
	if err != nil {
		return &sherpa.Error{Code: sherpa.SherpaClientError, Message: "could not encode request parameters: " + err.Error()}
	}
	url := c.BaseURL + functionName
	resp, err := http.Post(url, "application/json", buf)
	if err != nil {
		return &sherpa.Error{Code: sherpa.SherpaClientError, Message: "sending POST request: " + err.Error()}
	}
	switch resp.StatusCode {
	case 200:
		defer resp.Body.Close()
		var response struct {
			Result json.RawMessage `json:"result"`
			Error  *sherpa.Error   `json:"error"`
		}
		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return &sherpa.Error{Code: sherpa.SherpaBadResponse, Message: "could not parse JSON response: " + err.Error()}
		}
		if response.Error != nil {
			return response.Error
		}
		if result != nil {
			err = json.Unmarshal(response.Result, result)
			if err != nil {
				return &sherpa.Error{Code: sherpa.SherpaBadResponse, Message: "could not unmarshal JSON response"}
			}
		}
		return nil
	case 404:
		return &sherpa.Error{Code: sherpa.SherpaBadFunction, Message: "no such function"}
	default:
		return &sherpa.Error{Code: sherpa.SherpaHTTPError, Message: "HTTP error from server: " + resp.Status}
	}
}