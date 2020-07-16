package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/multierr"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	MarketURL               = "https://market.link/v1"
	MarketAccessKeyIDHeader = "x-access-key-id"
	MarketSecretKeyHeader   = "x-secret-key"
)

type Market struct {
	accessKey  string
	secretKey  string
	activeUser *MarketUser
}

func NewMarket(accessKey, secretKey string) (*Market, error) {
	m := &Market{
		accessKey:  accessKey,
		secretKey:  secretKey,
		activeUser: &MarketUser{},
	}
	err := m.SetActiveUser()
	return m, err
}

func (m *Market) ActiveUser() *MarketUser {
	return m.activeUser
}

func (m *Market) SetActiveUser() error {
	_, err := m.do(
		http.MethodGet,
		"/user",
		nil,
		http.StatusOK,
		m.activeUser,
	)
	return err
}

func (m *Market) CreateJob(spec *ChainlinkJobSpec) (*MarketCreated, error) {
	c := &MarketCreated{}
	spec.Initiators = spec.Attributes.Initiators
	spec.Tasks = spec.Attributes.Tasks
	_, err := m.do(
		http.MethodPost,
		"/jobs/spec",
		spec,
		http.StatusCreated,
		&c,
	)
	return c, err
}

func (m *Market) Jobs(nodeId uuid.UUID, page, size int) (*MarketJobPage, error) {
	j := &MarketJobPage{}
	_, err := m.do(
		http.MethodGet,
		fmt.Sprintf("/nodes?page=%d&size=%d&nodeId=%s", page, size, nodeId.String()),
		nil,
		http.StatusOK,
		j,
	)
	return j, err
}

func (m *Market) JobExists(jobNodeId string, networkId int) (bool, error) {
	j := &MarketJobPage{}
	_, err := m.do(
		http.MethodGet,
		fmt.Sprintf(
			"/jobs?nodeJobId[]=%s&networkId=%d",
			strings.Replace(jobNodeId, "-", "", -1),
			networkId,
		),
		nil,
		http.StatusOK,
		j,
	)
	if err != nil {
		return false, err
	} else if len(j.Data) == 0 {
		return false, nil
	}
	return true, nil
}

func (m *Market) NodeByOracleAddress(oracle common.Address, networkId int) (*MarketNode, error) {
	n := &MarketNodePage{}
	_, err := m.do(
		http.MethodGet,
		fmt.Sprintf("/search/nodes?search=%s&networkId=%d", oracle.String(), networkId),
		nil,
		http.StatusOK,
		n,
	)
	if len(n.Data) == 0 {
		return nil, errors.New("node not found")
	}
	return n.Data[0], err
}

func (m *Market) do(
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
		fmt.Sprintf("%s%s", MarketURL, endpoint),
		bytes.NewBuffer(b),
	)

	if err != nil {
		return nil, err
	}
	req.Header.Set(MarketAccessKeyIDHeader, m.accessKey)
	req.Header.Set(MarketSecretKeyHeader, m.secretKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)

	if err != nil {
		return resp, err
	} else if b, err := ioutil.ReadAll(resp.Body); err != nil {
		return resp, err
	} else if resp.StatusCode != code {
		e := MarketError{}
		defaultErr := fmt.Errorf(
			"market: unexpected response code, got %d, expected %d when calling %s",
			resp.StatusCode,
			code,
			endpoint,
		)
		if err := json.Unmarshal(b, &e); err != nil {
			return resp, defaultErr
		} else if len(e.Error) == 0 {
			return resp, defaultErr
		} else if len(e.InputErrors) > 0 {
			var merr error
			merr = multierr.Append(merr, defaultErr)
			for _, err := range e.InputErrors {
				merr = multierr.Append(merr, fmt.Errorf("%s: %s", err.Field, err.Error))
			}
			return resp, merr
		} else {
			return resp, errors.New(e.Error)
		}
	} else if obj == nil {
		return resp, nil
	} else if err := json.Unmarshal(b, obj); err != nil {
		return resp, err
	} else {
		return resp, nil
	}
}
