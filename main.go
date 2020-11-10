package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tcnksm/go-input"
	"github.com/ethereum/go-ethereum/common"
	"os"
	"strings"
)

const (
	ChainlinkEmailFlag         = "chainlink-email"
	ChainlinkPasswordFlag      = "chainlink-password"
	ChainlinkURLFlag           = "chainlink-url"
	ChainlinkOracleAddressFlag = "chainlink-oracle-address"
	MarketAccessKeyFlag        = "market-access-key"
	marketSecretKeyFlag        = "market-secret-key"
)

func generateCmd() *cobra.Command {
	newcmd := &cobra.Command{
		Use:  "market-sync",
		Args: cobra.MaximumNArgs(0),
		Long: `A LinkPool tool to sync a Chainlink node against the Market
All flags can be set as environment variables, eg: NODE_URL, NODE_PASSWORD`,
		Run: run,
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	newcmd.Flags().StringP(ChainlinkEmailFlag, "e", "", "chainlink node email")
	newcmd.Flags().StringP(ChainlinkPasswordFlag, "p", "", "chainlink node password")
	newcmd.Flags().StringP(ChainlinkURLFlag, "u", "", "chainlink node url")
	newcmd.Flags().StringP(ChainlinkOracleAddressFlag, "o", "", "chainlink oracle address")
	newcmd.Flags().StringP(MarketAccessKeyFlag, "a", "", "market access key")
	newcmd.Flags().StringP(marketSecretKeyFlag, "s", "", "market secret key")

	_ = newcmd.MarkFlagRequired(ChainlinkEmailFlag)
	_ = newcmd.MarkFlagRequired(ChainlinkPasswordFlag)
	_ = newcmd.MarkFlagRequired(ChainlinkURLFlag)
	_ = newcmd.MarkFlagRequired(ChainlinkOracleAddressFlag)
	_ = newcmd.MarkFlagRequired(MarketAccessKeyFlag)
	_ = newcmd.MarkFlagRequired(marketSecretKeyFlag)
	presetRequiredFlags(newcmd)
	return newcmd
}

func presetRequiredFlags(cmd *cobra.Command) {
	_ = viper.BindPFlags(cmd.Flags())
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			_ = cmd.Flags().Set(f.Name, viper.GetString(f.Name))
		}
	})
}

func run(_ *cobra.Command, _ []string) {
	yellow := color.New(color.FgYellow).SprintFunc()
	color.Blue("Starting the Market Sync CLI")
	a, err := NewApplication(&Config{
		UI:                     input.DefaultUI(),
		ChainlinkEmail:         viper.GetString(ChainlinkEmailFlag),
		ChainlinkPassword:      viper.GetString(ChainlinkPasswordFlag),
		ChainlinkURL:           viper.GetString(ChainlinkURLFlag),
		ChainlinkOracleAddress: parseOracleAddress(viper.GetString(ChainlinkOracleAddressFlag)),
		MarketAccessKey:        viper.GetString(MarketAccessKeyFlag),
		MarketSecretKey:        viper.GetString(marketSecretKeyFlag),
	})
	if err != nil {
		exit(err)
	}
	color.Green("Connected to Chainlink and the Market")

	node, err := a.MarketNode()
	if err != nil {
		exit(err)
	}
	fmt.Printf("%s %s\n", yellow("Market Node ID:"), node.ID.String())

	if err := a.SyncJobSpecs(node.ID, node.Network.ID); err != nil {
		exit(err)
	}

	color.Blue("Market Sync Complete")
	exit(nil)
}

func parseOracleAddress(address string) common.Address {
	return common.HexToAddress(address)
}

func main() {
	if err := generateCmd().Execute(); err != nil {
		exit(err)
	}
}

func displayError(err error) {
	color.Red("Error:")
	fmt.Printf("%s\n\n", err.Error())
}

func exit(err error) {
	if err == nil {
		os.Exit(0)
	}
	displayError(err)
	os.Exit(1)
}
