package provider

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lestrrat/go-pdebug"
	"github.com/pkg/errors"
)

// NewFS creates a new Provider that looks for JSON documents
// from the internet over HTTP(s)
func NewHTTP() *HTTP {
	return &HTTP{
		mp: NewMap(),
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Get fetches the document specified by the `key` argument, making
// a HTTP request if necessary.
// Note that once a document is read, it WILL be cached for the
// duration of this object, unless you call `Reset`
func (hp *HTTP) Get(key *url.URL) (interface{}, error) {
	d, err := hp.GetBytes(key)
	if err != nil {
		return nil, err
	}

	var x interface{}
	if err := json.Unmarshal(d, &x); err != nil {
		return nil, errors.Wrap(err, "failed to parse JSON from HTTP resource")
	}

	return x, nil
}

func (hp *HTTP) GetBytes(key *url.URL) ([]byte, error) {
	if pdebug.Enabled {
		g := pdebug.Marker("HTTP.Get(%s)", key)
		defer g.End()
	}

	switch strings.ToLower(key.Scheme) {
	case "http", "https":
	default:
		return nil, errors.New("key is not http/https URL")
	}

	v, err := hp.mp.Get(key)
	if err == nil { // Found!
		return v.([]byte), nil
	}

	res, err := hp.Client.Get(key.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch HTTP resource")
	}
	defer res.Body.Close()

	d, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read HTTP response body")
	}

	hp.mp.Set(key.String(), d)

	return d, nil
}

// Reset resets the in memory cache of JSON documents
func (hp *HTTP) Reset() error {
	return hp.mp.Reset()
}
