package encryption

import (
	"encoding/base64"

	"github.com/pkg/errors"
	"github.com/nguyentb/hlf_cclib/convert"
	"github.com/nguyentb/hlf_cclib/router"
	"github.com/nguyentb/hlf_cclib/state"
)

var (
	// ErrKeyNotDefinedInTransientMap occurs when key not defined in transient map
	ErrKeyNotDefinedInTransientMap = errors.New(`key not defined in transient map`)
)

// State encrypting the data before putting to state and decrypting the data after getting from state
func State(c router.Context, key []byte) (state.State, error) {
	s := state.New(c.Stub())

	s.KeyParts = KeyPartsEncryptedWith(key)
	s.StateGetTransformer = DecryptBytesWith(key)
	s.StatePutTransformer = EncryptBytesWith(key)

	return s, nil
}

// KeyFromTransient gets key for encrypting/decrypting from transient map
func KeyFromTransient(c router.Context) ([]byte, error) {
	tm, _ := c.Stub().GetTransient()
	key, ok := tm[TransientMapKey]
	if !ok {
		return nil, ErrKeyNotDefinedInTransientMap
	}

	return key, nil
}

// StateWithTransientKey creates encrypted state state with provided key for symmetric encryption/decryption
func StateWithTransientKey(c router.Context) (state.State, error) {
	key, err := KeyFromTransient(c)
	if err != nil {
		return nil, err
	}
	return State(c, key)
}

// KeyPartsEncryptedWith encrypts key parts
func KeyPartsEncryptedWith(encryptKey []byte) state.KeyPartsTransformer {
	return func(key interface{}) ([]string, error) {
		keyParts, err := state.KeyParts(key)

		if err != nil {
			return nil, err
		}
		for i, p := range keyParts {
			encP, err := Encrypt(encryptKey, p)
			if err != nil {
				return nil, errors.Wrap(err, `key part encrypt error`)
			}
			keyParts[i] = base64.StdEncoding.EncodeToString(encP)
		}
		return keyParts, nil
	}
}

// DecryptBytesWith decrypts by with key - used for decrypting data after reading from state
func DecryptBytesWith(key []byte) state.FromBytesTransformer {
	return func(bb []byte, config ...interface{}) (interface{}, error) {
		decrypted, err := Decrypt(key, bb)
		if err != nil {
			return nil, errors.Wrap(err, `decrypt bytes`)
		}
		if len(config) == 0 {
			return decrypted, nil
		}
		return convert.FromBytes(decrypted, config[0])
	}
}

// EncryptBytesWith encrypts bytes with key - used for encrypting data for state
func EncryptBytesWith(key []byte) state.ToBytesTransformer {
	return func(v interface{}, config ...interface{}) ([]byte, error) {
		bb, err := convert.ToBytes(v)
		if err != nil {
			return nil, err
		}
		return Encrypt(key, bb)
	}
}
