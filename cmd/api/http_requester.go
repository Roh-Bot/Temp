package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	url2 "net/url"
	"reflect"
	"strconv"
	"strings"
)

// DoHTTPRequest is a utility function to make HTTP requests
// It takes in the context, method, url, headers, query params and request body
// which must be strictly a json object for now
func DoHTTPRequest(
	ctx context.Context,
	method, url string,
	headers map[string]string,
	queryParams map[string]string,
	reqBody any) (*http.Response, error) {

	var body io.Reader

	if reqBody != nil {
		switch headers["Content-Type"] {
		case "application/json":
			bodyBytes, err := json.Marshal(reqBody)
			if err != nil {
				return nil, err
			}
			body = bytes.NewReader(bodyBytes)
		case "application/x-www-form-urlencoded":
			data, ok := reqBody.(map[string]string)
			if !ok {
				return nil, errors.New("reqBody must be a map[string]string for form data")
			}
			formData := url2.Values{}
			for key, value := range data {
				formData.Set(key, value)
			}
			body = strings.NewReader(formData.Encode())
		default:
			rawData, ok := reqBody.(string)
			if !ok {
				return nil, errors.New("reqBody must be a string for raw data")
			}
			body = strings.NewReader(rawData)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	for key, value := range queryParams {
		query.Add(key, value)
	}
	req.URL.RawQuery = query.Encode()

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	return client.Do(req)
}

func ParseStructToQueryParams(s interface{}) map[string]string {
	result := make(map[string]string)

	r := reflect.ValueOf(s)
	for i := 0; i < r.NumField(); i++ {
		field := r.Type().Field(i)
		fieldValue := r.Field(i).Interface()
		tag := field.Tag.Get("json")

		if tag == "-" || tag == "" {
			continue
		}

		switch v := fieldValue.(type) {
		case string:
			if v == "" {
				break
			}
			result[tag] = v
		case int:
			if v == 0 {
				break
			}
			result[tag] = strconv.Itoa(v)
		case float64:
			if v == 0 {
				break
			}
			result[tag] = strconv.FormatFloat(v, 'f', -1, 64)
		case bool:
			result[tag] = strconv.FormatBool(v)
		case []string:
			if len(v) == 0 {
				break
			}
			result[tag] = strings.Join(v, ",")
		default:
			continue
		}
	}

	return result
}
