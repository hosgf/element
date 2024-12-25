package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
	"github.com/gogf/gf/v2/encoding/gbase64"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/hosgf/element/config"
)

type ecb struct {
	b         cipher.Block
	blockSize int
}

type ecbEncrypter ecb
type ecbDecrypter ecb

func EncryptDefault(data string) string {
	return Encrypt(data, config.AesKey)
}

func DecryptDefault(data string) (string, error) {
	return Decrypt(data, config.AesKey)
}

func Encrypt(data, key string) string {
	return gconv.String(gbase64.Encode(aesEcbEncrypt(data, key)))
}

func Decrypt(data, key string) (string, error) {
	if len(data) < 1 {
		return data, nil
	}
	newData, err := gbase64.DecodeToString(data)
	if err != nil {
		return data, err
	}
	newData1, err := aesEcbDecrypt(gconv.Bytes(newData), gconv.Bytes(key))
	if err != nil {
		return data, err
	}
	return gconv.String(newData1), nil
}

func (x *ecbEncrypter) BlockSize() int {
	return x.blockSize
}

func (x *ecbEncrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		fmt.Println("crypto/cipher: input not full blocks")
		return
	}
	if len(dst) < len(src) {
		fmt.Println("crypto/cipher: output smaller than input")
		return
	}
	for len(src) > 0 {
		x.b.Encrypt(dst, src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

func (x *ecbDecrypter) BlockSize() int {
	return x.blockSize
}

func (x *ecbDecrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		fmt.Println("crypto/cipher: input not full blocks")
		return
	}
	if len(dst) < len(src) {
		fmt.Println("crypto/cipher: output smaller than input")
		return
	}
	for len(src) > 0 {
		x.b.Decrypt(dst, src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

func aesEcbEncrypt(src, key string) []byte {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil
	}
	if src == "" {
		return nil
	}
	ecb := newECBEncrypter(block)
	content := []byte(src)
	content = pkcs5Padding(content, block.BlockSize())
	crypted := make([]byte, len(content))
	old := crypted[0]
	ecb.CryptBlocks(crypted, content)
	if old == crypted[0] {
		return nil
	}
	return crypted
}

func aesEcbDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := newECBDecrypter(block)
	origData := make([]byte, len(crypted))
	old := origData[0]
	blockMode.CryptBlocks(origData, crypted)
	if old == origData[0] {
		return nil, errors.New("转换失败")
	}
	origData = pkcs5UnPadding(origData)
	if old == origData[0] {
		return nil, errors.New("转换失败")
	}
	return origData, nil
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	if length-1 < 0 {
		return nil
	}
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func newECBEncrypter(b cipher.Block) cipher.BlockMode {
	return (*ecbEncrypter)(newECB(b))
}

func newECBDecrypter(b cipher.Block) cipher.BlockMode {
	return (*ecbDecrypter)(newECB(b))
}

func newECB(b cipher.Block) *ecb {
	return &ecb{
		b:         b,
		blockSize: b.BlockSize(),
	}
}
