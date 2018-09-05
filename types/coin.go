package types

import (
	"fmt"
	"math"

	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// LNO - exposed type
type LNO = string

var (
	// LowerBoundRat - the lower bound of Rat
	LowerBoundRat = sdk.NewRat(1, Decimals)
	// UpperBoundRat - the upper bound of Rat
	UpperBoundRat = sdk.NewRat(math.MaxInt64 / Decimals)
)

// Coin - 10^5 Coin = 1 LNO
type Coin struct {
	// Amount *big.Int `json:"amount"`
	Amount sdk.Int `json:"amount"`
}

// NewCoinFromInt64 - return int64 amount of Coin
func NewCoinFromInt64(amount int64) Coin {
	// return Coin{big.NewInt(amount)}
	return Coin{sdk.NewInt(amount)}
}

// NewCoinFromBigInt - return big.Int amount of Coin
func NewCoinFromBigInt(amount *big.Int) Coin {
	sdkInt := sdk.NewIntFromBigInt(amount)
	return Coin{sdkInt}
}

// NewCoinFromString - return string amount of Coin
func NewCoinFromString(amount string) (Coin, bool) {
	res, ok := sdk.NewIntFromString(amount)
	return Coin{res}, ok
}

// LinoToCoin - convert 1 LNO to 10^5 Coin
func LinoToCoin(lino LNO) (Coin, sdk.Error) {
	rat, err := sdk.NewRatFromDecimal(lino, NewRatFromDecimalPrecision)
	if err != nil {
		return NewCoinFromInt64(0), ErrInvalidCoins("Illegal LNO")
	}
	if rat.GT(UpperBoundRat) {
		return NewCoinFromInt64(0), ErrInvalidCoins("LNO overflow")
	}
	if rat.LT(LowerBoundRat) {
		return NewCoinFromInt64(0), ErrInvalidCoins("LNO can't be less than lower bound")
	}
	return RatToCoin(rat.Mul(sdk.NewRat(Decimals, 1))), nil
}

var (
	zero  = big.NewInt(0)
	one   = big.NewInt(1)
	two   = big.NewInt(2)
	five  = big.NewInt(5)
	nFive = big.NewInt(-5)
	ten   = big.NewInt(10)
)

// RatToCoin - convert sdk.Rat to LNO coin
func RatToCoin(rat sdk.Rat) Coin {
	//return Coin{rat.EvaluateBig()}

	// num := rat.Num()
	// denom := rat.Denom()

	// d, rem := new(big.Int), new(big.Int)
	// d.QuoRem(num, denom, rem)
	// if rem.Cmp(zero) == 0 { // is the remainder zero
	// 	return NewCoinFromBigInt(d)
	// }

	// // evaluate the remainder using bankers rounding
	// tenNum := new(big.Int).Mul(num, ten)
	// tenD := new(big.Int).Mul(d, ten)
	// remainderDigit := new(big.Int).Sub(new(big.Int).Quo(tenNum, denom), tenD) // get the first remainder digit
	// isFinalDigit := (new(big.Int).Rem(tenNum, denom).Cmp(zero) == 0)          // is this the final digit in the remainder?

	// switch {
	// case isFinalDigit && (remainderDigit.Cmp(five) == 0 || remainderDigit.Cmp(nFive) == 0):
	// 	dRem2 := new(big.Int).Rem(d, two)
	// 	return NewCoinFromBigInt(new(big.Int).Add(d, dRem2)) // always rounds to the even number
	// case remainderDigit.Cmp(five) != -1: //remainderDigit >= 5:
	// 	d.Add(d, one)
	// case remainderDigit.Cmp(nFive) != 1: //remainderDigit <= -5:
	// 	d.Sub(d, one)
	// }
	return NewCoinFromBigInt(rat.EvaluateBig())
}

// ToRat - convert Coin to sdk.Rat
func (coin Coin) ToRat() sdk.Rat {
	return sdk.NewRatFromBigInt(coin.Amount.BigInt())
}

// ToInt64 - convert Coin to int64
func (coin Coin) ToInt64() (int64, sdk.Error) {
	if !coin.Amount.BigInt().IsInt64() {
		return 0, ErrAmountOverflow()
	}
	return coin.Amount.BigInt().Int64(), nil
}

// String - provides a human-readable representation of a coin
func (coin Coin) String() string {
	return fmt.Sprintf("coin:%v", coin.Amount)
}

// IsZero - returns if this represents no money
func (coin Coin) IsZero() bool {
	return coin.Amount.Sign() == 0
}

// IsGT - returns true if the receiver is greater value
func (coin Coin) IsGT(other Coin) bool {
	return coin.Amount.GT(other.Amount)
}

// IsGTE - returns true if they are the same type and the receiver is
// an equal or greater value
func (coin Coin) IsGTE(other Coin) bool {
	return coin.Amount.GT(other.Amount) || coin.Amount.Equal(other.Amount)
}

// IsEqual - returns true if the two sets of Coins have the same value
func (coin Coin) IsEqual(other Coin) bool {
	return coin.Amount.Equal(other.Amount)
}

// IsPositive - returns true if coin amount is positive
func (coin Coin) IsPositive() bool {
	return coin.Amount.Sign() > 0
}

// IsNotNegative - returns true if coin amount is not negative
func (coin Coin) IsNotNegative() bool {
	return coin.Amount.Sign() >= 0
}

// Plus - Adds amounts of two coins with same denom
func (coin Coin) Plus(coinB Coin) Coin {
	r := coin.Amount.Add(coinB.Amount)
	return Coin{r}
}

// Minus - Subtracts amounts of two coins with same denom
func (coin Coin) Minus(coinB Coin) Coin {
	sdkInt := coin.Amount.Sub(coinB.Amount)
	return Coin{sdkInt}
}
