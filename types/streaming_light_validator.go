package types

type StreamingLightValidators []StreamingLightValidator

type StreamingLightValidator struct {
	Index                     int
	VotingPowerDisplayPercent float64
	Moniker                   string
}
