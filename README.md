# ConsVP
A simple utility for watching pre-vote status on Tendermint/CometBFT chains. It will print out the current pre-vote status for each validator in the validator set. Useful for watching pre-votes during an upgrade or other network event causing a slowdown.
___
A full rework of [pvtop](https://github.com/blockpane/pvtop) by [@blockpane](https://github.com/blockpane) and plus some miracles
- Live streaming mode, sharing voting information to everyone: `cvp --streaming`
- Display Block Hash fingerprint which the validator voted on
- Allow scrolling on terminal UI (thanks to [@freak12techno](https://github.com/freak12techno))
- Update binary to the latest version: `cvp --update`

## Installation
```bash
go install -v github.com/bcdevtools/consvp/cmd/cvp@latest
# Require go 1.19+
```

## Basic usage
### PVTop-like command
```bash
cvp
# => use http://localhost:26657

# For streaming mode
cvp http://example.com:26657 --streaming
# or resume streaming in case of mistakenly exit
cvp http://example.com:26657 --resume-streaming
```

```bash
cvp 19000
# => use http://localhost:19000
```

```bash
cvp https://rpc.cosmos.network
# => use https://rpc.cosmos.network
```

```bash
cvp https://rpc.example-consumer.network https://rpc.cosmos.network
# => use https://rpc.example-consumer.network as consumer network RPC endpoint
# and use https://rpc.cosmos.network as producer network RPC endpoint (typically Cosmos Hub)
```

### Voting information format
| Prevote | Precommit | Block Hash | Order | Voting Power | Moniker |
|---------|-----------|------------|-------|--------------|---------|
| ✅       | ❌         | COFF       | 1     | 11.03%       | Val1    |
| 🤷      | ❌         | ----       | 2     | 10.23%       | Val2    |
| ❌       | ❌         | ----       | 3     | 08.07%       | Val3    |
| ✅       | ✅         | COFF       | 4     | 01.15%       | Val4    |

### Check binary version
```bash
cvp --version
```

### Update binary
```bash
cvp --update
# Actually it does: go install -v github.com/bcdevtools/consvp/cmd/cvp@latest
```
