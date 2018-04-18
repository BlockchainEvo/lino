package app

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lino-network/lino/genesis"
	"github.com/lino-network/lino/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"
)

var (
	priv1 = crypto.GenPrivKeyEd25519()
	addr1 = priv1.PubKey().Address()
	priv2 = crypto.GenPrivKeyEd25519()
	addr2 = priv2.PubKey().Address()
	priv3 = crypto.GenPrivKeyEd25519()
	addr3 = priv3.PubKey().Address()
	priv4 = crypto.GenPrivKeyEd25519()
	addr4 = priv4.PubKey().Address()
	priv5 = crypto.GenPrivKeyEd25519()
	addr5 = priv3.PubKey().Address()
	priv6 = crypto.GenPrivKeyEd25519()
	addr6 = priv4.PubKey().Address()

	genesisTotalLino    int64      = 10000000000
	genesisTotalCoin    types.Coin = types.NewCoin(10000000000 * types.Decimals)
	genesisAccount      string     = "Lino"
	growthRate          sdk.Rat    = sdk.NewRat(98, 1000)
	validatorAllocation sdk.Rat    = sdk.NewRat(10, 100)
)

func loggerAndDBs() (log.Logger, map[string]dbm.DB) {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
	dbs := map[string]dbm.DB{
		"acc":       dbm.NewMemDB(),
		"post":      dbm.NewMemDB(),
		"val":       dbm.NewMemDB(),
		"vote":      dbm.NewMemDB(),
		"infra":     dbm.NewMemDB(),
		"developer": dbm.NewMemDB(),
		"global":    dbm.NewMemDB(),
	}
	return logger, dbs
}

func newLinoBlockchain(t *testing.T) *LinoBlockchain {
	logger, dbs := loggerAndDBs()
	lb := NewLinoBlockchain(logger, dbs)
	result, err := genesis.GetDefaultGenesis(priv1.PubKey(), priv2.PubKey())
	assert.Nil(t, err)

	vals := []abci.Validator{}
	lb.InitChain(abci.RequestInitChain{vals, json.RawMessage(result)})
	lb.Commit()
	return lb
}

func TestGenesisAcc(t *testing.T) {
	logger, dbs := loggerAndDBs()
	lb := NewLinoBlockchain(logger, dbs)

	accs := []struct {
		genesisAccountName string
		numOfLino          int64
		pubKey             crypto.PubKey
		isValidator        bool
		valPubKey          crypto.PubKey
	}{
		{"Lino", 9000000000, priv1.PubKey(), true, priv2.PubKey()},
		{"Genesis", 500000000, priv3.PubKey(), true, priv4.PubKey()},
		{"NonValidator", 500000000, priv5.PubKey(), false, priv6.PubKey()},
	}

	globalState := genesis.GlobalState{
		TotalLino:                genesisTotalLino,
		GrowthRate:               sdk.NewRat(98, 1000),
		InfraAllocation:          sdk.NewRat(20, 100),
		ContentCreatorAllocation: sdk.NewRat(50, 100),
		DeveloperAllocation:      sdk.NewRat(20, 100),
		ValidatorAllocation:      sdk.NewRat(10, 100),
		ConsumptionFrictionRate:  sdk.NewRat(1, 100),
		FreezingPeriodHr:         24 * 7,
	}

	genesisState := genesis.GenesisState{
		Accounts:    []genesis.GenesisAccount{},
		GlobalState: globalState,
	}
	for _, acc := range accs {
		genesisAcc := genesis.GenesisAccount{
			Name:        acc.genesisAccountName,
			Lino:        acc.numOfLino,
			PubKey:      acc.pubKey,
			IsValidator: acc.isValidator,
			ValPubKey:   acc.valPubKey,
		}
		genesisState.Accounts = append(genesisState.Accounts, genesisAcc)
	}

	result, err := genesis.GetGenesisJson(genesisState)
	assert.Nil(t, err)

	vals := []abci.Validator{}
	lb.InitChain(abci.RequestInitChain{vals, json.RawMessage(result)})
	lb.Commit()

	ctx := lb.BaseApp.NewContext(true, abci.Header{})
	for _, acc := range accs {
		expectBalance, err := types.LinoToCoin(types.LNO(sdk.NewRat(acc.numOfLino)))
		assert.Nil(t, err)
		if acc.isValidator {
			expectBalance = expectBalance.Minus(
				types.ValidatorMinCommitingDeposit.Plus(types.ValidatorMinVotingDeposit))
		}
		balance, err :=
			lb.accountManager.GetBankBalance(ctx, types.AccountKey(acc.genesisAccountName))
		assert.Nil(t, err)
		assert.Equal(t, expectBalance, balance)
	}

	// reload app and ensure the account is still there
	lb = NewLinoBlockchain(logger, dbs)
	ctx = lb.BaseApp.NewContext(true, abci.Header{})
	for _, acc := range accs {
		expectBalance, err := types.LinoToCoin(types.LNO(sdk.NewRat(acc.numOfLino)))
		assert.Nil(t, err)
		if acc.isValidator {
			expectBalance = expectBalance.Minus(
				types.ValidatorMinCommitingDeposit.Plus(types.ValidatorMinVotingDeposit))
		}
		balance, err :=
			lb.accountManager.GetBankBalance(ctx, types.AccountKey(acc.genesisAccountName))
		assert.Nil(t, err)
		assert.Equal(t, expectBalance, balance)
	}
}

func TestDistributeInflationToValidators(t *testing.T) {
	logger, dbs := loggerAndDBs()
	lb := NewLinoBlockchain(logger, dbs)

	globalState := genesis.GlobalState{
		TotalLino:                genesisTotalLino,
		GrowthRate:               sdk.NewRat(98, 1000),
		InfraAllocation:          sdk.NewRat(20, 100),
		ContentCreatorAllocation: sdk.NewRat(50, 100),
		DeveloperAllocation:      sdk.NewRat(20, 100),
		ValidatorAllocation:      sdk.NewRat(10, 100),
		ConsumptionFrictionRate:  sdk.NewRat(1, 100),
		FreezingPeriodHr:         24 * 7,
	}

	genesisState := genesis.GenesisState{
		Accounts:    []genesis.GenesisAccount{},
		GlobalState: globalState,
	}

	// Generate 21 validators
	LNOPerValidator := int64(100000000)
	for i := 0; i < 21; i++ {
		privKey := crypto.GenPrivKeyEd25519()
		valPrivKey := crypto.GenPrivKeyEd25519()
		genesisAcc := genesis.GenesisAccount{
			Name:        "validator" + strconv.Itoa(i),
			Lino:        LNOPerValidator,
			PubKey:      privKey.PubKey(),
			IsValidator: true,
			ValPubKey:   valPrivKey.PubKey(),
		}
		genesisState.Accounts = append(genesisState.Accounts, genesisAcc)
	}

	result, err := genesis.GetGenesisJson(genesisState)
	assert.Nil(t, err)

	vals := []abci.Validator{}
	lb.InitChain(abci.RequestInitChain{vals, json.RawMessage(result)})
	lb.Commit()
	baseTime := time.Now().Unix()
	remainValidatorPool := types.RatToCoin(genesisTotalCoin.ToRat().Mul(growthRate).Mul(validatorAllocation))
	expectBalance := types.NewCoin(LNOPerValidator * types.Decimals).Minus(
		types.ValidatorMinCommitingDeposit.Plus(types.ValidatorMinVotingDeposit))

	testPastMinutes := int64(0)
	for i := baseTime; i < baseTime+3600*20; i += 50 {
		lb.BeginBlock(abci.RequestBeginBlock{
			Header: abci.Header{
				ChainID: "Lino", Time: baseTime + i}})
		lb.EndBlock(abci.RequestEndBlock{})
		lb.Commit()
		// simulate app
		if (baseTime+int64(i)-lb.chainStartTime)/int64(60) > testPastMinutes {
			testPastMinutes += 1
			if testPastMinutes%60 == 0 {
				// hourly inflation
				inflationForValidator :=
					types.RatToCoin(remainValidatorPool.ToRat().Mul(
						sdk.NewRat(1, types.HoursPerYear-lb.pastMinutes/60+1)))
				remainValidatorPool = remainValidatorPool.Minus(inflationForValidator)
				// expectBalance for all validators
				expectBalance = expectBalance.Plus(types.RatToCoin(inflationForValidator.ToRat().Quo(sdk.NewRat(21))))
				ctx := lb.BaseApp.NewContext(true, abci.Header{})
				for i := 0; i < 21; i++ {
					balance, err :=
						lb.accountManager.GetBankBalance(
							ctx, types.AccountKey("validator"+strconv.Itoa(i)))
					assert.Nil(t, err)
					assert.Equal(t, expectBalance, balance)
				}
			}
		}
	}
}
