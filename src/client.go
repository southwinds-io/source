/*
  Onix Config Manager - Source
  © 2022 southwinds.io - All rights reserved
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package src

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/invopop/jsonschema"
	"github.com/southwinds-io/source/service"
	"net/http"
	"time"
)

type ClientOptions struct {
	InsecureSkipVerify bool
	Timeout            time.Duration
}

func defaultOptions() *ClientOptions {
	return &ClientOptions{
		InsecureSkipVerify: true,
		Timeout:            60 * time.Second,
	}
}

type Client struct {
	*http.Client
	host, token string
}

func New(host, user, pwd string, opts *ClientOptions) Client {
	if opts == nil {
		opts = defaultOptions()
	}
	return Client{ // the http client instance
		host:  host,
		token: basicToken(user, pwd),
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: opts.InsecureSkipVerify,
				},
			},
			// set the client timeout period
			Timeout: opts.Timeout,
		}}
}

func (c *Client) SetType(key string, obj any) error {
	// reflects the json schema from the specified object
	schemaObj := jsonschema.Reflect(obj)
	// marshal the object to json bytes
	schema, err := json.Marshal(schemaObj)
	if err != nil {
		return err
	}
	request, err := http.NewRequest(http.MethodPut, c.url("/type/%s", key), bytes.NewReader(schema))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", c.token)
	request.Header.Set("User-Agent", fmt.Sprintf("SOURCE-CLIENT-%s", service.Version))
	resp, reqErr := c.Do(request)
	if reqErr != nil {
		return reqErr
	}
	if resp.StatusCode > 299 {
		return fmt.Errorf("cannot set type, source server responded with: %s", resp.Status)
	}
	return nil
}

func (c *Client) url(format string, args ...any) string {
	v := fmt.Sprintf("%s%s", c.host, fmt.Sprintf(format, args...))
	return v
}

func basicToken(user string, pwd string) string {
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", user, pwd))))
}
