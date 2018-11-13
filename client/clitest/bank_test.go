package clitest

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/stretchr/testify/require"
	"testing"
	sdk "github.com/cosmos/cosmos-sdk/types"
	irisInit "github.com/irisnet/irishub/init"
	"github.com/irisnet/irishub/app"
)

func init() {
	irisHome, iriscliHome = getTestingHomeDirs()
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(irisInit.Bech32PrefixAccAddr, irisInit.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(irisInit.Bech32PrefixValAddr, irisInit.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(irisInit.Bech32PrefixConsAddr, irisInit.Bech32PrefixConsPub)
	config.Seal()
}

func TestIrisCLIBankSend(t *testing.T) {
	chainID, servAddr, port := initializeFixtures(t)

	flags := fmt.Sprintf("--home=%s --node=%v --chain-id=%v", iriscliHome, servAddr, chainID)

	// start iris server
	proc := tests.GoExecuteTWithStdout(t, fmt.Sprintf("iris start --home=%s --rpc.laddr=%v", irisHome, servAddr))

	defer proc.Stop(false)
	tests.WaitForTMStart(port)
	tests.WaitForNextNBlocksTM(2, port)

	fooAddr, _ := executeGetAddrPK(t, fmt.Sprintf("iriscli keys show foo --output=json --home=%s", iriscliHome))
	barAddr, _ := executeGetAddrPK(t, fmt.Sprintf("iriscli keys show bar --output=json --home=%s", iriscliHome))

	fooAcc := executeGetAccount(t, fmt.Sprintf("iriscli bank account %s %v", fooAddr, flags))
	fooCoin := convertToIrisBaseAccount(t, fooAcc)
	require.Equal(t, "50iris", fooCoin)

	executeWrite(t, fmt.Sprintf("iriscli bank send %v --amount=10iris --to=%s --from=foo --gas=10000 --fee=0.04iris", flags, barAddr), app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(2, port)

	barAcc := executeGetAccount(t, fmt.Sprintf("iriscli bank account %s %v", barAddr, flags))
	barCoin := convertToIrisBaseAccount(t, barAcc)
	require.Equal(t, "10iris", barCoin)

	fooAcc = executeGetAccount(t, fmt.Sprintf("iriscli bank account %s %v", fooAddr, flags))
	fooCoin = convertToIrisBaseAccount(t, fooAcc)
	num := getAmountFromCoinStr(fooCoin)

	if !(num > 39 && num < 40) {
		t.Error("Test Failed: (39, 40) expected, recieved: {}", num)
	}

	// test autosequencing
	executeWrite(t, fmt.Sprintf("iriscli bank send %v --amount=10iris --to=%s --from=foo --gas=10000 --fee=0.04iris", flags, barAddr), app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(2, port)

	barAcc = executeGetAccount(t, fmt.Sprintf("iriscli bank account %s %v", barAddr, flags))
	barCoin = convertToIrisBaseAccount(t, barAcc)
	require.Equal(t, "20iris", barCoin)

	fooAcc = executeGetAccount(t, fmt.Sprintf("iriscli bank account %s %v", fooAddr, flags))
	fooCoin = convertToIrisBaseAccount(t, fooAcc)
	num = getAmountFromCoinStr(fooCoin)

	if !(num > 29 && num < 30) {
		t.Error("Test Failed: (29, 30) expected, recieved: {}", num)
	}

	// test memo
	executeWrite(t, fmt.Sprintf("iriscli bank send %v --amount=10iris --to=%s --from=foo --memo 'testmemo' --gas=10000 --fee=0.04iris", flags, barAddr), app.DefaultKeyPass)
	tests.WaitForNextNBlocksTM(2, port)

	barAcc = executeGetAccount(t, fmt.Sprintf("iriscli bank account %s %v", barAddr, flags))
	barCoin = convertToIrisBaseAccount(t, barAcc)
	require.Equal(t, "30iris", barCoin)

	fooAcc = executeGetAccount(t, fmt.Sprintf("iriscli bank account %s %v", fooAddr, flags))
	fooCoin = convertToIrisBaseAccount(t, fooAcc)
	num = getAmountFromCoinStr(fooCoin)

	if !(num > 19 && num < 20) {
		t.Error("Test Failed: (69, 70) expected, recieved: {}", num)
	}
}
