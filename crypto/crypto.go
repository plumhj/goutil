package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"encoding/binary"
	"encoding/base64"
	"strings"
)

type Crypto interface {
	Encrypt([]byte) []byte
	Decrypt([]byte) []byte
	EncryptInt64ToBase64String(int64) string
	DecryptBase64StringToInt64(string) (int64, error)
}

type AESCrypto struct {
	block cipher.Block
	iv [] byte
}

func NewAESCFB(key, iv string) (Crypto, error) {

	keyb := []byte(key)
	if len(keyb) >= 32 {
		keyb = keyb[:32]
	}else if len(keyb) >= 24 {
		keyb = keyb[:24]
	}else if len(keyb) >= 16 {
		keyb = keyb[:16]
	}else{
		return nil, errors.New("key length error, length >= 16")
	}

	ivb := []byte(iv)
	if len(ivb) < aes.BlockSize {
		copy(ivb, keyb)
	}
	ivb = ivb[:aes.BlockSize]

	aesBlockEncrypter, err := aes.NewCipher(keyb)
	if err != nil {
		return nil, err
	}
	return &AESCrypto{
		aesBlockEncrypter,
		ivb,
	}, nil
}

func (c *AESCrypto) Encrypt(data []byte) []byte {
	buff := make([]byte, len(data))
	cipher.NewCFBEncrypter(c.block, c.iv).XORKeyStream(buff, data)
	return buff
}

func (c *AESCrypto) EncryptInt64ToBase64String(i int64) string {
	buff := make([]byte, 8)
	binary.LittleEndian.PutUint64(buff, uint64(i))
	buff2 := make([]byte, 8)
	cipher.NewCFBEncrypter(c.block, c.iv).XORKeyStream(buff2, buff)
	return strings.Replace(base64.URLEncoding.EncodeToString(buff2), "=", "", -1)
}

func (c *AESCrypto) Decrypt(data []byte) []byte {
	buff := make([]byte, len(data))
	cipher.NewCFBDecrypter(c.block, c.iv).XORKeyStream(buff, data)
	return buff
}

func (c *AESCrypto) DecryptBase64StringToInt64(data string) (int64, error) {

	if l := len(data) % 4; l != 0 {
		data += strings.Repeat("=", 4-l)
	}

	buff, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return -1, err
	}

	dst := make([]byte, len(buff))
	cipher.NewCFBDecrypter(c.block, c.iv).XORKeyStream(dst, buff)

	var i int64 = int64(binary.LittleEndian.Uint64(dst))
	return i, nil
}
