/*
 * @Author: flwfdd
 * @Date: 2024-06-06 17:19:20
 * @LastEditTime: 2024-06-06 19:47:08
 * @Description: EncryptPassword.js
 * _(:з」∠)_
 */

package webvpn

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"math/rand"
)

const aesChars = "ABCDEFGHJKMNPQRSTWXYZabcdefhijkmnprstwxyz2345678"

func pad(data []byte) []byte {
	blockSize := 16
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

func encryptAES(data, key, iv string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	paddedData := pad([]byte(data))
	ciphertext := make([]byte, len(paddedData))
	mode := cipher.NewCBCEncrypter(block, []byte(iv))
	mode.CryptBlocks(ciphertext, paddedData)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func randomString(length int) string {
	result := make([]byte, length)

	for i := 0; i < length; i++ {
		result[i] = aesChars[rand.Intn(len(aesChars))]
	}

	return string(result)
}

func EncryptPassword(password, salt string) (string, error) {
	if salt == "" {
		return password, nil
	}

	data := randomString(64) + password
	iv := randomString(16)
	encrypted, err := encryptAES(data, salt, iv)
	if err != nil {
		return "", err
	}

	return encrypted, nil
}
