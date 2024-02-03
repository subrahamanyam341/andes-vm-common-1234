package mock

import vmcommon "github.com/subrahamanyam341/andes-vm-common-1234"

// DCTRoleHandlerStub -
type DCTRoleHandlerStub struct {
	CheckAllowedToExecuteCalled func(account vmcommon.UserAccountHandler, tokenID []byte, action []byte) error
}

// CheckAllowedToExecute -
func (e *DCTRoleHandlerStub) CheckAllowedToExecute(account vmcommon.UserAccountHandler, tokenID []byte, action []byte) error {
	if e.CheckAllowedToExecuteCalled != nil {
		return e.CheckAllowedToExecuteCalled(account, tokenID, action)
	}

	return nil
}

// IsInterfaceNil -
func (e *DCTRoleHandlerStub) IsInterfaceNil() bool {
	return e == nil
}
