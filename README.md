# ConsVP
Similar to [pvtop](https://github.com/blockpane/pvtop), but plus some miracle

## Installation
```bash
go install -v github.com/bcdevtools/consvp/cmd/cvp@latest
```

## Basic usage
### PVTop-like command
```bash
cvp pvtop
# => use http://localhost:26657

cvp pv # alias
```

```bash
cvp pv 19000
# => use http://localhost:19000
```

```bash
cvp pv https://rpc.cosmos.network
# => use https://rpc.cosmos.network
```

```bash
cvp pv https://rpc.example-consumer.network https://rpc.cosmos.network
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

## Update binary
```bash
cvp update
# Actually it does: go install -v github.com/bcdevtools/consvp/cmd/cvp@latest
```