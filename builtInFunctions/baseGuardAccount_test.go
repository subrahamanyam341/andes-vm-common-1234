package builtInFunctions

import (
	"testing"

	"github.com/stretchr/testify/require"
	vmcommon "github.com/subrahamanyam341/andes-vm-common-1234"
)

func createGuardAccountArgs() GuardAccountArgs {
	return GuardAccountArgs{
		BaseAccountGuarderArgs: createBaseAccountGuarderArgs(),
	}
}

func TestBaseGuardAccount_SetNewGasConfig(t *testing.T) {
	t.Parallel()

	args := createGuardAccountArgs()
	baseGuardAccount, _ := newBaseGuardAccount(args)
	require.Equal(t, args.FuncGasCost, baseGuardAccount.funcGasCost)

	newGuardAccountCost := args.FuncGasCost + 1
	newGasCost := &vmcommon.GasCost{BuiltInCost: vmcommon.BuiltInCost{GuardAccount: newGuardAccountCost}}

	baseGuardAccount.SetNewGasConfig(newGasCost)
	require.Equal(t, newGuardAccountCost, baseGuardAccount.funcGasCost)
}
