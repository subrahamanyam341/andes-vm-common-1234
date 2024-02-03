package builtInFunctions

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/subrahamanyam341/andes-core-16/core"
	"github.com/subrahamanyam341/andes-core-16/core/check"
	vmcommon "github.com/subrahamanyam341/andes-vm-common-1234"
)

type dctLocalMint struct {
	baseAlwaysActiveHandler
	keyPrefix             []byte
	marshaller            vmcommon.Marshalizer
	globalSettingsHandler vmcommon.DCTGlobalSettingsHandler
	rolesHandler          vmcommon.DCTRoleHandler
	enableEpochsHandler   vmcommon.EnableEpochsHandler
	funcGasCost           uint64
	mutExecution          sync.RWMutex
}

// NewDCTLocalMintFunc returns the dct local mint built-in function component
func NewDCTLocalMintFunc(
	funcGasCost uint64,
	marshaller vmcommon.Marshalizer,
	globalSettingsHandler vmcommon.DCTGlobalSettingsHandler,
	rolesHandler vmcommon.DCTRoleHandler,
	enableEpochsHandler vmcommon.EnableEpochsHandler,
) (*dctLocalMint, error) {
	if check.IfNil(marshaller) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(globalSettingsHandler) {
		return nil, ErrNilGlobalSettingsHandler
	}
	if check.IfNil(rolesHandler) {
		return nil, ErrNilRolesHandler
	}
	if check.IfNil(enableEpochsHandler) {
		return nil, ErrNilEnableEpochsHandler
	}

	e := &dctLocalMint{
		keyPrefix:             []byte(baseDCTKeyPrefix),
		marshaller:            marshaller,
		globalSettingsHandler: globalSettingsHandler,
		rolesHandler:          rolesHandler,
		funcGasCost:           funcGasCost,
		enableEpochsHandler:   enableEpochsHandler,
		mutExecution:          sync.RWMutex{},
	}

	return e, nil
}

// SetNewGasConfig is called whenever gas cost is changed
func (e *dctLocalMint) SetNewGasConfig(gasCost *vmcommon.GasCost) {
	if gasCost == nil {
		return
	}

	e.mutExecution.Lock()
	e.funcGasCost = gasCost.BuiltInCost.DCTLocalMint
	e.mutExecution.Unlock()
}

// ProcessBuiltinFunction resolves DCT local mint function call
func (e *dctLocalMint) ProcessBuiltinFunction(
	acntSnd, _ vmcommon.UserAccountHandler,
	vmInput *vmcommon.ContractCallInput,
) (*vmcommon.VMOutput, error) {
	e.mutExecution.RLock()
	defer e.mutExecution.RUnlock()

	err := checkInputArgumentsForLocalAction(acntSnd, vmInput, e.funcGasCost)
	if err != nil {
		return nil, err
	}

	tokenID := vmInput.Arguments[0]
	err = e.rolesHandler.CheckAllowedToExecute(acntSnd, tokenID, []byte(core.DCTRoleLocalMint))
	if err != nil {
		return nil, err
	}

	if len(vmInput.Arguments[1]) > core.MaxLenForDCTIssueMint {
		if e.enableEpochsHandler.IsConsistentTokensValuesLengthCheckEnabled() {
			return nil, fmt.Errorf("%w: max length for dct local mint value is %d", ErrInvalidArguments, core.MaxLenForDCTIssueMint)
		}
		// backward compatibility - return old error
		return nil, fmt.Errorf("%w max length for dct issue is %d", ErrInvalidArguments, core.MaxLenForDCTIssueMint)
	}

	value := big.NewInt(0).SetBytes(vmInput.Arguments[1])
	dctTokenKey := append(e.keyPrefix, tokenID...)
	err = addToDCTBalance(acntSnd, dctTokenKey, big.NewInt(0).Set(value), e.marshaller, e.globalSettingsHandler, vmInput.ReturnCallAfterError)
	if err != nil {
		return nil, err
	}

	vmOutput := &vmcommon.VMOutput{ReturnCode: vmcommon.Ok, GasRemaining: vmInput.GasProvided - e.funcGasCost}

	addDCTEntryInVMOutput(vmOutput, []byte(core.BuiltInFunctionDCTLocalMint), vmInput.Arguments[0], 0, value, vmInput.CallerAddr)

	return vmOutput, nil
}

// IsInterfaceNil returns true if underlying object in nil
func (e *dctLocalMint) IsInterfaceNil() bool {
	return e == nil
}
