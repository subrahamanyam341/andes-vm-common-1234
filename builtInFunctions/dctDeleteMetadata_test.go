package builtInFunctions

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/subrahamanyam341/andes-core-16/core"
	"github.com/subrahamanyam341/andes-core-16/data/dct"
	vmcommon "github.com/subrahamanyam341/andes-vm-common-1234"
	"github.com/subrahamanyam341/andes-vm-common-1234/mock"
)

func createMockArgsForNewDCTDelete() ArgsNewDCTDeleteMetadata {
	return ArgsNewDCTDeleteMetadata{
		FuncGasCost:    1,
		Marshalizer:    &mock.MarshalizerMock{},
		Accounts:       &mock.AccountsStub{},
		AllowedAddress: bytes.Repeat([]byte{1}, 32),
		Delete:         true,
		EnableEpochsHandler: &mock.EnableEpochsHandlerStub{
			IsSendAlwaysFlagEnabledField: true,
		},
	}
}

func TestNewDCTDeleteMetadataFunc(t *testing.T) {
	t.Parallel()

	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsForNewDCTDelete()
		args.Marshalizer = nil
		_, err := NewDCTDeleteMetadataFunc(args)
		assert.Equal(t, err, ErrNilMarshalizer)
	})
	t.Run("nil accounts adapter should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsForNewDCTDelete()
		args.Accounts = nil
		_, err := NewDCTDeleteMetadataFunc(args)
		assert.Equal(t, err, ErrNilAccountsAdapter)
	})
	t.Run("nil enable epochs handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsForNewDCTDelete()
		args.EnableEpochsHandler = nil
		_, err := NewDCTDeleteMetadataFunc(args)
		assert.Equal(t, err, ErrNilEnableEpochsHandler)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsForNewDCTDelete()
		e, err := NewDCTDeleteMetadataFunc(args)
		assert.Nil(t, err)
		assert.False(t, e.IsInterfaceNil())
		assert.True(t, e.IsActive())
	})
}

func TestDctDeleteMetaData_ProcessBuiltinFunctionErrors(t *testing.T) {
	t.Parallel()

	args := createMockArgsForNewDCTDelete()
	e, _ := NewDCTDeleteMetadataFunc(args)

	vmOutput, err := e.ProcessBuiltinFunction(nil, nil, nil)
	assert.Nil(t, vmOutput)
	assert.Equal(t, ErrNilVmInput, err)

	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, &vmcommon.ContractCallInput{VMInput: vmcommon.VMInput{CallValue: big.NewInt(10)}})
	assert.Nil(t, vmOutput)
	assert.Equal(t, ErrBuiltInFunctionCalledWithValue, err)

	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, &vmcommon.ContractCallInput{VMInput: vmcommon.VMInput{CallValue: big.NewInt(0)}})
	assert.Nil(t, vmOutput)
	assert.Equal(t, ErrAddressIsNotAllowed, err)

	vmInput := &vmcommon.ContractCallInput{VMInput: vmcommon.VMInput{CallValue: big.NewInt(0)}}
	vmInput.CallerAddr = e.allowedAddress
	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, vmInput)
	assert.Nil(t, vmOutput)
	assert.Equal(t, ErrInvalidRcvAddr, err)

	vmInput.RecipientAddr = e.allowedAddress
	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, vmInput)
	assert.Nil(t, vmOutput)
	assert.Equal(t, ErrInvalidNumOfArgs, err)

	e.delete = false
	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, vmInput)
	assert.Nil(t, vmOutput)
	assert.Equal(t, ErrInvalidNumOfArgs, err)

	vmInput.Arguments = [][]byte{{1}, {0}, {1}}
	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, vmInput)
	assert.Nil(t, vmOutput)
	assert.NotNil(t, err)

	e.delete = true
	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, vmInput)
	assert.Nil(t, vmOutput)
	assert.Equal(t, ErrInvalidNumOfArgs, err)

	vmInput.Arguments = [][]byte{{1}, {0}, {1}, {1}}
	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, vmInput)
	assert.Nil(t, vmOutput)
	assert.NotNil(t, err)

	e.delete = false

	acnt := mock.NewUserAccount(vmcommon.SystemAccountAddress)
	accounts := &mock.AccountsStub{LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
		return acnt, nil
	}}

	e.accounts = accounts
	vmInput.Arguments = [][]byte{{1}, {0}, {1}}
	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, vmInput)
	assert.Nil(t, vmOutput)
	assert.Equal(t, ErrInvalidNonce, err)

	vmInput.Arguments[0] = []byte("TOKEN-ABABAB")
	vmInput.Arguments[1] = []byte{1}
	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, vmInput)
	assert.Nil(t, vmOutput)
	assert.Equal(t, ErrInvalidTokenID, err)

	vmInput.Arguments[0] = []byte("TOKEN-ababab")
	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, vmInput)
	assert.Nil(t, vmOutput)
	assert.NotNil(t, err)

	dctMetadata := &dct.MetaData{Name: []byte("something"), Nonce: 1}
	marshalledData, _ := e.marshaller.Marshal(dctMetadata)
	vmInput.Arguments[2] = make([]byte, len(marshalledData))
	copy(vmInput.Arguments[2], marshalledData)

	dctTokenKey := append(e.keyPrefix, vmInput.Arguments[0]...)
	dctNftTokenKey := computeDCTNFTTokenKey(dctTokenKey, 1)
	err = acnt.SaveKeyValue(dctNftTokenKey, []byte("t"))
	assert.Nil(t, err)

	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, vmInput)
	assert.Nil(t, vmOutput)
	assert.NotNil(t, err)

	dctData := &dct.DCToken{Value: big.NewInt(0), TokenMetaData: &dct.MetaData{Name: []byte("data")}}
	marshalledData, _ = e.marshaller.Marshal(dctData)
	err = acnt.SaveKeyValue(dctNftTokenKey, marshalledData)
	assert.Nil(t, err)

	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, vmInput)
	assert.Nil(t, vmOutput)
	assert.Equal(t, ErrTokenHasValidMetadata, err)

	_ = acnt.SaveKeyValue(dctNftTokenKey, nil)
	testErr := errors.New("testError")
	accounts.SaveAccountCalled = func(account vmcommon.AccountHandler) error {
		return testErr
	}

	vmInput.Arguments[1] = []byte{2}
	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, vmInput)
	assert.Nil(t, vmOutput)
	assert.Equal(t, ErrInvalidMetadata, err)

	vmInput.Arguments[1] = []byte{1}
	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, vmInput)
	assert.Nil(t, vmOutput)
	assert.Equal(t, testErr, err)
}

func TestDctDeleteMetaData_ProcessBuiltinFunctionGetNodeFromDbErr(t *testing.T) {
	t.Parallel()

	args := createMockArgsForNewDCTDelete()
	args.Delete = false
	args.Accounts = &mock.AccountsStub{
		LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return &mock.AccountWrapMock{
				RetrieveValueCalled: func(_ []byte) ([]byte, uint32, error) {
					return nil, 0, core.NewGetNodeFromDBErrWithKey([]byte("key"), errors.New("error"), "")
				},
			}, nil
		},
	}
	dctMetadata := &dct.MetaData{Name: []byte("something"), Nonce: 1}
	marshalledData, _ := args.Marshalizer.Marshal(dctMetadata)
	e, _ := NewDCTDeleteMetadataFunc(args)
	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue:  big.NewInt(0),
			CallerAddr: e.allowedAddress,
			Arguments:  [][]byte{[]byte("TOKEN-ababab"), {1}, marshalledData},
		},
		RecipientAddr: e.allowedAddress,
	}

	output, err := e.ProcessBuiltinFunction(nil, nil, vmInput)
	assert.Nil(t, output)
	assert.True(t, core.IsGetNodeFromDBError(err))
}

func TestDctDeleteMetaData_ProcessBuiltinFunctionAdd(t *testing.T) {
	t.Parallel()

	args := createMockArgsForNewDCTDelete()
	args.Delete = false
	e, _ := NewDCTDeleteMetadataFunc(args)

	vmInput := &vmcommon.ContractCallInput{VMInput: vmcommon.VMInput{CallValue: big.NewInt(0)}}
	vmInput.CallerAddr = e.allowedAddress
	vmInput.RecipientAddr = e.allowedAddress
	vmInput.Arguments = [][]byte{{1}, {0}, {1}}
	vmInput.Arguments[0] = []byte("TOKEN-ababab")
	vmInput.Arguments[1] = []byte{1}
	dctMetadata := &dct.MetaData{Name: []byte("something"), Nonce: 1}
	marshalledData, _ := e.marshaller.Marshal(dctMetadata)
	vmInput.Arguments[2] = make([]byte, len(marshalledData))
	copy(vmInput.Arguments[2], marshalledData)

	acnt := mock.NewUserAccount(vmcommon.SystemAccountAddress)
	accounts := &mock.AccountsStub{
		LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return acnt, nil
		}}

	e.accounts = accounts

	vmOutput, err := e.ProcessBuiltinFunction(nil, nil, vmInput)
	assert.NotNil(t, vmOutput)
	assert.Nil(t, err)
}

func TestDctDeleteMetaData_ProcessBuiltinFunctionDelete(t *testing.T) {
	t.Parallel()

	args := createMockArgsForNewDCTDelete()
	e, _ := NewDCTDeleteMetadataFunc(args)

	vmInput := &vmcommon.ContractCallInput{VMInput: vmcommon.VMInput{CallValue: big.NewInt(0)}}
	vmInput.CallerAddr = e.allowedAddress
	vmInput.RecipientAddr = e.allowedAddress
	vmInput.Arguments = [][]byte{{1}, {2}, {1}, {1}}

	acnt := mock.NewUserAccount(vmcommon.SystemAccountAddress)
	accounts := &mock.AccountsStub{
		LoadAccountCalled: func(address []byte) (vmcommon.AccountHandler, error) {
			return acnt, nil
		}}

	e.accounts = accounts

	vmOutput, err := e.ProcessBuiltinFunction(nil, nil, vmInput)
	assert.Nil(t, vmOutput)
	assert.Equal(t, err, ErrInvalidTokenID)

	vmInput.Arguments[0] = []byte("TOKEN-ababab")
	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, vmInput)
	assert.Nil(t, vmOutput)
	assert.Equal(t, ErrInvalidNumOfArgs, err)

	vmInput.Arguments[2] = []byte{0}
	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, vmInput)
	assert.Nil(t, vmOutput)
	assert.Equal(t, err, ErrInvalidNonce)

	vmInput.Arguments[2] = []byte{10}
	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, vmInput)
	assert.Nil(t, vmOutput)
	assert.Equal(t, err, ErrInvalidArguments)

	vmInput.Arguments[1] = []byte{1}
	vmInput.Arguments[3] = []byte{11}

	vmInput.Arguments = append(vmInput.Arguments, []byte("TOKEN-ababab"), []byte{2})
	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, vmInput)
	assert.Nil(t, vmOutput)
	assert.Equal(t, ErrInvalidNumOfArgs, err)

	vmInput.Arguments = append(vmInput.Arguments, []byte{1}, []byte{2}, []byte{4}, []byte{10})
	vmOutput, err = e.ProcessBuiltinFunction(nil, nil, vmInput)
	assert.NotNil(t, vmOutput)
	assert.Nil(t, err)
}
