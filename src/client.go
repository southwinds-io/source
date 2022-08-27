/*
  Source Configuration Service
  © 2022 Southwinds Tech Ltd - www.southwinds.io
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
	"github.com/southwinds-io/source/cdb"
	"github.com/southwinds-io/source/service"
	"io"
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

func (c *Client) SetItem(key, itemType string, obj any) error {
	objBytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	request, err := http.NewRequest(http.MethodPut, c.url("/item/%s", key), bytes.NewReader(objBytes))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", c.token)
	request.Header.Set("User-Agent", fmt.Sprintf("SOURCE-CLIENT-%s", service.Version))
	if len(itemType) > 0 {
		request.Header.Set("Source-Type", itemType)
	}
	resp, reqErr := c.Do(request)
	if reqErr != nil {
		return reqErr
	}
	if resp.StatusCode > 299 {
		return fmt.Errorf("cannot set item, source server responded with: %s", resp.Status)
	}
	return nil
}

func (c *Client) GetChildren(itemKey string) ([]cdb.I, error) {
	request, err := http.NewRequest(http.MethodGet, c.url("/item/%s/children", itemKey), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", c.token)
	request.Header.Set("User-Agent", fmt.Sprintf("SOURCE-CLIENT-%s", service.Version))
	resp, reqErr := c.Do(request)
	if reqErr != nil {
		return nil, reqErr
	}
	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("cannot get children for item, source server responded with: %s", resp.Status)
	}
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("cannot read response body: %s", readErr)
	}
	var items []cdb.I
	err = json.Unmarshal(body, &items)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal response body: %s", err)
	}
	return items, nil
}

func (c *Client) GetParents(itemKey string) ([]cdb.I, error) {
	request, err := http.NewRequest(http.MethodGet, c.url("/item/%s/parents", itemKey), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", c.token)
	request.Header.Set("User-Agent", fmt.Sprintf("SOURCE-CLIENT-%s", service.Version))
	resp, reqErr := c.Do(request)
	if reqErr != nil {
		return nil, reqErr
	}
	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("cannot get parents for item, source server responded with: %s", resp.Status)
	}
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("cannot read response body: %s", readErr)
	}
	var items []cdb.I
	err = json.Unmarshal(body, &items)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal response body: %s", err)
	}
	return items, nil
}

func (c *Client) Tag(itemKey, tagName, tagValue string) error {
	var tag string
	if len(tagName) > 0 {
		if len(tagValue) > 0 {
			tag = fmt.Sprintf("%s|%s", tagName, tagValue)
		} else {
			tag = tagName
		}
	} else {
		return fmt.Errorf("a tag name is required")
	}
	request, err := http.NewRequest(http.MethodPut, c.url("/item/%s/tag/%s", itemKey, tag), nil)
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
		return fmt.Errorf("cannot tag item, source server responded with: %s", resp.Status)
	}
	return nil
}

func (c *Client) Untag(itemKey, tagName string) error {
	if len(tagName) == 0 {
		return fmt.Errorf("a tag name is required")
	}
	request, err := http.NewRequest(http.MethodDelete, c.url("/item/%s/tag/%s", itemKey, tagName), nil)
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
		return fmt.Errorf("cannot tag item, source server responded with: %s", resp.Status)
	}
	return nil
}

func (c *Client) Link(fromKey, toKey string) error {
	request, err := http.NewRequest(http.MethodPut, c.url("/link/%s/to/%s", fromKey, toKey), nil)
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
		return fmt.Errorf("cannot link items, source server responded with: %s", resp.Status)
	}
	return nil
}

func (c *Client) Unlink(fromKey, toKey string) error {
	request, err := http.NewRequest(http.MethodDelete, c.url("/link/%s/to/%s", fromKey, toKey), nil)
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
		return fmt.Errorf("cannot unlink items, source server responded with: %s", resp.Status)
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
