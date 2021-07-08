package main

import (
	_ "crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"github.com/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
	"log"
)


type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	//PubKey *ecdsa.PublicKey
	//type PublicKey struct {
	//elliptic.Curve
	//X, Y *big.Int
	//}
	//这里不用上面的公钥结构，换用结构里 的X，Y拼接的形式，然后再校验端进行拆分
	PubKey []byte
}

//创建钱包
func NewWallet() *Wallet {
	//创建曲线
	curve := elliptic.P256()
	//生成私钥
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	//生成公钥
	pubKeyOrig := privateKey.PublicKey
	pubKey := append(pubKeyOrig.X.Bytes(), pubKeyOrig.Y.Bytes()...)
	//big.int{}的.byte方法将big.int转为[]byte
	//[]byte的setBytes()方法将[]byte转为big.int{}
	wallet := Wallet{privateKey, pubKey}
	return &wallet
}

//生成地址
func (wallet *Wallet)NewAddress() string {
	pubKey := wallet.PubKey

	ripe160HashValue := HashPubKey(pubKey)

	//将该hash与version进行拼接
	version := byte(00)  //TODO
	payload := append([]byte{version}, ripe160HashValue...)

	//checksum过程
	hash1 := sha256.Sum256(payload)
	hash2 := sha256.Sum256(hash1[:])

	//4字节校验码
	checkCode := hash2[:4]

	//最后凭借的25字节数据
	payload = append(payload, checkCode...)

	//base58生成地址
	address := base58.Encode(payload)

	return address
}
func HashPubKey(data []byte) []byte {
	hash := sha256.Sum256(data)
	//编码器
	ripe160hasher := ripemd160.New()
	_, err := ripe160hasher.Write(hash[:])
	if err != nil{
		log.Panic(err)
	}
	ripe160HashValue := ripe160hasher.Sum(nil)
	return ripe160HashValue
}

