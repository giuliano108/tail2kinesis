package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
)

type Transform interface {
	Do(string) (string, error)
}

// `Identity` doesn't transform the input at all

type Identity struct{}

func (t *Identity) Do(s string) (string, error) {
	return s, nil
}

func NewIdentity() (Transform, error) {
	return &Identity{}, nil
}

// `AccessLog2JSON`

type AccessLog2JSON struct {
	format string
	re     *regexp.Regexp
}

/*
   `AccessLog2JSONQuery` transforms NGINX logs in this format:

    log_format timing '$remote_addr - $remote_user [$time_local]  '
        '"$request" $status $body_bytes_sent '
        '"$http_referer" "$http_user_agent" '
        '$request_time ($upstream_response_time)';

	This input:

		10.40.24.157 - - [08/May/2019:08:17:15 -0700]  "GET /endpoint?k1=v1&k2=2 HTTP/1.1" 200 0 "-" "Ruby" 0.000 (-)

	Yields the following JSON formatted output (all in one line):

		{
		  "_meta": {
			"path": "/endpoint",
			"time": "08/May/2019:08:17:15 -0700"
		  },
		  "k1": "v1",
		  "k2": 2
		}
*/

type AccessLog2JSONQuery struct {
	AccessLog2JSON
}

func (t *AccessLog2JSONQuery) Do(s string) (string, error) {
	matches := t.re.FindAllStringSubmatch(s, -1)
	if matches == nil {
		return "", errors.New("no match")
	}

	data := make(map[string]interface{})
	data["_meta"] = make(map[string]string)
	data["_meta"].(map[string]string)["time"] = matches[0][4]
	url, err := url.Parse(matches[0][6])
	if err == nil {
		data["_meta"].(map[string]string)["path"] = url.Path
		for k, v := range url.Query() {
			// if a value looks like an integer, store it as such
			value, err := strconv.Atoi(v[0])
			if err == nil {
				data[k] = value
			} else {
				data[k] = v[0]
			}
		}
	}
	msg, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", string(msg)), nil
}

func NewAccessLog2JSON(format string) (Transform, error) {
	var err error
	var re *regexp.Regexp

	if format == "query" {
		re, err = regexp.Compile(`(\S+)\s+(\S+)\s+(\S+)\s+\[(.*?)\]\s+"(\S+)\s+(\S+)\s+(\S+)"\s+(\S+)\s+(\S+)\s+"(\S+)"\s+"(.*?)\s+(\S+)\s+\((.*?)\)`)
		if err != nil {
			return nil, err
		}
		return &AccessLog2JSONQuery{AccessLog2JSON{format: format, re: re}}, nil
	} else {
		return nil, fmt.Errorf("unsupported log format: %s", format)
	}
}
