package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey"
	"github.com/stretchr/testify/assert"
)

func TestHashFromPassword(t *testing.T) {
	type args struct {
		password string
	}

	tests := []struct {
		name            string
		args            args
		wantEncodedHash string
		wantErr         bool
	}{
		{
			name: "Success",
			args: args{
				password: "mysecretpassword",
			},
			wantEncodedHash: "$argon2id$v=19$m=65536,t=3,p=2$bW9ja2Vkc2FsdA$4fziyesgS8X1eccXj3E5WacLsqUeX28HHqxprFZHlY8",
			wantErr:         false,
		},
		{
			name: "Error in generateRandomBytes",
			args: args{
				password: "mysecretpassword",
			},
			wantEncodedHash: "",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a monkey patch for the generateRandomBytes function.
			patch := gomonkey.ApplyFunc(generateRandomBytes, func(n uint32) ([]byte, error) {
				if tt.wantErr {
					return nil, fmt.Errorf("error in generateRandomBytes")
				}
				return []byte("mockedsalt"), nil
			})
			defer patch.Reset()

			encodedHash, err := HashFromPassword(tt.args.password)

			assert.Equal(t, tt.wantErr, err != nil, "Error status should match the expectation")
			assert.Equal(t, tt.wantEncodedHash, encodedHash, "Encoded hash should match the expected value")
		})
	}
}

func TestComparePasswordAndHash(t *testing.T) {
	type args struct {
		password string
	}

	tests := []struct {
		name            string
		args            args
		wantEncodedHash string
		wantErr         bool
		wantMatch       bool
	}{
		{
			name: "normal",
			args: args{
				password: "mysecretpassword",
			},
			wantEncodedHash: "$argon2id$v=19$m=65536,t=3,p=2$bW9ja2Vkc2FsdA$4fziyesgS8X1eccXj3E5WacLsqUeX28HHqxprFZHlY8",
			wantErr:         false,
			wantMatch:       true,
		},
		{
			name: "Error in decodeHash",
			args: args{
				password: "mysecretpassword",
			},
			wantEncodedHash: "",
			wantErr:         true,
			wantMatch:       false,
		},
		{
			name: "password not match",
			args: args{
				password: "notmatch",
			},
			wantEncodedHash: "$argon2id$v=19$m=65536,t=3,p=2$bW9ja2Vkc2FsdA$4fziyesgS8X1eccXj3E5WacLsqUeX28HHqxprFZHlY8",
			wantErr:         false,
			wantMatch:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patch := gomonkey.ApplyFunc(decodeHash, func(encodedHash string) (p *params, salt, hash []byte, err error) {
				if tt.wantErr {
					return nil, nil, nil, errors.New("error in decodeHash")
				}

				// Mocked return values for successful decoding.
				p = &params{
					memory:      65536,
					iterations:  3,
					parallelism: 2,
					saltLength:  12,
					keyLength:   32,
				}
				salt, _ = base64.RawStdEncoding.Strict().DecodeString("bW9ja2Vkc2FsdA")
				hash, _ = base64.RawStdEncoding.Strict().DecodeString("4fziyesgS8X1eccXj3E5WacLsqUeX28HHqxprFZHlY8")
				return p, salt, hash, nil
			})
			defer patch.Reset()

			match, err := ComparePasswordAndHash(tt.args.password, tt.wantEncodedHash)
			assert.Equal(t, tt.wantErr, err != nil, "Error status should match the expectation")
			assert.Equal(t, tt.wantMatch, match, "Password and hash match status should match the expectation")
		})
	}
}

func TestGenerateRandomBytes(t *testing.T) {
	tests := []struct {
		name       string
		args       uint32
		wantLength int
		wantErr    bool
	}{

		{
			name:       "normal",
			args:       10,
			wantLength: 10,
			wantErr:    false,
		},
		{
			name:       "Error from rand.Read",
			args:       8,
			wantLength: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches := gomonkey.ApplyFunc(rand.Read, func(b []byte) (int, error) {
				if tt.wantErr {
					return 0, errors.New("err")
				}
				for i := 0; i < len(b); i++ {
					b[i] = byte(i)
				}
				return len(b), nil
			})
			defer patches.Reset()

			got, err := generateRandomBytes(tt.args)

			if tt.wantErr {
				assert.Error(t, err, "generateRandomBytes() should return an error")
				assert.Nil(t, got, "generateRandomBytes() should return a nil byte slice")
			} else {
				assert.NoError(t, err, "generateRandomBytes() should not return an error")
				assert.NotNil(t, got, "generateRandomBytes() should return a non-nil byte slice")
				assert.Equal(t, tt.wantLength, len(got), "generateRandomBytes() length mismatch")
			}
		})
	}
}

func TestDecodeHash(t *testing.T) {
	t.Run("valid hash", func(t *testing.T) {
		encodedHash := "$argon2id$v=19$m=65536,t=3,p=2$2ikdYjIFWtoN9+c87yHywA$nz2mhbQ0wqWB+ozZObSqIWBYqwlzru0mF7Ez+CFwHxE"
		expectedParams := &params{
			memory:      65536,
			iterations:  3,
			parallelism: 2,
			saltLength:  16,
			keyLength:   32,
		}
		expectedSalt, _ := base64.RawStdEncoding.Strict().DecodeString("2ikdYjIFWtoN9+c87yHywA")
		expectedHash, _ := base64.RawStdEncoding.Strict().DecodeString("nz2mhbQ0wqWB+ozZObSqIWBYqwlzru0mF7Ez+CFwHxE")

		patch := gomonkey.ApplyFunc(base64.RawStdEncoding.Strict().DecodeString, func(s string) ([]byte, error) {
			if s == "2ikdYjIFWtoN9+c87yHywA" {
				return expectedSalt, nil
			} else if s == "nz2mhbQ0wqWB+ozZObSqIWBYqwlzru0mF7Ez+CFwHxE" {
				return expectedHash, nil
			}
			return nil, errors.New("mocked error")
		})
		defer patch.Reset()

		params, salt, hash, err := decodeHash(encodedHash)
		assert.NoError(t, err)
		assert.Equal(t, expectedParams, params)
		assert.Equal(t, expectedSalt, salt)
		assert.Equal(t, expectedHash, hash)
	})

	t.Run("invalid hash - incorrect number of values", func(t *testing.T) {
		encodedHash := "dummy$dummy$dummy"
		params, salt, hash, err := decodeHash(encodedHash)
		assert.Nil(t, params)
		assert.Nil(t, salt)
		assert.Nil(t, hash)
		assert.EqualError(t, err, errInvalidHash.Error())
	})

	t.Run("invalid hash - incompatible version", func(t *testing.T) {
		encodedHash := "dummy$dummy$v=20$m=1024,t=2,p=2$dGVzdA==$dGVzdA=="
		params, salt, hash, err := decodeHash(encodedHash)
		assert.Nil(t, params)
		assert.Nil(t, salt)
		assert.Nil(t, hash)
		assert.EqualError(t, err, errIncompatibleVersion.Error())
	})

	t.Run("invalid hash - error decoding salt", func(t *testing.T) {
		encodedHash := "dummy$dummy$v=19$m=1024,t=2,p=2$dGVzdA==$dGVzdA=="
		patch := gomonkey.ApplyFunc(base64.RawStdEncoding.Strict().DecodeString, func(s string) ([]byte, error) {
			if s == "dGVzdA==" {
				return nil, errors.New("mocked error")
			}
			return nil, nil
		})
		defer patch.Reset()

		params, salt, hash, err := decodeHash(encodedHash)
		assert.Nil(t, params)
		assert.Nil(t, salt)
		assert.Nil(t, hash)
		assert.Error(t, err)
	})

	t.Run("invalid hash - error decoding hash", func(t *testing.T) {
		encodedHash := "dummy$dummy$v=19$m=1024,t=2,p=2$SGVsbG8gd29ybGQ$TWFuIGlzIGRpc3Rpbmd1aXNoZWQ="

		params, salt, hash, err := decodeHash(encodedHash)
		assert.Nil(t, params)
		assert.Nil(t, salt)
		assert.Nil(t, hash)
		assert.Error(t, err)
	})

	t.Run("invalid hash - param", func(t *testing.T) {
		encodedHash := "dummy$dummy$v=19$m=1024,t=invalid,p=2$dGVzdA==$dGVzdA=="

		params, salt, hash, err := decodeHash(encodedHash)
		assert.Nil(t, params)
		assert.Nil(t, salt)
		assert.Nil(t, hash)
		assert.Error(t, err)
	})

	t.Run("invalid hash - error get version", func(t *testing.T) {
		encodedHash := "dummy$dummy$v=invalid$m=1024,t=2,p=2$dGVzdA==$dGVzdA=="

		params, salt, hash, err := decodeHash(encodedHash)
		assert.Nil(t, params)
		assert.Nil(t, salt)
		assert.Nil(t, hash)
		assert.Error(t, err)
	})
}
