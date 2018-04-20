package vote

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gmodel "github.com/lino-network/lino/tx/global/model"
	model "github.com/lino-network/lino/tx/vote/model"
	"github.com/lino-network/lino/types"
	"github.com/stretchr/testify/assert"
)

func TestDecideProposal(t *testing.T) {
	ctx, am, vm, gm := setupTest(t, 0)
	vm.InitGenesis(ctx)
	user1 := createTestAccount(ctx, am, "user1")
	user2 := createTestAccount(ctx, am, "user2")
	user3 := createTestAccount(ctx, am, "user3")
	user4 := createTestAccount(ctx, am, "user4")

	vm.AddVoter(ctx, user1, types.VoterMinDeposit.Plus(types.NewCoin(20)))
	vm.AddVoter(ctx, user2, types.VoterMinDeposit.Plus(types.NewCoin(30)))
	vm.AddVoter(ctx, user3, types.VoterMinDeposit.Plus(types.NewCoin(50)))
	vm.AddVoter(ctx, user4, types.VoterMinDeposit.Plus(types.NewCoin(10)))

	globalStorage := gmodel.NewGlobalStorage(TestGlobalKVStoreKey)
	prevAllocation, _ := globalStorage.GetGlobalAllocation(ctx)
	e := DecideProposalEvent{}

	des1 := &model.ChangeParameterDescription{
		InfraAllocation: sdk.NewRat(50, 100),
	}
	des2 := &model.ChangeParameterDescription{
		InfraAllocation: sdk.NewRat(80, 100),
	}
	id1, _ := vm.AddProposal(ctx, types.AccountKey("c1"), des1)
	id2, _ := vm.AddProposal(ctx, types.AccountKey("c2"), des2)

	cases := []struct {
		decideProposal        bool
		voter                 types.AccountKey
		proposalID            types.ProposalKey
		voterRes              bool
		expectOngoingProposal []types.ProposalKey
		expectDecidedProposal []types.ProposalKey
		expectInfraAllocation sdk.Rat
	}{
		{false, user1, id1, true, []types.ProposalKey{id1, id2}, []types.ProposalKey{}, prevAllocation.InfraAllocation},
		{false, user2, id1, false, []types.ProposalKey{id1, id2}, []types.ProposalKey{}, prevAllocation.InfraAllocation},
		{true, types.AccountKey(""), id1, false, []types.ProposalKey{id2}, []types.ProposalKey{id1}, prevAllocation.InfraAllocation},
		{false, user1, id2, true, []types.ProposalKey{id2}, []types.ProposalKey{id1}, prevAllocation.InfraAllocation},
		{false, user2, id1, true, []types.ProposalKey{id2}, []types.ProposalKey{id1}, prevAllocation.InfraAllocation},
		{false, user4, id1, true, []types.ProposalKey{id2}, []types.ProposalKey{id1}, prevAllocation.InfraAllocation},
		{false, user3, id1, false, []types.ProposalKey{id2}, []types.ProposalKey{id1}, prevAllocation.InfraAllocation},
		{true, types.AccountKey(""), id2, false, []types.ProposalKey{}, []types.ProposalKey{id1, id2}, des2.InfraAllocation},
	}

	for _, cs := range cases {
		if cs.decideProposal {
			e.Execute(ctx, vm, am, gm)
		} else {
			vm.AddVote(ctx, cs.proposalID, cs.voter, cs.voterRes)
		}

		lst, _ := vm.storage.GetProposalList(ctx)
		curAllocation, _ := globalStorage.GetGlobalAllocation(ctx)

		assert.Equal(t, cs.expectInfraAllocation, curAllocation.InfraAllocation)
		assert.Equal(t, cs.expectOngoingProposal, lst.OngoingProposal)
		assert.Equal(t, cs.expectDecidedProposal, lst.PastProposal)
	}
}

func TestForceValidatorVote(t *testing.T) {
	ctx, am, vm, gm := setupTest(t, 0)
	vm.InitGenesis(ctx)
	user1 := createTestAccount(ctx, am, "user1")
	user2 := createTestAccount(ctx, am, "user2")

	vm.AddVoter(ctx, user1, types.VoterMinDeposit.Plus(types.NewCoin(20)))
	vm.AddVoter(ctx, user2, types.VoterMinDeposit.Plus(types.NewCoin(30)))
	ctx = WithOncallValidators(ctx, []types.AccountKey{user2, user1})

	e := DecideProposalEvent{}

	des1 := &model.ChangeParameterDescription{
		InfraAllocation: sdk.NewRat(50, 100),
	}
	id1, _ := vm.AddProposal(ctx, types.AccountKey("c1"), des1)
	cases := []struct {
		decideProposal    bool
		voter             types.AccountKey
		proposalID        types.ProposalKey
		voterRes          bool
		expectPenaltyList []types.AccountKey
	}{
		{false, user1, id1, true, []types.AccountKey{}},
		{true, types.AccountKey(""), id1, true, []types.AccountKey{user2}},
	}

	for _, cs := range cases {
		if cs.decideProposal {
			e.Execute(ctx, vm, am, gm)
		} else {
			vm.AddVote(ctx, cs.proposalID, cs.voter, cs.voterRes)
		}
		lst, _ := vm.GetValidatorPenaltyList(ctx)
		assert.Equal(t, cs.expectPenaltyList, lst.Validators)
	}
}