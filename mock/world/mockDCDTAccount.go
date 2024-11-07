package worldmock

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/kalyan3104/k-chain-core-go/core"
	"github.com/kalyan3104/k-chain-core-go/data/dcdt"
)

// ErrNegativeValue signals that a negative value has been detected and it is not allowed
var ErrNegativeValue = errors.New("negative value")

// MakeTokenKey creates the storage key corresponding to the given tokenName.
func MakeTokenKey(tokenName []byte, nonce uint64) []byte {
	nonceBytes := big.NewInt(0).SetUint64(nonce).Bytes()
	tokenKey := append(DCDTTokenKeyPrefix, tokenName...)
	tokenKey = append(tokenKey, nonceBytes...)
	return tokenKey
}

// MakeTokenRolesKey creates the storage key corresponding to the roles for the
// given tokenName.
func MakeTokenRolesKey(tokenName []byte) []byte {
	tokenRolesKey := append(DCDTRoleKeyPrefix, tokenName...)
	return tokenRolesKey
}

// MakeLastNonceKey creates the storage key corresponding to the last nonce of
// the given tokenName.
func MakeLastNonceKey(tokenName []byte) []byte {
	tokenNonceKey := append(DCDTNonceKeyPrefix, tokenName...)
	return tokenNonceKey
}

// IsDCDTKey returns true if the given storage key is DCDT-related
func IsDCDTKey(key []byte) bool {
	return IsTokenKey(key) || IsRoleKey(key) || IsNonceKey(key)
}

// IsTokenKey returns true if the given storage key belongs to an DCDT token.
func IsTokenKey(key []byte) bool {
	return bytes.HasPrefix(key, DCDTTokenKeyPrefix)
}

// IsRoleKey returns true if the given storage key belongs to an DCDT role.
func IsRoleKey(key []byte) bool {
	return bytes.HasPrefix(key, DCDTRoleKeyPrefix)
}

// IsNonceKey returns true if the given storage key belongs to an DCDT nonce.
func IsNonceKey(key []byte) bool {
	return bytes.HasPrefix(key, DCDTNonceKeyPrefix)
}

// GetTokenNameFromKey extracts the token name from the given storage key; it
// does not check whether the key is indeed a token key or not.
func GetTokenNameFromKey(key []byte) []byte {
	return key[len(DCDTTokenKeyPrefix):]
}

// GetTokenBalanceByName returns the DCDT balance of the account, specified by
// the token name.
func (a *Account) GetTokenBalanceByName(tokenName string) (*big.Int, error) {
	tokenKey := MakeTokenKey([]byte(tokenName), 0)
	return a.GetTokenBalance(tokenKey)
}

// GetTokenBalance returns the DCDT balance of the account, specified by the
// token key.
func (a *Account) GetTokenBalance(tokenKey []byte) (*big.Int, error) {
	tokenData, err := a.GetTokenData(tokenKey)
	if err != nil {
		return nil, err
	}

	return tokenData.Value, nil
}

// SetTokenBalance sets the DCDT balance of the account, specified by the token
// key.
func (a *Account) SetTokenBalance(tokenKey []byte, balance *big.Int) error {
	tokenData, err := a.GetTokenData(tokenKey)
	if err != nil {
		return err
	}

	if balance.Sign() < 0 {
		return ErrNegativeValue
	}

	tokenData.Value = balance
	return a.SetTokenData(tokenKey, tokenData)
}

// GetTokenBalanceUint64 returns the DCDT balance of the account, specified by
// the token key.
func (a *Account) GetTokenBalanceUint64(tokenKey []byte) (uint64, error) {
	tokenData, err := a.GetTokenData(tokenKey)
	if err != nil {
		return 0, err
	}

	return tokenData.Value.Uint64(), nil
}

// SetTokenBalanceUint64 sets the DCDT balance of the account, specified by the
// token key.
func (a *Account) SetTokenBalanceUint64(tokenKey []byte, balance uint64) error {
	tokenData, err := a.GetTokenData(tokenKey)
	if err != nil {
		return err
	}

	if balance < 0 {
		return ErrNegativeValue
	}

	tokenData.Value = big.NewInt(0).SetUint64(balance)
	return a.SetTokenData(tokenKey, tokenData)
}

// GetTokenData gets the DCDT information related to a token from the storage of the account.
func (a *Account) GetTokenData(tokenKey []byte) (*dcdt.DCDigitalToken, error) {
	dcdtData := &dcdt.DCDigitalToken{
		Value: big.NewInt(0),
		Type:  uint32(core.Fungible),
		TokenMetaData: &dcdt.MetaData{
			Name:  GetTokenNameFromKey(tokenKey),
			Nonce: 0,
		},
	}

	marshaledData, _, err := a.AccountDataHandler().RetrieveValue(tokenKey)
	if err != nil || len(marshaledData) == 0 {
		return dcdtData, nil
	}

	err = WorldMarshalizer.Unmarshal(dcdtData, marshaledData)
	if err != nil {
		return nil, err
	}

	return dcdtData, nil
}

// SetTokenData sets the DCDT information related to a token into the storage of the account.
func (a *Account) SetTokenData(tokenKey []byte, tokenData *dcdt.DCDigitalToken) error {
	marshaledData, err := WorldMarshalizer.Marshal(tokenData)
	if err != nil {
		return err
	}
	a.Storage[string(tokenKey)] = marshaledData
	return nil
}

// SetTokenRoles sets the specified roles to the account, corresponding to the given tokenName.
func (a *Account) SetTokenRoles(tokenName []byte, roles [][]byte) error {
	tokenRolesKey := MakeTokenRolesKey(tokenName)
	tokenRolesData := &dcdt.DCDTRoles{
		Roles: roles,
	}

	marshaledData, err := WorldMarshalizer.Marshal(tokenRolesData)
	if err != nil {
		return err
	}

	a.Storage[string(tokenRolesKey)] = marshaledData
	return nil
}

// SetTokenRolesAsStrings sets the specified roles to the account, corresponding to the given tokenName.
func (a *Account) SetTokenRolesAsStrings(tokenName []byte, rolesAsStrings []string) error {
	roles := make([][]byte, len(rolesAsStrings))
	for i := 0; i < len(roles); i++ {
		roles[i] = []byte(rolesAsStrings[i])
	}

	return a.SetTokenRoles(tokenName, roles)
}

// SetLastNonce writes the last nonce of a specified DCDT into the storage.
func (a *Account) SetLastNonce(tokenName []byte, lastNonce uint64) error {
	tokenNonceKey := MakeLastNonceKey(tokenName)
	nonceBytes := big.NewInt(0).SetUint64(lastNonce).Bytes()
	a.Storage[string(tokenNonceKey)] = nonceBytes
	return nil
}

// SetLastNonces writes the last nonces of each specified DCDT into the storage.
func (a *Account) SetLastNonces(lastNonces map[string]uint64) error {
	for tokenName, nonce := range lastNonces {
		err := a.SetLastNonce([]byte(tokenName), nonce)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetTokenRoles returns the roles of the account for the specified tokenName.
func (a *Account) GetTokenRoles(tokenName []byte) ([][]byte, error) {
	tokenRolesKey := MakeTokenRolesKey(tokenName)
	tokenRolesData := &dcdt.DCDTRoles{
		Roles: make([][]byte, 0),
	}

	marshaledData, _, err := a.AccountDataHandler().RetrieveValue(tokenRolesKey)
	if err != nil || len(marshaledData) == 0 {
		return tokenRolesData.Roles, nil
	}

	err = WorldMarshalizer.Unmarshal(tokenRolesData, marshaledData)
	if err != nil {
		return nil, err
	}

	return tokenRolesData.Roles, nil

}

// GetTokenKeys returns the storage keys of all the DCDT tokens owned by the account.
func (a *Account) GetTokenKeys() [][]byte {
	tokenKeys := make([][]byte, 0)
	for key := range a.Storage {
		if IsTokenKey([]byte(key)) {
			tokenKeys = append(tokenKeys, []byte(key))
		}
	}

	return tokenKeys
}

// MockDCDTData groups together all instances of a token (same token name, different nonces).
type MockDCDTData struct {
	TokenIdentifier []byte
	Instances       []*dcdt.DCDigitalToken
	LastNonce       uint64
	Roles           [][]byte
}

// GetFullMockDCDTData returns the information about all the DCDT tokens held by the account.
func (a *Account) GetFullMockDCDTData() (map[string]*MockDCDTData, error) {
	resultMap := make(map[string]*MockDCDTData)
	for key := range a.Storage {
		storageKeyBytes := []byte(key)
		if IsTokenKey(storageKeyBytes) {
			tokenName, tokenInstance, err := a.loadMockDCDTDataInstance(storageKeyBytes)
			if err != nil {
				return nil, err
			}
			if tokenInstance.Value.Sign() > 0 {
				resultObj := getOrCreateMockDCDTData(tokenName, resultMap)
				resultObj.Instances = append(resultObj.Instances, tokenInstance)
			}
		} else if IsNonceKey(storageKeyBytes) {
			tokenName := key[len(DCDTNonceKeyPrefix):]
			resultObj := getOrCreateMockDCDTData(tokenName, resultMap)
			resultObj.LastNonce = big.NewInt(0).SetBytes(a.Storage[key]).Uint64()
		} else if IsRoleKey(storageKeyBytes) {
			tokenName := key[len(DCDTRoleKeyPrefix):]
			roles, err := a.GetTokenRoles([]byte(tokenName))
			if err != nil {
				return nil, err
			}
			resultObj := getOrCreateMockDCDTData(tokenName, resultMap)
			resultObj.Roles = roles
		}
	}

	return resultMap, nil
}

// loads and prepared the DCDT instance
func (a *Account) loadMockDCDTDataInstance(tokenKey []byte) (string, *dcdt.DCDigitalToken, error) {
	tokenInstance, err := a.GetTokenData(tokenKey)
	if err != nil {
		return "", nil, err
	}

	tokenNameFromKey := GetTokenNameFromKey(tokenKey)

	var tokenName string
	if tokenInstance.TokenMetaData == nil || tokenInstance.TokenMetaData.Nonce == 0 {
		// DCDT, no nonce in the key
		tokenInstance.TokenMetaData = &dcdt.MetaData{
			Name:  tokenNameFromKey,
			Nonce: 0,
		}
		tokenName = string(tokenNameFromKey)
	} else {
		nonceAsBytes := big.NewInt(0).SetUint64(tokenInstance.TokenMetaData.Nonce).Bytes()
		tokenNameLen := len(tokenNameFromKey) - len(nonceAsBytes)

		if !bytes.Equal(nonceAsBytes, tokenNameFromKey[tokenNameLen:]) {
			return "", nil, errors.New("Invalid key for NFT (key does not end in nonce)")
		}

		tokenName = string(tokenNameFromKey[:tokenNameLen])
	}

	return tokenName, tokenInstance, nil
}

func getOrCreateMockDCDTData(tokenName string, resultMap map[string]*MockDCDTData) *MockDCDTData {
	resultObj := resultMap[tokenName]
	if resultObj == nil {
		resultObj = &MockDCDTData{
			TokenIdentifier: []byte(tokenName),
			Instances:       nil,
			LastNonce:       0,
			Roles:           nil,
		}
		resultMap[tokenName] = resultObj
	}
	return resultObj
}
