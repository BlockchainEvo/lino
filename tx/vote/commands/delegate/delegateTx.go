package delegate

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/lino-network/lino/client"
	"github.com/lino-network/lino/tx/vote"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/wire"
)

const (
	FlagUsername = "username"
	FlagVoter    = "voter"
	FlagAmount   = "amount"
)

// DelegateTxCmd will create a send tx and sign it with the given key
func DelegateTxCmd(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegate",
		Short: "delegate power to a voter",
		RunE:  sendDelegateTx(cdc),
	}
	cmd.Flags().String(FlagUsername, "", "delegate user")
	cmd.Flags().String(FlagVoter, "", "voter to accept delegate")
	cmd.Flags().String(FlagAmount, "", "amount to delegate")
	return cmd
}

func sendDelegateTx(cdc *wire.Codec) client.CommandTxCallback {
	return func(cmd *cobra.Command, args []string) error {
		ctx := context.NewCoreContextFromViper()
		user := viper.GetString(FlagUsername)
		voter := viper.GetString(FlagVoter)
		// create the message
		msg := vote.NewDelegateMsg(user, voter, viper.GetString(FlagAmount))

		// build and sign the transaction, then broadcast to Tendermint
		res, signErr := ctx.SignBuildBroadcast(user, msg, cdc)

		if signErr != nil {
			return signErr
		}

		fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
		return nil
	}
}