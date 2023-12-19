# ConsVP
Similar to [pvtop](https://github.com/blockpane/pvtop), but plus some miracle

## Installation
```bash
go install -v github.com/bcdevtools/consvp/cmd/consvpd@latest
```

## Basic usage
### PVTop-like command
```bash
consvpd pvtop
# => use http://localhost:26657

consvpd pv # alias
```

```bash
consvpd pv 19000
# => use http://localhost:19000
```

```bash
consvpd pv https://rpc.cosmos.network
# => use https://rpc.cosmos.network
```

```bash
consvpd pv https://rpc.example-consumer.network https://rpc.cosmos.network
# => use https://rpc.example-consumer.network as consumer network RPC endpoint
# and use https://rpc.cosmos.network as producer network RPC endpoint (typically Cosmos Hub)
```

## Update binary
```bash
consvpd update
# Actually it does: go install -v github.com/bcdevtools/consvp/cmd/consvpd@latest
```