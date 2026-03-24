package security

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"crypto/sha256"
	"rgt-server/log"
)

type Cipher interface {
	Encrypt(msg []byte) []byte
	Decrypt(msg []byte) []byte
}

type CipherFactory interface {
	New() Cipher
	NewWithKey(key []byte) Cipher
}

type TripleDESFactory struct{}

type TripleDesCipher struct {
	encrypter cipher.BlockMode
	decrypter cipher.BlockMode
}

var (
	internalKey = []byte{55, 48, 48, 51, 57, 52, 97, 99, 53, 102, 52, 48, 57, 53, 100, 50, 99, 50, 55, 49, 99, 101, 99,
		98, 101, 101, 55, 100, 54, 53, 52, 55, 98, 98, 57, 57, 50, 56, 53, 52, 55, 98, 57, 48, 49, 53,
		55, 97, 56, 98, 48, 101, 98, 49, 49, 53, 48, 97, 52, 49, 101, 49, 99, 57}
	ciphers    map[string]CipherFactory
	defaultKey []byte
)

func init() {
	ciphers = make(map[string]CipherFactory)
	shaKey := sha256.Sum256(internalKey)
	defaultKey = shaKey[:]
	tripleDESFactory := &TripleDESFactory{}
	ciphers["DESede"] = tripleDESFactory
	ciphers["3DES"] = tripleDESFactory
}

func CiphersFactories() map[string]CipherFactory {
	return ciphers
}

func GetCipher(id string) Cipher {
	c, found := ciphers[id]
	if found {
		return c.New()
	} else {
		return nil
	}
}

func RegisterCipher(id string, cipherFactory CipherFactory) {
	ciphers[id] = cipherFactory
}

func UnregisterCipher(id string) {
	delete(ciphers, id)
}

func (f TripleDESFactory) New() Cipher {
	return f.NewWithKey(defaultKey)
}

func (f TripleDESFactory) NewWithKey(key []byte) Cipher {
	var iv []byte
	if len(key) < 24 {
		log.Errorf("Invalid key size for 3DES. Key size: %d.", len(key))
		return nil
	}
	tripleDes, err := des.NewTripleDESCipher(key[:24])
	if err != nil {
		log.Error("Error cretaing 3Des cipher")
		return nil
	}
	if len(key) >= 32 {
		iv = key[24:32]
	} else {
		iv = key[:8]
	}
	return &TripleDesCipher{
		encrypter: cipher.NewCBCEncrypter(tripleDes, iv),
		decrypter: cipher.NewCBCDecrypter(tripleDes, iv)}
}

func NewTripleDES(key string) Cipher {
	var f TripleDESFactory
	shaKey := sha256.Sum256([]byte(key))
	return f.NewWithKey(shaKey[:])
}

func pkcs5padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func pkcs5unpadding(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])
	return src[:(length - unpadding)]
}

func (c *TripleDesCipher) Encrypt(msg []byte) []byte {
	msg = pkcs5padding(msg, c.encrypter.BlockSize())
	encrypted := make([]byte, len(msg))
	c.encrypter.CryptBlocks(encrypted, msg)
	return encrypted
}

func (c *TripleDesCipher) Decrypt(criptBuffer []byte) []byte {
	decrypted := make([]byte, len(criptBuffer))
	c.decrypter.CryptBlocks(decrypted, criptBuffer)
	return pkcs5unpadding(decrypted)
}
