package tools

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"encoding/hex"
)

func NewAesTool(appSecret string) (aesTool *AesEncrypt,err error) {
	if len(appSecret) <16 {
		err = fmt.Errorf("invalid param appsecret of %s",appSecret)
		return
	}
	aesTool = &AesEncrypt{AppSecret:appSecret}
	return
}

type AesEncrypt struct {
	AppSecret string
}

func (p *AesEncrypt) getKey() []byte {
	strKey := p.AppSecret
	keyLen := len(strKey)
	if keyLen < 16 {
		panic("res key 长度不能小于16")
	}
	arrKey := []byte(strKey)
	if keyLen >= 32 {
		//取前32个字节
		return arrKey[:32]
	}
	if keyLen >= 24 {
		//取前24个字节
		return arrKey[:24]
	}
	//取前16个字节
	return arrKey[:16]
}

//加密字符串
func (p *AesEncrypt) Encrypt(strMesg string) (string, error) {
	key := p.getKey()
	var iv = []byte(key)[:aes.BlockSize]
	encrypted := make([]byte, len(strMesg))
	aesBlockEncrypter, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesEncrypter := cipher.NewCFBEncrypter(aesBlockEncrypter, iv)
	aesEncrypter.XORKeyStream(encrypted, []byte(strMesg))
	encodeString := fmt.Sprintf("%x",encrypted)
	//encodeString := base64.StdEncoding.EncodeToString([]byte(encrypted))
	return encodeString, nil
}

//解密字符串
func (p *AesEncrypt) Decrypt(aesEncryptString string) (strDesc string, err error) {
	src, err := hex.DecodeString(aesEncryptString)
	//src, err := base64.StdEncoding.DecodeString(aesEncryptString)
	if err != nil {
		return
	}
	defer func() {
		//错误处理
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	key := p.getKey()
	var iv = []byte(key)[:aes.BlockSize]
	decrypted := make([]byte, len(src))
	var aesBlockDecrypter cipher.Block
	aesBlockDecrypter, err = aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	aesDecrypter := cipher.NewCFBDecrypter(aesBlockDecrypter, iv)
	aesDecrypter.XORKeyStream(decrypted, src)
	return string(decrypted), nil
}