package main

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"github.com/btcutil/base58"
	"io/ioutil"
	"log"
	"os"
)

const walletFile = `wallet.dat`

//该结构包含所有的wallet及其地址
type Wallets struct {
	//map[地址] 钱包
	WalletsMap map[string] *Wallet
}

//创建一个新钱包
func NewWallets() *Wallets {
	var ws Wallets
	ws.WalletsMap = make(map[string]*Wallet)
	ws.LoadFile()
	//fmt.Println(ws)
	return &ws
}
func (ws *Wallets)CreateWallet() string {
	wallet := NewWallet()
	address := wallet.NewAddress()

	ws.WalletsMap[address] = wallet

	ws.saveToFile()
	return address
}

//保存方法，把新建的wallet添加进去
func (ws *Wallets) saveToFile() {

	var buffer bytes.Buffer

	//panic: gob: type not registered for interface: elliptic.p256Curve
	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(ws)
	//一定要注意校验！！！
	if err != nil {
		log.Panic(err)
	}

	err = ioutil.WriteFile(walletFile, buffer.Bytes(), 0600)
	if err != nil {
		log.Panic(err)
	}
}

//读取方法，将所有的wallet读出来
func (ws *Wallets)LoadFile()  {
	//在读取之前，要先确认文件是否在，如果不存在，直接退出
	_, err := os.Stat(walletFile)
	if os.IsNotExist(err) {
		//ws.WalletsMap = make(map[string]*Wallet)
		return
	}
	//读取内容
	content, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Fatal(err)
	}

	//解码
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(content))

	var wsLocal Wallets

	err = decoder.Decode(&wsLocal)
	if err != nil {
		log.Fatal(err)
	}

	//ws = &wsLocal
	ws.WalletsMap = wsLocal.WalletsMap
}
//获得所有的钱包地址
func (ws *Wallets)GetAllAddresses() []string {
	var addresses []string
	for address := range ws.WalletsMap{
		addresses = append(addresses, address)
	}
	return addresses
}

//通过地址反推公钥hash
func GetPubKeyFromAddress(address string) []byte {
	//1.解码
	addressByte := base58.Decode(address)
	//2.去除一字节的版本号，四字节的校验码
	pubKeyHash := addressByte[1 : len(addressByte) - 4]
	return pubKeyHash
}