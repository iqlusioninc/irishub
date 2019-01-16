  // nolint
package cli

import (
	"fmt"
	"os"

	"github.com/irisnet/irishub/client/context"
	"github.com/irisnet/irishub/client/utils"
	"github.com/irisnet/irishub/codec"
	"github.com/irisnet/irishub/modules/distribution/types"
	sdk "github.com/irisnet/irishub/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	flagOnlyFromValidator = "only-from-validator"
	flagIsValidator       = "is-validator"
)

// command to withdraw rewards
func GetCmdWithdrawRewards(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "withdraw-rewards",
		Short:   "withdraw rewards for either: all-delegations, a delegation, or a validator",
		Example: "iriscli distribution withdraw-rewards --from <key name> --fee=0.004iris --chain-id=<chain-id>",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			onlyFromVal := viper.GetString(flagOnlyFromValidator)
			isVal := viper.GetBool(flagIsValidator)

			if onlyFromVal != "" && isVal {
				return fmt.Errorf("cannot use --%v, and --%v flags together",
					flagOnlyFromValidator, flagIsValidator)
			}

			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithLogger(os.Stdout).
				WithAccountDecoder(utils.GetAccountDecoder(cdc))
			txCtx := utils.NewTxContextFromCLI().WithCodec(cdc).WithCliCtx(cliCtx)

			var msg sdk.Msg
			switch {
			case isVal:
				addr, err := cliCtx.GetFromAddress()
				if err != nil {
					return err
				}
				valAddr := sdk.ValAddress(addr.Bytes())
				msg = types.NewMsgWithdrawValidatorRewardsAll(valAddr)
			case onlyFromVal != "":
				delAddr, err := cliCtx.GetFromAddress()
				if err != nil {
					return err
				}

				valAddr, err := sdk.ValAddressFromBech32(onlyFromVal)
				if err != nil {
					return err
				}

				msg = types.NewMsgWithdrawDelegatorReward(delAddr, valAddr)
			default:
				delAddr, err := cliCtx.GetFromAddress()
				if err != nil {
					return err
				}
				msg = types.NewMsgWithdrawDelegatorRewardsAll(delAddr)
			}

			// build and sign the transaction, then broadcast to Tendermint
			return utils.SendOrPrintTx(txCtx, cliCtx, []sdk.Msg{msg})
		},
	}
	cmd.Flags().String(flagOnlyFromValidator, "", "only withdraw from this validator address (in bech)")
	cmd.Flags().Bool(flagIsValidator, false, "also withdraw validator's commission")
	return cmd
}
