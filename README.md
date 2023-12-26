# ConsVP
A simple utility for watching pre-vote status on Tendermint/CometBFT chains. It will print out the current pre-vote status for each validator in the validator set. Useful for watching pre-votes during an upgrade or other network event causing a slowdown.
___
A full rework of [pvtop](https://github.com/blockpane/pvtop) by [@blockpane](https://github.com/blockpane) and plus some miracles
- Live-streaming mode ([view sample](https://cvp.bcdev.tools/pvtop/sample-chain-1_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA)), sharing pre-voting information to everyone: `cvp --streaming`
- Display Block Hash fingerprint which the validator voted on
- Allow scrolling on terminal UI (thanks to [@freak12techno](https://github.com/freak12techno))
- Update binary to the latest version: `cvp --update`

## Installation
```bash
go install -v github.com/bcdevtools/consvp/cmd/cvp@latest
# Require go 1.19+
```

## Basic usage
```bash
cvp
# => use http://localhost:26657

# For streaming mode
cvp https://example.com:26657 --streaming
# or resume streaming in case of mistakenly exit
cvp https://example.com:26657 --resume-streaming
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

Notes:
- Default fetching consensus state is 3 seconds, can reduce to 1s by adding `-r` flag.
- In case interrupted from streaming mode, should resume instead of start a new session. Resume by adding `--resume-streaming` flag and provide the latest session id and key printed in previous run.
- Streaming session has default expiration time is 12 hours.

### Pre-voting information format
| Pre-Vote | Pre-Commit | Block Hash | Order | Voting Power | Moniker |
|----------|----------------|------------|-------|--------------|---------|
| ✅        | ❌              | C0FF       | 1     | 11.03%       | Val1    |
| 🤷       | ❌              | 0000       | 2     | 10.23%       | Val2    |
| ❌        | ❌              | ----       | 3     | 08.07%       | Val3    |
| ✅        | ✅              | C0FF       | 4     | 01.15%       | Val4    |

### Check binary version
```bash
cvp --version
```

### Update binary
```bash
cvp --update
# Actual command: go install -v github.com/bcdevtools/consvp/cmd/cvp@latest
```

#### 🌟 We are very pleased to accompany blockchain developers around the world 🌟