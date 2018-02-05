package state

import (
	"github.com/lino-network/lino/plugins/ibc"
	"github.com/lino-network/lino/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tmlibs/log"
	"testing"
	abci "github.com/tendermint/abci/types"
	ttx "github.com/lino-network/lino/types/tx"
)

//--------------------------------------------------------
// test environment is a bunch of lists of accountns

type execTest struct {
	chainID string
	store   types.KVStore
	state   *State
	accIn   types.PrivAccount
	accOut  types.PrivAccount
}

func newExecTest() *execTest {
	et := &execTest{
		chainID: "test_chain_id",
	}
	et.reset()
	return et
}

func (et *execTest) signTx(tx *ttx.SendTx, accsIn ...types.PrivAccount) {
	ttx.SignTx(et.chainID, tx, accsIn...)
}

// returns the final balance and expected balance for input and output accounts
func (et *execTest) exec(tx *ttx.SendTx, checkTx bool) (res abci.Result, inGot, inExp, outGot, outExp types.Coins) {
	initBalIn := et.state.GetAccount(et.accIn.Account.PubKey.Address()).Balance
	initBalOut := et.state.GetAccount(et.accOut.Account.PubKey.Address()).Balance

	res = ExecTx(et.state, nil, tx, checkTx, nil)

	endBalIn := et.state.GetAccount(et.accIn.Account.PubKey.Address()).Balance
	endBalOut := et.state.GetAccount(et.accOut.Account.PubKey.Address()).Balance
	decrBalInExp := tx.Outputs[0].Coins.Plus(types.Coins{tx.Fee}) //expected decrease in balance In
	return res, endBalIn, initBalIn.Minus(decrBalInExp), endBalOut, initBalOut.Plus(tx.Outputs[0].Coins)
}

func (et *execTest) acc2State(accs ...types.PrivAccount) {
	for _, acc := range accs {
		et.state.SetAccount(acc.Account.PubKey.Address(), &acc.Account)
	}
}

//reset everything. state is empty
func (et *execTest) reset() {
	et.accIn = ttx.MakeAcc("foo")
	et.accOut = ttx.MakeAcc("bar")

	et.store = types.NewMemKVStore()
	et.state = NewState(et.store)
	et.state.SetLogger(log.TestingLogger())
	et.state.SetChainID(et.chainID)

	// NOTE we dont run acc2State here
	// so we can test non-existing accounts

}

//--------------------------------------------------------

func TestGetInputs(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//nil submissions
	acc, res := getInputs(nil, nil)
	assert.True(res.IsOK(), "getInputs: error on nil submission")
	assert.Zero(len(acc), "getInputs: accounts returned on nil submission")

	//test getInputs for registered, non-registered account
	et.reset()
	inputs := ttx.Accs2TxInputs(1, et.accIn)
	acc, res = getInputs(et.state, inputs)
	assert.True(res.IsErr(), "getInputs: expected error when using getInput with non-registered Input")

	et.acc2State(et.accIn)
	acc, res = getInputs(et.state, inputs)
	assert.True(res.IsOK(), "getInputs: expected to getInput from registered Input")

	//test sending duplicate accounts
	et.reset()
	et.acc2State(et.accIn, et.accIn, et.accIn)
	inputs = ttx.Accs2TxInputs(1, et.accIn, et.accIn, et.accIn)
	acc, res = getInputs(et.state, inputs)
	assert.True(res.IsErr(), "getInputs: expected error when sending duplicate accounts")
}

func TestGetOrMakeOutputs(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//nil submissions
	acc, res := getOrMakeOutputs(nil, nil, nil)
	assert.True(res.IsOK(), "getOrMakeOutputs: error on nil submission")
	assert.Zero(len(acc), "getOrMakeOutputs: accounts returned on nil submission")

	//test sending duplicate accounts
	et.reset()
	outputs := ttx.Accs2TxOutputs(et.accIn, et.accIn, et.accIn)
	_, res = getOrMakeOutputs(et.state, nil, outputs)
	assert.True(res.IsErr(), "getOrMakeOutputs: expected error when sending duplicate accounts")

	//test sending to existing/new account
	et.reset()
	outputs1 := ttx.Accs2TxOutputs(et.accIn)
	outputs2 := ttx.Accs2TxOutputs(et.accOut)

	et.acc2State(et.accIn)
	_, res = getOrMakeOutputs(et.state, nil, outputs1)
	assert.True(res.IsOK(), "getOrMakeOutputs: error when sending to existing account")

	mapRes2, res := getOrMakeOutputs(et.state, nil, outputs2)
	assert.True(res.IsOK(), "getOrMakeOutputs: error when sending to new account")

	//test the map results
	_, map2ok := mapRes2[string(outputs2[0].Address)]
	assert.True(map2ok, "getOrMakeOutputs: account output does not contain new account map item")

}

func TestValidateInputsBasic(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//validate input basic
	inputs := ttx.Accs2TxInputs(1, et.accIn)
	res := validateInputsBasic(inputs)
	assert.True(res.IsOK(), "validateInputsBasic: expected no error on good tx input. Error: %v", res.Error())

	inputs[0].Coins[0].Amount = 0
	res = validateInputsBasic(inputs)
	assert.True(res.IsErr(), "validateInputsBasic: expected error on bad tx input")

}

func TestValidateInputsAdvanced(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//create three temp accounts for the test
	accIn1 := ttx.MakeAcc("foox")
	accIn2 := ttx.MakeAcc("fooy")
	accIn3 := ttx.MakeAcc("fooz")

	//validate inputs advanced
	tx := ttx.MakeSendTx(1, et.accOut, accIn1, accIn2, accIn3)

	et.acc2State(accIn1, accIn2, accIn3, et.accOut)
	accMap, res := getInputs(et.state, tx.Inputs)
	assert.True(res.IsOK(), "validateInputsAdvanced: error retrieving accMap. Error: %v", res.Error())
	signBytes := tx.SignBytes(et.chainID)

	//test bad case, unsigned
	totalCoins, res := validateInputsAdvanced(accMap, signBytes, tx.Inputs)
	assert.True(res.IsErr(), "validateInputsAdvanced: expected an error on an unsigned tx input")

	//test good case sgined
	et.signTx(tx, accIn1, accIn2, accIn3, et.accOut)
	totalCoins, res = validateInputsAdvanced(accMap, signBytes, tx.Inputs)
	assert.True(res.IsOK(), "validateInputsAdvanced: expected no error on good tx input. Error: %v", res.Error())

	txTotalCoins := tx.Inputs[0].Coins.
		Plus(tx.Inputs[1].Coins).
		Plus(tx.Inputs[2].Coins)

	assert.True(totalCoins.IsEqual(txTotalCoins),
		"ValidateInputsAdvanced: transaction total coins are not equal: got %v, expected %v", txTotalCoins, totalCoins)
}

func TestValidateInputAdvanced(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//validate input advanced
	tx := ttx.MakeSendTx(1, et.accOut, et.accIn)

	et.acc2State(et.accIn, et.accOut)
	signBytes := tx.SignBytes(et.chainID)

	//unsigned case
	res := validateInputAdvanced(&et.accIn.Account, signBytes, tx.Inputs[0])
	assert.True(res.IsErr(), "validateInputAdvanced: expected error on tx input without signature")

	//good signed case
	et.signTx(tx, et.accIn, et.accOut)
	res = validateInputAdvanced(&et.accIn.Account, signBytes, tx.Inputs[0])
	assert.True(res.IsOK(), "validateInputAdvanced: expected no error on good tx input. Error: %v", res.Error())

	//bad sequence case
	et.accIn.Sequence = 1
	et.signTx(tx, et.accIn, et.accOut)
	res = validateInputAdvanced(&et.accIn.Account, signBytes, tx.Inputs[0])
	assert.Equal(abci.CodeType_BaseInvalidSequence, res.Code, "validateInputAdvanced: expected error on tx input with bad sequence")
	et.accIn.Sequence = 0 //restore sequence

	//bad balance case
	et.accIn.Balance = types.Coins{{"mycoin", 2}}
	et.signTx(tx, et.accIn, et.accOut)
	res = validateInputAdvanced(&et.accIn.Account, signBytes, tx.Inputs[0])
	assert.Equal(abci.CodeType_BaseInsufficientFunds, res.Code,
		"validateInputAdvanced: expected error on tx input with insufficient funds %v", et.accIn.Sequence)
}

func TestValidateOutputsAdvanced(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//validateOutputsBasic
	tx := ttx.Accs2TxOutputs(et.accIn)
	res := validateOutputsBasic(tx)
	assert.True(res.IsOK(), "validateOutputsBasic: expected no error on good tx output. Error: %v", res.Error())

	tx[0].Coins[0].Amount = 0
	res = validateOutputsBasic(tx)
	assert.True(res.IsErr(), "validateInputBasic: expected error on bad tx output. Error: %v", res.Error())
}

func TestValidatePostAdvanced(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	acc := ttx.MakeAcc("post")

	//validatePostAdvanced
	tx := ttx.MakePostTx(1, acc)
	signBytes := tx.SignBytes(et.chainID)

	//unsigned case
	res := validatePostAdvanced(&acc.Account, signBytes, *tx)
	assert.True(res.IsErr(), "validateInputAdvanced: expected error on tx input without signature")

	//good signed case
	tx.Signature = acc.Sign(signBytes)
	res = validatePostAdvanced(&acc.Account, signBytes, *tx)
	assert.True(res.IsOK(), "validateInputAdvanced: expected no error on good tx input. Error: %v", res.Error())

	//bad sequence case
	acc.LastPost = 1
	tx.Signature = acc.Sign(signBytes)
	res = validatePostAdvanced(&acc.Account, signBytes, *tx)
	assert.Equal(abci.CodeType_BaseInvalidSequence, res.Code, "validateInputAdvanced: expected error on tx input with bad sequence")
}

func TestSumOutput(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//SumOutput
	tx := ttx.Accs2TxOutputs(et.accIn, et.accOut)
	total := sumOutputs(tx)
	assert.True(total.IsEqual(tx[0].Coins.Plus(tx[1].Coins)), "sumOutputs: total coins are not equal")
}

func TestAdjustBy(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//adjustByInputs/adjustByOutputs
	//sending transaction from accIn to accOut
	initBalIn := et.accIn.Account.Balance
	initBalOut := et.accOut.Account.Balance
	et.acc2State(et.accIn, et.accOut)

	txIn := ttx.Accs2TxInputs(1, et.accIn)
	txOut := ttx.Accs2TxOutputs(et.accOut)
	accMap, _ := getInputs(et.state, txIn)
	accMap, _ = getOrMakeOutputs(et.state, accMap, txOut)

	adjustByInputs(et.state, accMap, txIn)
	adjustByOutputs(et.state, accMap, txOut, false)

	endBalIn := accMap[string(et.accIn.Account.PubKey.Address())].Balance
	endBalOut := accMap[string(et.accOut.Account.PubKey.Address())].Balance
	decrBalIn := initBalIn.Minus(endBalIn)
	incrBalOut := endBalOut.Minus(initBalOut)

	assert.True(decrBalIn.IsEqual(txIn[0].Coins),
		"adjustByInputs: total coins are not equal. diff: %v, tx: %v", decrBalIn.String(), txIn[0].Coins.String())
	assert.True(incrBalOut.IsEqual(txOut[0].Coins),
		"adjustByInputs: total coins are not equal. diff: %v, tx: %v", incrBalOut.String(), txOut[0].Coins.String())

}

func TestSendTx(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//ExecTx
	tx := ttx.MakeSendTx(1, et.accOut, et.accIn)
	et.acc2State(et.accIn)
	et.acc2State(et.accOut)
	et.signTx(tx, et.accIn)

	//Bad Balance
	et.accIn.Balance = types.Coins{{"mycoin", 2}}
	et.acc2State(et.accIn)
	res, _, _, _, _ := et.exec(tx, true)
	assert.True(res.IsErr(), "ExecTx/Bad CheckTx: Expected error return from ExecTx, returned: %v", res)

	res, balIn, balInExp, balOut, balOutExp := et.exec(tx, false)
	assert.True(res.IsErr(), "ExecTx/Bad DeliverTx: Expected error return from ExecTx, returned: %v", res)
	assert.False(balIn.IsEqual(balInExp),
		"ExecTx/Bad DeliverTx: balance shouldn't be equal for accIn: got %v, expected: %v", balIn, balInExp)
	assert.False(balOut.IsEqual(balOutExp),
		"ExecTx/Bad DeliverTx: balance shouldn't be equal for accOut: got %v, expected: %v", balOut, balOutExp)

	//Regular CheckTx
	et.reset()
	et.acc2State(et.accIn)
	et.acc2State(et.accOut)
	res, _, _, _, _ = et.exec(tx, true)
	assert.True(res.IsOK(), "ExecTx/Good CheckTx: Expected OK return from ExecTx, Error: %v", res)

	//Regular DeliverTx
	et.reset()
	et.acc2State(et.accIn)
	et.acc2State(et.accOut)
	res, balIn, balInExp, balOut, balOutExp = et.exec(tx, false)
	assert.True(res.IsOK(), "ExecTx/Good DeliverTx: Expected OK return from ExecTx, Error: %v", res)
	assert.True(balIn.IsEqual(balInExp),
		"ExecTx/good DeliverTx: unexpected change in input balance, got: %v, expected: %v", balIn, balInExp)
	assert.True(balOut.IsEqual(balOutExp),
		"ExecTx/good DeliverTx: unexpected change in output balance, got: %v, expected: %v", balOut, balOutExp)
}

func TestSendTxIBC(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//ExecTx
	chainID2 := "otherchain"
	tx := ttx.MakeSendTx(1, et.accOut, et.accIn)
	dstAddress := tx.Outputs[0].Address
	tx.Outputs[0].Address = []byte(chainID2 + "/" + string(tx.Outputs[0].Address))
	et.acc2State(et.accIn)
	et.signTx(tx, et.accIn)

	//Regular DeliverTx
	et.reset()
	et.acc2State(et.accIn)

	initBalIn := et.state.GetAccount(et.accIn.Account.PubKey.Address()).Balance

	res := ExecTx(et.state, nil, tx, false, nil)

	balIn := et.state.GetAccount(et.accIn.Account.PubKey.Address()).Balance
	decrBalInExp := tx.Outputs[0].Coins.Plus(types.Coins{tx.Fee}) //expected decrease in balance In
	balInExp := initBalIn.Minus(decrBalInExp)

	assert.True(res.IsOK(), "ExecTx/Good DeliverTx: Expected OK return from ExecTx, Error: %v", res)
	assert.True(balIn.IsEqual(balInExp),
		"ExecTx/good DeliverTx: unexpected change in input balance, got: %v, expected: %v", balIn, balInExp)

	packet, err := ibc.GetIBCPacket(et.state, et.chainID, chainID2, 0)
	assert.Nil(err)

	assert.Equal(packet.SrcChainID, et.chainID)
	assert.Equal(packet.DstChainID, chainID2)
	assert.Equal(packet.Sequence, uint64(0))
	assert.Equal(packet.Type, "coin")

	coins, ok := packet.Payload.(ibc.CoinsPayload)
	assert.True(ok)
	assert.Equal(coins.Coins, tx.Outputs[0].Coins)
	assert.EqualValues(coins.Address, dstAddress)
}

func TestPostTx(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	tx := ttx.MakePostTx(1, et.accOut)
	signBytes := tx.SignBytes(et.chainID)
	tx.Signature = et.accOut.Sign(signBytes)
	//ExecTx
	et.acc2State(et.accOut)

	initPostSeq := et.state.GetAccount(et.accOut.Account.PubKey.Address()).LastPost

	// Test seq equal to 1
	res := ExecTx(et.state, nil, tx, false, nil)
	assert.True(res.IsOK(), "ExecTx/Good PostTx: Expected OK return from ExecTx, Error: %v", res)
	endPostSeq := et.state.GetAccount(et.accOut.Account.PubKey.Address()).LastPost
	assert.Equal(endPostSeq, initPostSeq + 1)

	// Test seq larger than 1
	tx = ttx.MakePostTx(2, et.accOut)
	signBytes = tx.SignBytes(et.chainID)
	tx.Signature = et.accOut.Sign(signBytes)
	initPostSeq = et.state.GetAccount(et.accOut.Account.PubKey.Address()).LastPost

	res = ExecTx(et.state, nil, tx, false, nil)
	assert.True(res.IsOK(), "ExecTx/Good PostTx: Expected OK return from ExecTx, Error: %v", res)
	endPostSeq = et.state.GetAccount(et.accOut.Account.PubKey.Address()).LastPost
	assert.Equal(endPostSeq, initPostSeq + 1)

	// First post
	acc := ttx.MakeAcc("post")
	tx = ttx.MakePostTx(1, acc)
	signBytes = tx.SignBytes(et.chainID)
	tx.Signature = acc.Sign(signBytes)
	preAcc := et.state.GetAccount(acc.Account.PubKey.Address())

	assert.True(preAcc == nil)
	res = ExecTx(et.state, nil, tx, false, nil)
	assert.True(res.IsOK(), "ExecTx/Good PostTx: Expected OK return from ExecTx, Error: %v", res)
	endPostSeq = et.state.GetAccount(acc.Account.PubKey.Address()).LastPost
	assert.Equal(1, endPostSeq)

	// Test comment
	tx = ttx.MakePostTx(2, acc)
	tx.Parent = types.PostID(acc.Account.PubKey.Address(), 1)
	signBytes = tx.SignBytes(et.chainID)
	tx.Signature = acc.Sign(signBytes)

	res = ExecTx(et.state, nil, tx, false, nil)
	assert.True(res.IsOK(), "ExecTx/Good PostTx: Expected OK return from ExecTx, Error: %v", res)
	endPostSeq = et.state.GetAccount(acc.Account.PubKey.Address()).LastPost
	assert.Equal(2, endPostSeq)

	// Invalid seq no
	tx = ttx.MakePostTx(100, et.accOut)
	signBytes = tx.SignBytes(et.chainID)
	tx.Signature = et.accOut.Sign(signBytes)
	initPostSeq = et.state.GetAccount(et.accOut.Account.PubKey.Address()).LastPost

	res = ExecTx(et.state, nil, tx, false, nil)
	assert.Equal(abci.CodeType_BaseInvalidSequence, res.Code, "ExecTx/Bad PostTx: expected error on tx input with bad sequence")
	endPostSeq = et.state.GetAccount(et.accOut.Account.PubKey.Address()).LastPost
	assert.Equal(endPostSeq, initPostSeq)

	// Unsigned post
	tx = ttx.MakePostTx(3, et.accOut)
	signBytes = tx.SignBytes(et.chainID)
	initPostSeq = et.state.GetAccount(et.accOut.Account.PubKey.Address()).LastPost

	res = ExecTx(et.state, nil, tx, false, nil)
	assert.Equal(abci.ErrBaseInvalidSignature.Code, res.Code, "ExecTx/Bad PostTx: expected error on tx input with bad sequence")
	endPostSeq = et.state.GetAccount(et.accOut.Account.PubKey.Address()).LastPost
	assert.Equal(endPostSeq, initPostSeq)

	// Invalid comment
	tx = ttx.MakePostTx(2, acc)
	tx.Parent = []byte("parent")
	signBytes = tx.SignBytes(et.chainID)
	tx.Signature = acc.Sign(signBytes)

	res = ExecTx(et.state, nil, tx, false, nil)
	assert.Equal(abci.ErrBaseUnknownAddress.Code, res.Code, "ExecTx/Bad PostTx: expected error on tx input with bad sequence")
	endPostSeq = et.state.GetAccount(acc.Account.PubKey.Address()).LastPost
	assert.Equal(2, endPostSeq)
}

func TestDonateTx(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	pstx := ttx.MakePostTx(1, et.accOut)
	signBytes := pstx.SignBytes(et.chainID)
	pstx.Signature = et.accOut.Sign(signBytes)
	et.acc2State(et.accOut)
	initOutputBalance := et.state.GetAccount(et.accOut.Account.PubKey.Address()).Balance
	initPostSeq := et.state.GetAccount(et.accOut.Account.PubKey.Address()).LastPost

	// Test seq equal to 1
	res := ExecTx(et.state, nil, pstx, false, nil)
	assert.True(res.IsOK(), "ExecTx/Good PostTx: Expected OK return from ExecTx, Error: %v", res)
	endPostSeq := et.state.GetAccount(et.accOut.Account.PubKey.Address()).LastPost
	assert.Equal(endPostSeq, initPostSeq + 1)

	// Test valid donate
	var testCost int64 = 5
	var testFee int64 = 1
	var initBalance int64 = 7

	dtx := ttx.MakeDonateTx(1, testCost, testFee, types.PostID(et.accOut.Account.PubKey.Address(), 1), et.accIn)
	et.acc2State(et.accIn)
	et.acc2State(et.accOut)
	dtxSignBytes := dtx.SignBytes(et.chainID)
	dtx.Input.Signature = et.accIn.Sign(dtxSignBytes)

	// Execute donate
	initOutputBalance = et.state.GetAccount(et.accOut.Account.PubKey.Address()).Balance
	assert.True(initOutputBalance.IsEqual(types.Coins{{"mycoin", initBalance}}),
			"ExecTx/Valid DeliverTx: balance should be equal for accIn: got %v, expected: %v", initOutputBalance, initBalance)
	res = ExecTx(et.state, nil, dtx, true, nil)
	assert.True(res.IsOK(), "ExecTx/Good DonateTx: Expected OK return from ExecTx, Error: %v", res)
	endInputBalance := et.state.GetAccount(et.accIn.Account.PubKey.Address()).Balance
	assert.True(endInputBalance.IsEqual(types.Coins{{"mycoin", initBalance - testCost}}),
			"ExecTx/Valid DeliverTx: balance should be equal for accIn: got %v, expected: %v", endInputBalance, initBalance - testCost)

	endOutputBalance := et.state.GetAccount(et.accOut.Account.PubKey.Address()).Balance
	assert.True(endOutputBalance.IsEqual(types.Coins{{"mycoin", initBalance}}),
			"ExecTx/Valid DeliverTx: balance should be equal for accIn: got %v, expected: %v", endOutputBalance, initBalance)

	et.reset()
	et.acc2State(et.accIn)
	et.acc2State(et.accOut)
	res = ExecTx(et.state, nil, pstx, false, nil)
	res = ExecTx(et.state, nil, dtx, false, nil)
	assert.True(res.IsOK(), "ExecTx/Good DonateTx: Expected OK return from ExecTx, Error: %v", res)
	endInputBalance = et.state.GetAccount(et.accIn.Account.PubKey.Address()).Balance
	assert.True(endInputBalance.IsEqual(types.Coins{{"mycoin", initBalance - testCost}}),
			"ExecTx/Valid DeliverTx: balance should be equal for accIn: got %v, expected: %v", endInputBalance, initBalance - testCost)

	endOutputBalance = et.state.GetAccount(et.accOut.Account.PubKey.Address()).Balance
	assert.True(endOutputBalance.IsEqual(types.Coins{{"mycoin", initBalance + testCost - testFee}}),
			"ExecTx/Valid DeliverTx: balance should be equal for accIn: got %v, expected: %v", endOutputBalance, initBalance)
}

func TestLikeTx(t *testing.T) {
	// set up environment
	assert := assert.New(t)
	et := newExecTest()
	seq := 1
	et.acc2State(et.accOut)
	pstTx := ttx.MakePostTx(seq, et.accOut)
	pstID := types.PostID(et.accOut.Account.PubKey.Address(), seq)
	pstSignBytes := pstTx.SignBytes(et.chainID)
	pstTx.Signature = et.accOut.Sign(pstSignBytes)
	res := ExecTx(et.state, nil, pstTx, false, nil)
	assert.True(res.IsOK(), "ExecTx/Good PostTx: Expected OK return from ExecTx, Error: %v", res)

	// Valid Like
	tx1 := ttx.MakeLikeTx(10000, et.accOut, pstID, true, true)
	signBytes := tx1.SignBytes(et.chainID)
	tx1.Signature = et.accOut.Sign(signBytes)
	rst := ExecTx(et.state, nil, tx1, false, nil)
	assert.True(rst.IsOK(), "LikeTx error: %v", rst)
	likes := et.state.GetLikesByPostId(pstID)
	assert.Equal(1, len(likes), "Unexpeted Likes array: %v", likes)
	assert.Equal(et.accOut.PubKey.Address(), likes[0].From)
	assert.Equal(pstID, likes[0].To)
	assert.Equal(10000, likes[0].Weight)

	// Invalid post Id
	tx1 = ttx.MakeLikeTx(0, et.accOut, []byte("wrong_pid"), true, false)
	signBytes = tx1.SignBytes(et.chainID)
	tx1.Signature = et.accOut.Sign(signBytes)
	rst = ExecTx(et.state, nil, tx1, false, nil)
	assert.Equal(abci.ErrBaseUnknownAddress.Code, rst.Code, "ExecTx/Bad PostTx: expected error on tx input with bad sequence")
	likes = et.state.GetLikesByPostId(pstID)
	assert.Equal(1, len(likes), "Unexpeted Likes array: %v", likes)

	// Valid Dislike
	tx1 = ttx.MakeLikeTx(-10000, et.accOut, pstID, false, false)
	signBytes = tx1.SignBytes(et.chainID)
	tx1.Signature = et.accOut.Sign(signBytes)
	rst = ExecTx(et.state, nil, tx1, false, nil)
	assert.True(rst.IsOK(), "LikeTx error: %v", rst)
	likes = et.state.GetLikesByPostId(pstID)
	assert.Equal(1, len(likes), "Unexpeted Likes array: %v", likes)
	assert.Equal(pstID, likes[0].To)
	assert.Equal(-10000, likes[0].Weight)
}
