package account

import (
	"testing"
	"time"

	"github.com/lino-network/lino/types"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"
	dbm "github.com/tendermint/tmlibs/db"
)

var (
	TestAccountKVStoreKey = sdk.NewKVStoreKey("account")

	l0    = types.LNO(sdk.NewRat(0))
	l100  = types.LNO(sdk.NewRat(100))
	l200  = types.LNO(sdk.NewRat(200))
	l1600 = types.LNO(sdk.NewRat(1600))
	l1800 = types.LNO(sdk.NewRat(1800))
	l1900 = types.LNO(sdk.NewRat(1900))
	l2000 = types.LNO(sdk.NewRat(2000))
	c0    = types.NewCoin(0)
	c100  = types.NewCoin(100 * types.Decimals)
	c200  = types.NewCoin(200 * types.Decimals)
	c300  = types.NewCoin(300 * types.Decimals)
	c500  = types.NewCoin(500 * types.Decimals)
	c600  = types.NewCoin(600 * types.Decimals)
	c1000 = types.NewCoin(1000 * types.Decimals)
	c1500 = types.NewCoin(1500 * types.Decimals)
	c1600 = types.NewCoin(1600 * types.Decimals)
	c1800 = types.NewCoin(1800 * types.Decimals)
	c1900 = types.NewCoin(1900 * types.Decimals)
	c2000 = types.NewCoin(2000 * types.Decimals)

	coin0   = types.NewCoin(0)
	coin1   = types.NewCoin(1)
	coin50  = types.NewCoin(50)
	coin100 = types.NewCoin(100)
	coin200 = types.NewCoin(200)
)

func setupTest(t *testing.T, height int64) (sdk.Context, AccountManager) {
	ctx := getContext(height)
	accManager := NewAccountManager(TestAccountKVStoreKey)
	return ctx, accManager
}

func getContext(height int64) sdk.Context {
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(TestAccountKVStoreKey, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()

	return sdk.NewContext(ms, abci.Header{ChainID: "Lino", Height: height, Time: time.Now().Unix()}, false, nil)
}

func createTestAccount(ctx sdk.Context, am AccountManager, username string) crypto.PrivKey {
	priv := crypto.GenPrivKeyEd25519()
	am.AddCoinToAddress(ctx, priv.PubKey().Address(), types.NewCoin(100*types.Decimals))
	am.CreateAccount(ctx, types.AccountKey(username), priv.PubKey(), types.NewCoin(0))
	return priv.Wrap()
}
