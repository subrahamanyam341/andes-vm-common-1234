package datafield

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"unicode"

	"github.com/subrahamanyam341/andes-core-16/core"
)

const (
	dctIdentifierSeparator  = "-"
	dctRandomSequenceLength = 6
)

// TODO refactor this part to use the built-in container for the list of all the built-in functions
func getAllBuiltInFunctions() []string {
	return []string{
		core.BuiltInFunctionClaimDeveloperRewards,
		core.BuiltInFunctionChangeOwnerAddress,
		core.BuiltInFunctionSetUserName,
		core.BuiltInFunctionSaveKeyValue,
		core.BuiltInFunctionDCTTransfer,
		core.BuiltInFunctionDCTBurn,
		core.BuiltInFunctionDCTFreeze,
		core.BuiltInFunctionDCTUnFreeze,
		core.BuiltInFunctionDCTWipe,
		core.BuiltInFunctionDCTPause,
		core.BuiltInFunctionDCTUnPause,
		core.BuiltInFunctionSetDCTRole,
		core.BuiltInFunctionUnSetDCTRole,
		core.BuiltInFunctionDCTSetLimitedTransfer,
		core.BuiltInFunctionDCTUnSetLimitedTransfer,
		core.BuiltInFunctionDCTLocalMint,
		core.BuiltInFunctionDCTLocalBurn,
		core.BuiltInFunctionDCTNFTTransfer,
		core.BuiltInFunctionDCTNFTCreate,
		core.BuiltInFunctionDCTNFTAddQuantity,
		core.BuiltInFunctionDCTNFTCreateRoleTransfer,
		core.BuiltInFunctionDCTNFTBurn,
		core.BuiltInFunctionDCTNFTAddURI,
		core.BuiltInFunctionDCTNFTUpdateAttributes,
		core.BuiltInFunctionMultiDCTNFTTransfer,
		core.BuiltInFunctionMigrateDataTrie,
		core.DCTRoleLocalMint,
		core.DCTRoleLocalBurn,
		core.DCTRoleNFTCreate,
		core.DCTRoleNFTCreateMultiShard,
		core.DCTRoleNFTAddQuantity,
		core.DCTRoleNFTBurn,
		core.DCTRoleNFTAddURI,
		core.DCTRoleNFTUpdateAttributes,
		core.DCTRoleTransfer,
		core.BuiltInFunctionSetGuardian,
		core.BuiltInFunctionUnGuardAccount,
		core.BuiltInFunctionGuardAccount,
	}
}

func isBuiltInFunction(builtInFunctionsList []string, function string) bool {
	for _, builtInFunction := range builtInFunctionsList {
		if builtInFunction == function {
			return true
		}
	}

	return false
}

func computeTokenIdentifier(token string, nonce uint64) string {
	if token == "" || nonce == 0 {
		return ""
	}

	nonceBig := big.NewInt(0).SetUint64(nonce)
	hexEncodedNonce := hex.EncodeToString(nonceBig.Bytes())
	return fmt.Sprintf("%s-%s", token, hexEncodedNonce)
}

func extractTokenAndNonce(arg []byte) (string, uint64) {
	argsSplit := bytes.Split(arg, []byte(dctIdentifierSeparator))
	if len(argsSplit) < 2 {
		return string(arg), 0
	}

	if len(argsSplit[1]) <= dctRandomSequenceLength {
		return string(arg), 0
	}

	identifier := []byte(fmt.Sprintf("%s-%s", argsSplit[0], argsSplit[1][:dctRandomSequenceLength]))
	nonce := big.NewInt(0).SetBytes(argsSplit[1][dctRandomSequenceLength:])

	return string(identifier), nonce.Uint64()
}

func isEmptyAddr(addrLength int, address []byte) bool {
	emptyAddr := make([]byte, addrLength)

	return bytes.Equal(address, emptyAddr)
}

func isASCIIString(input string) bool {
	for i := 0; i < len(input); i++ {
		if input[i] > unicode.MaxASCII {
			return false
		}
	}

	return true
}
