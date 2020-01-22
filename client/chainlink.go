package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/multierr"
	"io/ioutil"
	"net/http"
)

type Chainlink struct {
	config *ChainlinkClientConfig
	cookie *http.Cookie
}

func NewChainlink(c *ChainlinkClientConfig) (*Chainlink, error) {
	cc := &Chainlink{config: c}
	return cc, cc.setSessionCookie()
}

func (c *Chainlink) CreateSpec(spec *ChainlinkJobSpec) (*ChainlinkJobSpecCreated, error) {
	j := ChainlinkJobSpecCreated{}
	_, err := c.do(http.MethodPost, "/v2/specs", spec, http.StatusOK, &j)
	return &j, err
}

func (c *Chainlink) Config() (*ChainlinkConfig, error) {
	cfg := &ChainlinkConfig{}
	_, err := c.do(
		http.MethodGet,
		"/v2/config",
		nil,
		http.StatusOK,
		cfg,
	)
	return cfg, err
}

func (c *Chainlink) ReadSpec(id string) (*ChainlinkJobSpec, error) {
	j := &ChainlinkJobSpec{}
	_, err := c.do(
		http.MethodGet,
		fmt.Sprintf("/v2/specs/%s", id),
		nil,
		http.StatusOK,
		j,
	)
	return j, err
}

func (c *Chainlink) GetSpecs(page, size int) (*ChainlinkJobSpecs, error) {
	j := &ChainlinkJobSpecs{}
	_, err := c.do(
		http.MethodGet,
		fmt.Sprintf("/v2/specs?page=%d&size=%d", page, size),
		nil,
		http.StatusOK,
		j,
	)
	return j, err
}

func (c *Chainlink) CreateBridgeType(name, url string) error {
	bta := ChainlinkBridgeTypeAttributes{Name: name, URL: url}
	_, err := c.do(
		http.MethodPost,
		"/v2/bridge_types",
		&bta,
		http.StatusOK,
		nil,
	)
	return err
}

func (c *Chainlink) ReadBridgeType(id string) (*ChainlinkBridgeType, error) {
	bt := ChainlinkBridgeType{}
	_, err := c.do(
		http.MethodGet,
		fmt.Sprintf("/v2/bridge_types/%s", id),
		nil,
		http.StatusOK,
		&bt,
	)
	return &bt, err
}

func (c *Chainlink) DeleteBridgeType(id string) error {
	_, err := c.do(
		http.MethodDelete,
		fmt.Sprintf("/v2/bridge_types/%s", id),
		nil,
		http.StatusOK,
		nil,
	)
	return err
}

func (c *Chainlink) setSessionCookie() error {
	resp, err := c.do(
		http.MethodPost,
		"/sessions",
		&ChainlinkSession{Email: c.config.Email, Password: c.config.Password},
		http.StatusOK,
		nil,
	)
	if err != nil {
		return err
	}
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "clsession" {
			c.cookie = cookie
			break
		}
	}
	if c.cookie == nil {
		return fmt.Errorf("chainlink: session cookie wasn't returned on login")
	}
	return nil
}

func (c *Chainlink) do(
	method string,
	endpoint string,
	body interface{},
	code int,
	obj interface{},
) (*http.Response, error) {
	var b []byte
	if body != nil {
		var err error
		b, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	client := http.Client{}
	req, err := http.NewRequest(
		method,
		fmt.Sprintf("%s%s", c.config.URL, endpoint),
		bytes.NewBuffer(b),
	)

	if err != nil {
		return nil, err
	}
	if c.cookie != nil {
		req.AddCookie(c.cookie)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		return resp, err
	} else if b, err := ioutil.ReadAll(resp.Body); err != nil {
		return resp, err
	} else if resp.StatusCode != code {
		errs := ChainlinkErrors{}
		defaultErr := fmt.Errorf(
			"chainlink: unexpected response code, got %d, expected %d when calling %s",
			resp.StatusCode,
			code,
			endpoint,
		)
		if err := json.Unmarshal(b, &errs); err != nil {
			return resp, defaultErr
		} else if len(errs.Errors) == 0 {
			return resp, defaultErr
		} else {
			var merr error
			merr = multierr.Append(merr, defaultErr)
			for _, err := range errs.Errors {
				merr = multierr.Append(merr, errors.New(err.Detail))
			}
			return resp, merr
		}
	} else if obj == nil {
		return resp, nil
	} else if err := json.Unmarshal(b, obj); err != nil {
		return resp, err
	} else {
		return resp, nil
	}
}
