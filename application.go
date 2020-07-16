package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/fatih/color"
	uuid "github.com/satori/go.uuid"
	"github.com/tcnksm/go-input"
	"github.com/tidwall/pretty"
	"market-sync/client"
	"regexp"
	"strconv"
	"strings"
)

type Application struct {
	config    *Config
	chainlink *client.Chainlink
	market    *client.Market
}

type Config struct {
	UI                *input.UI
	ChainlinkEmail    string
	ChainlinkPassword string
	ChainlinkURL      string

	MarketAccessKey string
	MarketSecretKey string
}

func NewApplication(config *Config) (*Application, error) {
	c, err := client.NewChainlink(&client.ChainlinkClientConfig{
		Email:    config.ChainlinkEmail,
		Password: config.ChainlinkPassword,
		URL:      config.ChainlinkURL,
	})
	if err != nil {
		return nil, err
	}

	m, err := client.NewMarket(config.MarketAccessKey, config.MarketSecretKey)
	if err != nil {
		return nil, err
	}

	return &Application{
		config:    config,
		chainlink: c,
		market:    m,
	}, nil
}

func (a *Application) MarketNode() (*client.MarketNode, error) {
	yellow := color.New(color.FgYellow).SprintFunc()

	oracleNilError := errors.New("Chainlink oracle address is nil, please ensure `ORACLE_CONTRACT_ADDRESS` is set in configuration")
	cfg, err := a.chainlink.Config()
	if err != nil {
		return nil, err
	} else if cfg.Data.Attributes.OracleContractAddress == nil {
		return nil, oracleNilError
	}
	oracle := cfg.Data.Attributes.OracleContractAddress
	chainId := cfg.Data.Attributes.ETHChainID

	fmt.Printf("%s %s\n", yellow("Oracle Address:"), oracle.String())
	if oracle.String() == common.HexToAddress("0x0").String() {
		return nil, oracleNilError
	}

	node, err := a.market.NodeByOracleAddress(oracle, chainId)
	if err != nil {
		return nil, errors.New("Chainlink node not found on the Market, create it before running this tool")
	}
	return node, nil
}

func (a *Application) SyncJobSpecs(nodeId uuid.UUID, networkId int) error {
	yellow := color.New(color.FgYellow).SprintFunc()

	specs, err := a.chainlink.GetSpecs(1, 1)
	if err != nil {
		return err
	}
	specCount := specs.Meta.Count
	fmt.Printf("%s %d\n\n", yellow("Job Spec Count:"), specCount)

	page := 1
	loopBatch := 5
	for i := 0; i < specCount; i = i + loopBatch {
		specs, err := a.chainlink.GetSpecs(page, loopBatch)
		if err != nil {
			return err
		}
		for j, spec := range specs.Data {
			color.Green("Job Spec %d", i+j+1)
			exists, err := a.market.JobExists(spec.ID, networkId)
			if err != nil {
				return err
			} else if exists {
				fmt.Printf("%s %s\n", yellow("Job ID Exists on Market:"), spec.ID)
				continue
			}
			spec.NodeID = &nodeId
			a.promptJobSpec(spec)
		}
		page++
	}
	return nil
}

func (a *Application) promptJobSpec(spec *client.ChainlinkJobSpec) {
	a.outputJSON(spec)
	if answer, err := a.config.UI.Ask("Sync this job spec to the Market? [y/n]", &input.Options{
		Default:  "n",
		Loop:     true,
		Required: true,
		ValidateFunc: booleanInputValidation,
	}); err != nil {
		exit(err)
	} else if answer == "y" {
		a.syncJob(spec)
	}
}

func (a *Application) promptJobName() string {
	if answer, err := a.config.UI.Ask("Job name", &input.Options{
		Loop:     true,
		Required: true,
		ValidateFunc: func(s string) error {
			if matcher, err := regexp.Compile(`^[a-zA-Z0-9_\-\.\ \+\>\=]{2,30}$`); err != nil {
				return err
			} else if !matcher.Match([]byte(s)) {
				return errors.New("invalid job name, must be: (2-30 length, a-z, A-Z, 0-9, ), -, ., , +, >, =)")
			}
			return nil
		},
	}); err != nil {
		exit(err)
	} else {
		return answer
	}
	return ""
}

func (a *Application) promptJobCost() string {
	if answer, err := a.config.UI.Ask("Job cost", &input.Options{
		Default:  "100000000000000000",
		Loop:     true,
		Required: true,
		ValidateFunc: func(s string) error {
			if _, err := strconv.ParseInt(s, 10, 64); err != nil {
				return errors.New("answer must be an int64")
			}
			return nil
		},
	}); err != nil {
		exit(err)
	} else {
		 return answer
	}
	return ""
}

func (a *Application) promptRetry(spec *client.ChainlinkJobSpec) {
	if answer, err := a.config.UI.Ask("Retry adding this job? [y/n]", &input.Options{
		Default:  "n",
		Loop:     true,
		Required: true,
		ValidateFunc: booleanInputValidation,
	}); err != nil {
		return
	} else if answer == "y" {
		a.syncJob(spec)
	}
}

func (a *Application) promptEdit(spec *client.ChainlinkJobSpec) error {
	if answer, err := a.config.UI.Ask("Edit job spec parameters? [y/n]", &input.Options{
		Default:  "n",
		Loop:     true,
		Required: true,
		ValidateFunc: booleanInputValidation,
	}); err != nil {
		return err
	} else if answer == "n" {
		return nil
	}

	var types []string
	taskMap := map[string]*client.ChainlinkTaskSpec{}
	for _, t := range spec.Attributes.Tasks {
		types = append(types, t.Type)
		taskMap[t.Type] = t
	}
	if len(types) == 0 {
		return errors.New("no job spec tasks")
	}
	answer, err := a.config.UI.Select("Select task type", types, &input.Options{})
	if err != nil {
		return err
	}
	t := taskMap[answer]
	var params []string
	for k := range t.Params {
		params = append(params, k)
	}
	if len(params) == 0 {
		return errors.New("no parameters for this task type")
	}
	if key, err := a.config.UI.Select("Select parameter", params, &input.Options{}); err != nil {
		return err
	} else if value, err := a.config.UI.Ask("Enter new value", &input.Options{
		Default: fmt.Sprint(t.Params[key]),
	}); err != nil {
		return err
	} else {
		t.Params[key] = value
	}
	return a.promptEdit(spec)
}

func (a *Application) syncJob(spec *client.ChainlinkJobSpec) {
	spec.Name = a.promptJobName()
	if len(spec.MinPayment) == 0 {
		spec.MinPayment = a.promptJobCost()
	}
	if err := a.promptEdit(spec); err != nil {
		displayError(err)
		a.promptRetry(spec)
	} else if err := a.createMarketJob(spec); err != nil {
		displayError(err)
		a.promptRetry(spec)
	}
}

func (a *Application) createMarketJob(spec *client.ChainlinkJobSpec) error {
	id, err := a.market.CreateJob(spec)
	if err != nil {
		return err
	}
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s %s\n", green("Job created:"), id.ID.String())
	return nil
}

func (a *Application) outputJSON(obj interface{}) {
	b, _ := json.Marshal(obj)
	fmt.Println(string(pretty.Color(pretty.Pretty(b), nil)))
}

func booleanInputValidation(s string) error {
	i := strings.ToLower(s)
	if i != "y" && i != "n" {
		return errors.New("Answer needs to be y/n")
	}
	return nil
}
