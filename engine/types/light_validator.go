package types

import "fmt"

// LightValidator is a light version of Validator, it contains minimal information needed for application business logic.
type LightValidator struct {
	Index                     int // index in the validator set returned by RPC '/validators'
	Moniker                   string
	Address                   string
	PubKey                    string
	VotingPower               int64
	VotingPowerDisplayPercent float64 // the value is rounded so only use for display purpose
}

// GetFingerPrintAddress returns the first 6 bytes of the address.
// Tendermint/CometBFT is using this to identify the validator in pre-(commit)-votes.
func (lv LightValidator) GetFingerPrintAddress() string {
	return lv.Address[:12]
}

// LightValidators is a list of LightValidator.
//
// CONTRACT: must maintain the same order as the result from the RPC server.
type LightValidators []LightValidator

func (lvs LightValidators) TotalVotingPower() uint64 {
	var sumVotingPower uint64
	for _, lv := range lvs {
		if lv.VotingPower <= 0 {
			panic(fmt.Errorf("un-expected voting power %d at this point", lv.VotingPower))
		}

		sumVotingPower += uint64(lv.VotingPower)
	}
	return sumVotingPower
}

func (lvs LightValidators) GetLightValidatorByIndex(index int) LightValidator {
	for _, lv := range lvs {
		if lv.Index == index {
			return lv
		}
	}
	panic(fmt.Errorf("light validator with index %d not found", index))
}
