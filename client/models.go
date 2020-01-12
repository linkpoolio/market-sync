package client

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/satori/go.uuid"
)

type ChainlinkClientConfig struct {
	Email    string
	Password string
	URL      string
}

type ChainlinkErrors struct {
	Errors []ChainlinkError `json:"errors"`
}

type ChainlinkError struct {
	Detail string `json:"detail"`
}

type ChainlinkBridgeType struct {
	Data ChainlinkBridgeTypeData `json:"data"`
}

type ChainlinkBridgeTypeData struct {
	Attributes ChainlinkBridgeTypeAttributes `json:"attributes"`
}

type ChainlinkBridgeTypeAttributes struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type ChainlinkSession struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ChainlinkMeta struct {
	Count int `json:"count"`
}

type ChainlinkResponseArray struct {
	Data []map[string]interface{}
}

type ChainlinkResponse struct {
	Data map[string]interface{}
}

type ChainlinkInitiator struct {
	Type                     string `json:"type"`
	ChainlinkInitiatorParams `json:"params,omitempty"`
}

type ChainlinkInitiatorParams struct {
	Address common.Address `json:"address,omitempty" gorm:"index"`
}

type ChainlinkJobSpec struct {
	ID         string                     `json:"id"`
	Name       string                     `json:"name,omitempty"`
	NodeID     *uuid.UUID                 `json:"nodeId,omitempty"`
	Attributes ChainlinkJobSpecAttributes `json:"attributes"`
	MinPayment string                     `json:"minPayment,omitempty"`
	Initiators []*ChainlinkInitiator 	  `json:"initiators,omitempty"`
	Tasks      []*ChainlinkTaskSpec       `json:"tasks,omitempty"`
}

type ChainlinkJobSpecAttributes struct {
	Initiators []*ChainlinkInitiator `json:"initiators,omitempty"`
	Tasks      []*ChainlinkTaskSpec  `json:"tasks,omitempty"`
}

type ChainlinkTaskSpec struct {
	Type          string                 `json:"type"`
	Confirmations uint64                 `json:"confirmations,omitempty"`
	Params        map[string]interface{} `json:"params,omitempty" gorm:"type:text"`
}

type ChainlinkJobSpecs struct {
	Data []*ChainlinkJobSpec `json:"data"`
	Meta ChainlinkMeta       `json:"meta"`
}

type ChainlinkJobSpecCreated struct {
	Data *ChainlinkJobSpec `json:"data"`
}

type ChainlinkConfig struct {
	Data struct {
		Attributes struct {
			OracleContractAddress common.Address `json:"oracleContractAddress"`
			ETHChainID            int            `json:"ethChainId"`
		} `json:"attributes"`
	} `json:"data"`
}

type MarketJob struct {
	Name      string        `json:"name"`
	NodeID    uuid.UUID     `json:"nodeId"`
	NodeJobID string        `form:"nodeJobId,omitempty"`
	Tasks     []*MarketTask `form:"tasks,omitempty"`
	Cost      string        `json:"cost"`
}

type MarketTask struct {
	AdapterID uuid.UUID          `json:"adapterId"`
	Param     []*MarketTaskParam `json:"param"`
	Index     uint               `json:"index"`
}

type MarketTaskParam struct {
	Key    string        `json:"key"`
	Values []interface{} `json:"values"`
}

type MarketCreated struct {
	ID uuid.UUID `json:"id"`
}

type MarketError struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
	InputErrors []struct{
		Field string `json:"field"`
		Error string `json:"error"`
	} `json:"inputErrors"`
}

type MarketUser struct {
	ID uuid.UUID `json:"id"`
}

type MarketNode struct {
	ID            uuid.UUID         `json:"id"`
	OracleAddress common.Address    `json:"oracleAddress"`
	Network       MarketNodeNetwork `json:"network"`
}

type MarketNodeNetwork struct {
	ID int `json:"id"`
}

type MarketNodePage struct {
	Data []*MarketNode `json:"data"`
	MarketMeta
}

type MarketJobPage struct {
	Data []*MarketJob `json:"data"`
	MarketMeta
}

type MarketMeta struct {
	TotalCount int `json:"totalCount"`
}
