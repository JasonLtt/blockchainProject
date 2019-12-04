package main

import (
	"./base58"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"log"
)

//1.创建一个结构walletkeypair密钥对，保存公钥和秘钥
//2.给这个结构提供一个方法GetAddress：秘钥->公钥->地址

type WalletKeyPair struct {
	PrivateKey *ecdsa.PrivateKey
	//	type PublicKey struct {
	//	elliptic.Curve
	//	X, Y *big.Int
	//	}
	//	将publickey的X和Y进行字节流拼接后传输，之后再进行切割，方便编码
	PublicKey []byte
}

func NewWalletKeyPair() *WalletKeyPair {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err != nil {
		log.Panic(err)
	}

	publicKeyRaw := privateKey.PublicKey
	publicKey := append(publicKeyRaw.X.Bytes(), publicKeyRaw.Y.Bytes()...)

	return &WalletKeyPair{privateKey, publicKey}
}

//通过公钥获得地址
func (w *WalletKeyPair) GetAddress() string {
	//hash := sha256.Sum256(w.PublicKey)
	//// 创建一个hash160对象
	//// 像hash160中write数据
	//// 做哈希运算
	//rip160Hasher := ripemd160.New()
	//_, err := rip160Hasher.Write(hash[:])
	//if err != nil {
	//	log.Panic(err)
	//}
	////sum函数会将我们所需的结果与sum参数append一起返回，我们输入nil，返回还是原值
	//publicHash := rip160Hasher.Sum(nil) //20bytes

	publicHash := HashPubKey(w.PublicKey)

	version := 0x00 //1bytes
	//21bytes
	payload := append([]byte{byte(version)}, publicHash...)

	//first := sha256.Sum256(payload)
	//second := sha256.Sum256(first[:])
	//checksum := second[:4]
	checksum := CheckSum(payload)
	//25bytes
	payload = append(payload, checksum...)

	address := base58.Encode(payload)

	return address
}

//校验传输地址是否正确
func IsValidAddress(address string) bool {
	//1.将输入的地址进行解码到25字节
	decodeInfo := base58.Decode(address)
	if len(decodeInfo) != 25 {
		return false
	}
	payload := decodeInfo[:len(decodeInfo)-4]
	//2.取出前21字节，运行CheckSum，得到checksum1
	checksum1 := CheckSum(payload)
	//3.取出后4字节，得到checksum2
	checksum2 := decodeInfo[len(decodeInfo)-4:]
	//4.checksum1和checksum2进行校验
	if bytes.Equal(checksum1, checksum2) {
		return true
	}
	return false
}

func HashPubKey(pubKey []byte) []byte {
	hash := sha256.Sum256(pubKey)
	// 创建一个hash160对象
	// 像hash160中write数据
	// 做哈希运算
	rip160Hasher := ripemd160.New()
	_, err := rip160Hasher.Write(hash[:])
	if err != nil {
		log.Panic(err)
	}
	//sum函数会将我们所需的结果与sum参数append一起返回，我们输入nil，返回还是原值
	publicHash := rip160Hasher.Sum(nil) //20bytes
	return publicHash
}

func CheckSum(payload []byte) []byte {
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	checksum := second[:4]
	return checksum
}
