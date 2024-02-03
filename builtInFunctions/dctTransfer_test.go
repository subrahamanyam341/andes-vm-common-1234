package builtInFunctions

import (
	"bytes"
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/subrahamanyam341/andes-core-16/core"
	"github.com/subrahamanyam341/andes-core-16/core/check"
	"github.com/subrahamanyam341/andes-core-16/data/dct"
	"github.com/subrahamanyam341/andes-core-16/data/vm"
	vmcommon "github.com/subrahamanyam341/andes-vm-common-1234"
	"github.com/subrahamanyam341/andes-vm-common-1234/mock"
)

func TestNewDCTTransferFunc(t *testing.T) {
	t.Parallel()

	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		transferFunc, err := NewDCTTransferFunc(10, nil, nil, nil, nil, nil)
		assert.Equal(t, ErrNilMarshalizer, err)
		assert.True(t, check.IfNil(transferFunc))
	})
	t.Run("nil global settings handler should error", func(t *testing.T) {
		t.Parallel()

		transferFunc, err := NewDCTTransferFunc(10, &mock.MarshalizerMock{}, nil, nil, nil, nil)
		assert.Equal(t, ErrNilGlobalSettingsHandler, err)
		assert.True(t, check.IfNil(transferFunc))
	})
	t.Run("nil shard coordinator should error", func(t *testing.T) {
		t.Parallel()

		transferFunc, err := NewDCTTransferFunc(10, &mock.MarshalizerMock{}, &mock.GlobalSettingsHandlerStub{}, nil, nil, nil)
		assert.Equal(t, ErrNilShardCoordinator, err)
		assert.True(t, check.IfNil(transferFunc))
	})
	t.Run("nil roles handler should error", func(t *testing.T) {
		t.Parallel()

		transferFunc, err := NewDCTTransferFunc(10, &mock.MarshalizerMock{}, &mock.GlobalSettingsHandlerStub{}, &mock.ShardCoordinatorStub{}, nil, nil)
		assert.Equal(t, ErrNilRolesHandler, err)
		assert.True(t, check.IfNil(transferFunc))
	})
	t.Run("nil enable epochs handler should error", func(t *testing.T) {
		t.Parallel()

		transferFunc, err := NewDCTTransferFunc(10, &mock.MarshalizerMock{}, &mock.GlobalSettingsHandlerStub{}, &mock.ShardCoordinatorStub{}, &mock.DCTRoleHandlerStub{}, nil)
		assert.Equal(t, ErrNilEnableEpochsHandler, err)
		assert.True(t, check.IfNil(transferFunc))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		transferFunc, err := NewDCTTransferFunc(10, &mock.MarshalizerMock{}, &mock.GlobalSettingsHandlerStub{}, &mock.ShardCoordinatorStub{}, &mock.DCTRoleHandlerStub{}, &mock.EnableEpochsHandlerStub{})
		assert.Nil(t, err)
		assert.False(t, check.IfNil(transferFunc))
	})
}
func TestDCTTransfer_ProcessBuiltInFunctionErrors(t *testing.T) {
	t.Parallel()

	shardC := &mock.ShardCoordinatorStub{}
	transferFunc, _ := NewDCTTransferFunc(10, &mock.MarshalizerMock{}, &mock.GlobalSettingsHandlerStub{}, shardC, &mock.DCTRoleHandlerStub{}, &mock.EnableEpochsHandlerStub{
		IsTransferToMetaFlagEnabledField:                     false,
		IsCheckCorrectTokenIDForTransferRoleFlagEnabledField: true,
	})
	_ = transferFunc.SetPayableChecker(&mock.PayableHandlerStub{})
	_, err := transferFunc.ProcessBuiltinFunction(nil, nil, nil)
	assert.Equal(t, err, ErrNilVmInput)

	input := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue: big.NewInt(0),
		},
	}
	_, err = transferFunc.ProcessBuiltinFunction(nil, nil, input)
	assert.Equal(t, err, ErrInvalidArguments)

	input = &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			GasProvided: 50,
			CallValue:   big.NewInt(0),
		},
	}
	key := []byte("key")
	value := []byte("value")
	input.Arguments = [][]byte{key, value}
	_, err = transferFunc.ProcessBuiltinFunction(nil, nil, input)
	assert.Nil(t, err)

	input.GasProvided = transferFunc.funcGasCost - 1
	accSnd := mock.NewUserAccount([]byte("address"))
	_, err = transferFunc.ProcessBuiltinFunction(accSnd, nil, input)
	assert.Equal(t, err, ErrNotEnoughGas)

	input.GasProvided = transferFunc.funcGasCost
	input.RecipientAddr = core.DCTSCAddress
	shardC.ComputeIdCalled = func(address []byte) uint32 {
		return core.MetachainShardId
	}
	_, err = transferFunc.ProcessBuiltinFunction(accSnd, nil, input)
	assert.Equal(t, err, ErrInvalidRcvAddr)
}

func TestDCTTransfer_ProcessBuiltInFunctionSingleShard(t *testing.T) {
	t.Parallel()

	marshaller := &mock.MarshalizerMock{}
	dctRoleHandler := &mock.DCTRoleHandlerStub{
		CheckAllowedToExecuteCalled: func(account vmcommon.UserAccountHandler, tokenID []byte, action []byte) error {
			assert.Equal(t, core.DCTRoleTransfer, string(action))
			return nil
		},
	}
	enableEpochsHandler := &mock.EnableEpochsHandlerStub{
		IsTransferToMetaFlagEnabledField:                     false,
		IsCheckCorrectTokenIDForTransferRoleFlagEnabledField: true,
	}
	transferFunc, _ := NewDCTTransferFunc(10, marshaller, &mock.GlobalSettingsHandlerStub{}, &mock.ShardCoordinatorStub{}, dctRoleHandler, enableEpochsHandler)
	_ = transferFunc.SetPayableChecker(&mock.PayableHandlerStub{})

	input := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			GasProvided: 50,
			CallValue:   big.NewInt(0),
		},
	}
	key := []byte("key")
	value := big.NewInt(10).Bytes()
	input.Arguments = [][]byte{key, value}
	accSnd := mock.NewUserAccount([]byte("snd"))
	accDst := mock.NewUserAccount([]byte("dst"))

	_, err := transferFunc.ProcessBuiltinFunction(accSnd, accDst, input)
	assert.Equal(t, err, ErrInsufficientFunds)

	dctKey := append(transferFunc.keyPrefix, key...)
	dctToken := &dct.DCToken{Value: big.NewInt(100)}
	marshaledData, _ := marshaller.Marshal(dctToken)
	_ = accSnd.AccountDataHandler().SaveKeyValue(dctKey, marshaledData)

	_, err = transferFunc.ProcessBuiltinFunction(accSnd, accDst, input)
	assert.Nil(t, err)
	marshaledData, _, _ = accSnd.AccountDataHandler().RetrieveValue(dctKey)
	_ = marshaller.Unmarshal(dctToken, marshaledData)
	assert.True(t, dctToken.Value.Cmp(big.NewInt(90)) == 0)

	marshaledData, _, _ = accDst.AccountDataHandler().RetrieveValue(dctKey)
	_ = marshaller.Unmarshal(dctToken, marshaledData)
	assert.True(t, dctToken.Value.Cmp(big.NewInt(10)) == 0)
}

func TestDCTTransfer_ProcessBuiltInFunctionSenderInShard(t *testing.T) {
	t.Parallel()

	marshaller := &mock.MarshalizerMock{}
	transferFunc, _ := NewDCTTransferFunc(10, marshaller, &mock.GlobalSettingsHandlerStub{}, &mock.ShardCoordinatorStub{}, &mock.DCTRoleHandlerStub{}, &mock.EnableEpochsHandlerStub{
		IsTransferToMetaFlagEnabledField:                     false,
		IsCheckCorrectTokenIDForTransferRoleFlagEnabledField: true,
	})
	_ = transferFunc.SetPayableChecker(&mock.PayableHandlerStub{})

	input := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			GasProvided: 50,
			CallValue:   big.NewInt(0),
		},
	}
	key := []byte("key")
	value := big.NewInt(10).Bytes()
	input.Arguments = [][]byte{key, value}
	accSnd := mock.NewUserAccount([]byte("snd"))

	dctKey := append(transferFunc.keyPrefix, key...)
	dctToken := &dct.DCToken{Value: big.NewInt(100)}
	marshaledData, _ := marshaller.Marshal(dctToken)
	_ = accSnd.AccountDataHandler().SaveKeyValue(dctKey, marshaledData)

	_, err := transferFunc.ProcessBuiltinFunction(accSnd, nil, input)
	assert.Nil(t, err)
	marshaledData, _, _ = accSnd.AccountDataHandler().RetrieveValue(dctKey)
	_ = marshaller.Unmarshal(dctToken, marshaledData)
	assert.True(t, dctToken.Value.Cmp(big.NewInt(90)) == 0)
}

func TestDCTTransfer_ProcessBuiltInFunctionDestInShard(t *testing.T) {
	t.Parallel()

	marshaller := &mock.MarshalizerMock{}
	transferFunc, _ := NewDCTTransferFunc(10, marshaller, &mock.GlobalSettingsHandlerStub{}, &mock.ShardCoordinatorStub{}, &mock.DCTRoleHandlerStub{}, &mock.EnableEpochsHandlerStub{
		IsTransferToMetaFlagEnabledField:                     false,
		IsCheckCorrectTokenIDForTransferRoleFlagEnabledField: true,
	})
	_ = transferFunc.SetPayableChecker(&mock.PayableHandlerStub{})

	input := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			GasProvided: 50,
			CallValue:   big.NewInt(0),
		},
	}
	key := []byte("key")
	value := big.NewInt(10).Bytes()
	input.Arguments = [][]byte{key, value}
	accDst := mock.NewUserAccount([]byte("dst"))

	vmOutput, err := transferFunc.ProcessBuiltinFunction(nil, accDst, input)
	assert.Nil(t, err)
	dctKey := append(transferFunc.keyPrefix, key...)
	dctToken := &dct.DCToken{}
	marshaledData, _, _ := accDst.AccountDataHandler().RetrieveValue(dctKey)
	_ = marshaller.Unmarshal(dctToken, marshaledData)
	assert.True(t, dctToken.Value.Cmp(big.NewInt(10)) == 0)
	assert.Equal(t, uint64(0), vmOutput.GasRemaining)
}

func TestDCTTransfer_ProcessBuiltInFunctionTooLongValue(t *testing.T) {
	t.Parallel()

	marshaller := &mock.MarshalizerMock{}
	transferFunc, _ := NewDCTTransferFunc(10, marshaller, &mock.GlobalSettingsHandlerStub{}, &mock.ShardCoordinatorStub{}, &mock.DCTRoleHandlerStub{}, &mock.EnableEpochsHandlerStub{
		IsTransferToMetaFlagEnabledField:                     false,
		IsCheckCorrectTokenIDForTransferRoleFlagEnabledField: true,
	})
	_ = transferFunc.SetPayableChecker(&mock.PayableHandlerStub{})

	bigValueStr := "1" + strings.Repeat("0", 1000)
	bigValue, _ := big.NewInt(0).SetString(bigValueStr, 10)
	input := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			GasProvided: 50,
			CallValue:   big.NewInt(0),
			Arguments:   [][]byte{[]byte("tkn"), bigValue.Bytes()},
		},
	}
	accDst := mock.NewUserAccount([]byte("dst"))

	// before the activation of the flag, large values should not return error
	vmOutput, err := transferFunc.ProcessBuiltinFunction(nil, accDst, input)
	assert.Nil(t, err)
	assert.NotEmpty(t, vmOutput)

	// after the activation, it should return an error
	transferFunc.enableEpochsHandler = &mock.EnableEpochsHandlerStub{
		IsConsistentTokensValuesLengthCheckEnabledField: true,
	}
	vmOutput, err = transferFunc.ProcessBuiltinFunction(nil, accDst, input)
	assert.Equal(t, "invalid arguments to process built-in function: max length for dct transfer value is 100", err.Error())
	assert.Empty(t, vmOutput)
}

func TestDCTTransfer_SndDstFrozen(t *testing.T) {
	t.Parallel()

	marshaller := &mock.MarshalizerMock{}
	accountStub := &mock.AccountsStub{}
	dctGlobalSettingsFunc, _ := NewDCTGlobalSettingsFunc(accountStub, marshaller, true, core.BuiltInFunctionDCTPause, trueHandler)
	transferFunc, _ := NewDCTTransferFunc(10, marshaller, dctGlobalSettingsFunc, &mock.ShardCoordinatorStub{}, &mock.DCTRoleHandlerStub{}, &mock.EnableEpochsHandlerStub{
		IsTransferToMetaFlagEnabledField:                     false,
		IsCheckCorrectTokenIDForTransferRoleFlagEnabledField: true,
	})
	_ = transferFunc.SetPayableChecker(&mock.PayableHandlerStub{})

	input := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			GasProvided: 50,
			CallValue:   big.NewInt(0),
		},
	}
	key := []byte("key")
	value := big.NewInt(10).Bytes()
	input.Arguments = [][]byte{key, value}
	accSnd := mock.NewUserAccount([]byte("snd"))
	accDst := mock.NewUserAccount([]byte("dst"))

	dctFrozen := DCTUserMetadata{Frozen: true}
	dctNotFrozen := DCTUserMetadata{Frozen: false}

	dctKey := append(transferFunc.keyPrefix, key...)
	dctToken := &dct.DCToken{Value: big.NewInt(100), Properties: dctFrozen.ToBytes()}
	marshaledData, _ := marshaller.Marshal(dctToken)
	_ = accSnd.AccountDataHandler().SaveKeyValue(dctKey, marshaledData)

	_, err := transferFunc.ProcessBuiltinFunction(accSnd, accDst, input)
	assert.Equal(t, err, ErrDCTIsFrozenForAccount)

	dctToken = &dct.DCToken{Value: big.NewInt(100), Properties: dctNotFrozen.ToBytes()}
	marshaledData, _ = marshaller.Marshal(dctToken)
	_ = accSnd.AccountDataHandler().SaveKeyValue(dctKey, marshaledData)

	dctToken = &dct.DCToken{Value: big.NewInt(100), Properties: dctFrozen.ToBytes()}
	marshaledData, _ = marshaller.Marshal(dctToken)
	_ = accDst.AccountDataHandler().SaveKeyValue(dctKey, marshaledData)

	_, err = transferFunc.ProcessBuiltinFunction(accSnd, accDst, input)
	assert.Equal(t, err, ErrDCTIsFrozenForAccount)

	marshaledData, _, _ = accDst.AccountDataHandler().RetrieveValue(dctKey)
	_ = marshaller.Unmarshal(dctToken, marshaledData)
	assert.True(t, dctToken.Value.Cmp(big.NewInt(100)) == 0)

	input.ReturnCallAfterError = true
	_, err = transferFunc.ProcessBuiltinFunction(accSnd, accDst, input)
	assert.Nil(t, err)

	dctToken = &dct.DCToken{Value: big.NewInt(100), Properties: dctNotFrozen.ToBytes()}
	marshaledData, _ = marshaller.Marshal(dctToken)
	_ = accDst.AccountDataHandler().SaveKeyValue(dctKey, marshaledData)

	systemAccount := mock.NewUserAccount(vmcommon.SystemAccountAddress)
	dctGlobal := DCTGlobalMetadata{Paused: true}
	pauseKey := []byte(baseDCTKeyPrefix + string(key))
	_ = systemAccount.AccountDataHandler().SaveKeyValue(pauseKey, dctGlobal.ToBytes())

	accountStub.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, vmcommon.SystemAccountAddress) {
			return systemAccount, nil
		}
		return accDst, nil
	}

	input.ReturnCallAfterError = false
	_, err = transferFunc.ProcessBuiltinFunction(accSnd, accDst, input)
	assert.Equal(t, err, ErrDCTTokenIsPaused)

	input.ReturnCallAfterError = true
	_, err = transferFunc.ProcessBuiltinFunction(accSnd, accDst, input)
	assert.Nil(t, err)
}

func TestDCTTransfer_SndDstWithLimitedTransfer(t *testing.T) {
	t.Parallel()

	marshaller := &mock.MarshalizerMock{}
	accountStub := &mock.AccountsStub{}
	rolesHandler := &mock.DCTRoleHandlerStub{
		CheckAllowedToExecuteCalled: func(account vmcommon.UserAccountHandler, tokenID []byte, action []byte) error {
			if bytes.Equal(action, []byte(core.DCTRoleTransfer)) {
				return ErrActionNotAllowed
			}
			return nil
		},
	}
	dctGlobalSettingsFunc, _ := NewDCTGlobalSettingsFunc(accountStub, marshaller, true, core.BuiltInFunctionDCTSetLimitedTransfer, trueHandler)
	transferFunc, _ := NewDCTTransferFunc(10, marshaller, dctGlobalSettingsFunc, &mock.ShardCoordinatorStub{}, rolesHandler, &mock.EnableEpochsHandlerStub{
		IsTransferToMetaFlagEnabledField:                     false,
		IsCheckCorrectTokenIDForTransferRoleFlagEnabledField: true,
	})
	_ = transferFunc.SetPayableChecker(&mock.PayableHandlerStub{})

	input := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			GasProvided: 50,
			CallValue:   big.NewInt(0),
		},
	}
	key := []byte("key")
	value := big.NewInt(10).Bytes()
	input.Arguments = [][]byte{key, value}
	accSnd := mock.NewUserAccount([]byte("snd"))
	accDst := mock.NewUserAccount([]byte("dst"))

	dctKey := append(transferFunc.keyPrefix, key...)
	dctToken := &dct.DCToken{Value: big.NewInt(100)}
	marshaledData, _ := marshaller.Marshal(dctToken)
	_ = accSnd.AccountDataHandler().SaveKeyValue(dctKey, marshaledData)

	dctToken = &dct.DCToken{Value: big.NewInt(100)}
	marshaledData, _ = marshaller.Marshal(dctToken)
	_ = accDst.AccountDataHandler().SaveKeyValue(dctKey, marshaledData)

	systemAccount := mock.NewUserAccount(vmcommon.SystemAccountAddress)
	dctGlobal := DCTGlobalMetadata{LimitedTransfer: true}
	pauseKey := []byte(baseDCTKeyPrefix + string(key))
	_ = systemAccount.AccountDataHandler().SaveKeyValue(pauseKey, dctGlobal.ToBytes())

	accountStub.LoadAccountCalled = func(address []byte) (vmcommon.AccountHandler, error) {
		if bytes.Equal(address, vmcommon.SystemAccountAddress) {
			return systemAccount, nil
		}
		return accDst, nil
	}

	_, err := transferFunc.ProcessBuiltinFunction(nil, accDst, input)
	assert.Nil(t, err)

	_, err = transferFunc.ProcessBuiltinFunction(accSnd, accDst, input)
	assert.Equal(t, err, ErrActionNotAllowed)

	_, err = transferFunc.ProcessBuiltinFunction(accSnd, nil, input)
	assert.Equal(t, err, ErrActionNotAllowed)

	input.ReturnCallAfterError = true
	_, err = transferFunc.ProcessBuiltinFunction(accSnd, accDst, input)
	assert.Nil(t, err)

	input.ReturnCallAfterError = false
	rolesHandler.CheckAllowedToExecuteCalled = func(account vmcommon.UserAccountHandler, tokenID []byte, action []byte) error {
		if bytes.Equal(account.AddressBytes(), accSnd.Address) && bytes.Equal(tokenID, key) {
			return nil
		}
		return ErrActionNotAllowed
	}

	_, err = transferFunc.ProcessBuiltinFunction(accSnd, accDst, input)
	assert.Nil(t, err)

	rolesHandler.CheckAllowedToExecuteCalled = func(account vmcommon.UserAccountHandler, tokenID []byte, action []byte) error {
		if bytes.Equal(account.AddressBytes(), accDst.Address) && bytes.Equal(tokenID, key) {
			return nil
		}
		return ErrActionNotAllowed
	}

	_, err = transferFunc.ProcessBuiltinFunction(accSnd, accDst, input)
	assert.Nil(t, err)
}

func TestDCTTransfer_ProcessBuiltInFunctionOnAsyncCallBack(t *testing.T) {
	t.Parallel()

	marshaller := &mock.MarshalizerMock{}
	transferFunc, _ := NewDCTTransferFunc(10, marshaller, &mock.GlobalSettingsHandlerStub{}, &mock.ShardCoordinatorStub{}, &mock.DCTRoleHandlerStub{}, &mock.EnableEpochsHandlerStub{
		IsTransferToMetaFlagEnabledField:                     false,
		IsCheckCorrectTokenIDForTransferRoleFlagEnabledField: true,
	})
	_ = transferFunc.SetPayableChecker(&mock.PayableHandlerStub{})

	input := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			GasProvided: 50,
			CallValue:   big.NewInt(0),
			CallType:    vm.AsynchronousCallBack,
		},
	}
	key := []byte("key")
	value := big.NewInt(10).Bytes()
	input.Arguments = [][]byte{key, value}
	accSnd := mock.NewUserAccount([]byte("snd"))
	accDst := mock.NewUserAccount(core.DCTSCAddress)

	dctKey := append(transferFunc.keyPrefix, key...)
	dctToken := &dct.DCToken{Value: big.NewInt(100)}
	marshaledData, _ := marshaller.Marshal(dctToken)
	_ = accSnd.AccountDataHandler().SaveKeyValue(dctKey, marshaledData)

	vmOutput, err := transferFunc.ProcessBuiltinFunction(nil, accDst, input)
	assert.Nil(t, err)

	marshaledData, _, _ = accDst.AccountDataHandler().RetrieveValue(dctKey)
	_ = marshaller.Unmarshal(dctToken, marshaledData)
	assert.True(t, dctToken.Value.Cmp(big.NewInt(10)) == 0)

	assert.Equal(t, vmOutput.GasRemaining, input.GasProvided)

	vmOutput, err = transferFunc.ProcessBuiltinFunction(accSnd, accDst, input)
	assert.Nil(t, err)
	vmOutput.GasRemaining = input.GasProvided - transferFunc.funcGasCost

	marshaledData, _, _ = accSnd.AccountDataHandler().RetrieveValue(dctKey)
	_ = marshaller.Unmarshal(dctToken, marshaledData)
	assert.True(t, dctToken.Value.Cmp(big.NewInt(90)) == 0)
}
