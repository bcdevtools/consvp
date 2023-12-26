Today is a wonderful day 🍀 for a breakfast at your favourite restaurant.

But chain upgrade ✨? Upgrade panic 💢? Dev forgot to add upgrade handler that caused the panic 🤯?

You want your breakfast, But!!! you also want to know Pre-Vote status on-chain? You NEED this tool! Streaming Pre-Vote to your mobile and enjoying your wonderful morning ☀️ out of the house!

Not done yet, core teams patched the issue but other validators went offline 😱? You NEED this tool! Once again ❤️‍🔥! Keep streaming Pre-Vote to your mobile and enjoying your wonderful afternoon ⛅️ in the forest with your partner 🌳

This tool magically turns boring day into a wonderful day for FREE 🤩

Now read the following carefully 👇

# ConsVP
A simple utility for watching pre-vote status on Tendermint/CometBFT chains. It will print out the current pre-vote status for each validator in the validator set. Useful for watching pre-votes during an upgrade or other network event causing a slowdown.
___
A full rework of [pvtop](https://github.com/blockpane/pvtop) by [@blockpane](https://github.com/blockpane) and plus some miracles
- Live-streaming mode ([view sample](https://cvp.bcdev.tools/pvtop/sample-chain-1_AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA)), sharing pre-voting information to everyone: `cvp --streaming`
- Display Block Hash fingerprint which the validator voted on
- Allow scrolling on terminal UI (thanks to [@freak12techno](https://github.com/freak12techno))

## Installation
```bash
# Require go 1.19+

# New install
git clone https://github.com/bcdevtools/consvp ~/consvp
cd ~/consvp && make install

# Or update to the latest version
cd ~/consvp ; git pull ; make install
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

#### 🌟 We are very pleased to accompany blockchain developers around the world 🌟