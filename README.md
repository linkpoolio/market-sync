# Market Sync
Synchronise job specifications from any given Chainlink node to the [LinkPool Market](https://market.link):

- Detects if there's job specifications on the provided Chainlink node that don't exist on the market.
- Ability to specify job name's and cost before added in the Market.
- Edit any job specification within the CLI to remove any secrets such as API keys.

**Important:** This tool will not sync a job unless it is confirmed first, at the risk of uploading secrets. Ensure that you edit your job specifications when prompted by the CLI if they contain sensitive information such as API keys.

## Install

Download the latest version from [releases](https://github.com/linkpoolio/market-sync/releases).

Alternatively, you can use the Docker container:
```
docker pull linkpool/market-sync
```

## Usage

Market sync is a simple CLI tool that requires you to provide your Chainlink node login credentials, and LinkPool Market 
API keys.

### Preconditions

- API key pair created on the Market, documentation found [here](https://docs.linkpool.io/docs/market_api_keys).
- `ORACLE_CONTRACT_ADDRESS` configuration variable is set in Chainlink.
- Chainlink node is already created within the Market.

### Using Flags

```
market-sync \
    -e admin@node.local \
    -p twochains \
    -u http://localhost:6688 \
    -a 31896afb-fa1c-4b30-b9a7-d7b5284cfab7 \
    -s RnscNLRnfWVRBuuRipWDRnscNLRnfWVRBuuRipWDRnscNLRnfWVRBuuRipWD
```

### Using Environment Variables

```
CHAINLINK_EMAIL=admin@node.local; \
CHAINLINK_PASSWORD=twochains; \
CHAINLINK_URL=http://localhost:6688; \
MARKET_ACCESS_KEY=31896afb-fa1c-4b30-b9a7-d7b5284cfab7; \
MARKET_SECRET_KEY=RnscNLRnfWVRBuuRipWDRnscNLRnfWVRBuuRipWDRnscNLRnfWVRBuuRipWD; \
market-sync
```

### Contributing

We welcome all contributors, please raise any issues for any feature request, issue or suggestion you may have.
