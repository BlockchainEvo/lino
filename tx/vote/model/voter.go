package model

import (
	types "github.com/lino-network/lino/types"
)

type Voter struct {
	Username       types.AccountKey `json:"username"`
	Deposit        types.Coin       `json:"deposit"`
	DelegatedPower types.Coin       `json:"delegated_power"`
}

type Vote struct {
	Voter  types.AccountKey `json:"voter"`
	Result bool             `json:"result"`
}

type Delegation struct {
	Delegator types.AccountKey `json:"delegator"`
	Amount    types.Coin       `json:"amount"`
}

type ValidatorPenaltyList struct {
	Validators []types.AccountKey `json:"validators"`
}
