package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"github.com/btcutil/base58"
	"log"
	"math/big"
	"strings"
)

const reward = 50

//交易结构定义
type Transaction struct {
	TXID []byte
	TXInputs []TXInput
	TXOuputs []TXOuput
}

type TXInput struct {
	//引用的交易的ID
	TXID []byte
	//引用的output的索引(即TXTD中的第index个output)
	Index int64
	//解锁脚本
	//Sig string
	Signature []byte //数字签名 r，s拼接的[]byte数组
	PubKey []byte //wallet里的公钥格式
}

type TXOuput struct {
	//金额
	Value float64
	//锁定脚本
	//PuKKeyHash string
	PubKeyHash []byte //收款方的公钥的hash
}
//与之前直接赋值不同，现在只知道地址，需要通过地址值反推出公钥的hash，见图
//通过Lock来处理
func (output *TXOuput)Lock(address string)  {
	//由图可知
	//1.解码
	addressByte := base58.Decode(address)
	//2.去除一字节的版本号，四字节的校验码
	pubKey := addressByte[1 : len(addressByte) - 4]

	output.PubKeyHash = pubKey
}
//提供一个创建TXOuput的方法
func NewTXOutput(value float64, address string) *TXOuput {
	output := TXOuput{Value:value}
	output.Lock(address)
	return &output
}

//设置交易ID
func (tx *Transaction)setHash()  {
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(&tx)
	if err != nil {
		log.Panic(err)
	}

	data := buffer.Bytes()
	hash := sha256.Sum256(data)
	tx.TXID = hash[:]
}
//判断是否为挖矿交易
func (tx *Transaction)IsCoinbase() bool {
	if len(tx.TXInputs) == 1 && len(tx.TXInputs[0].TXID) == 0 && tx.TXInputs[0].Index == -1 {
		return true
	}
	return false
}

//挖矿交易
//只有一个input 无需指定签名、id等
func NewCoinbaseTX(address string, data string) (*Transaction) {
	//挖矿交易无须指定签名,pubKey字段可以由矿工自由指定，一般写矿池名
	//签名展示为空
	input := TXInput{[]byte{}, -1, nil, []byte(data)}
	//output := TXOuput{reward, address}
	output := *NewTXOutput(reward, address)
	tx := Transaction{[]byte{}, []TXInput{input}, []TXOuput{output}}
	tx.setHash()
	return &tx
}

//普通的转账交易
func NewTransaction(from, to string, amount float64, bc *BlockChain) *Transaction {
	//1.创建交易需要数字签名->需要私钥->打开钱包
	ws := NewWallets()
	//2.根据地址找到自己的wallet
	wallet := ws.WalletsMap[from]
	if wallet == nil {
		//可能没有该地址
		fmt.Println("没有找到该地址，创建交易失败！！！")
		return nil
	}
	//3.找到对应的公钥私钥
	pubKey := wallet.PubKey
	private := wallet.PrivateKey

	//传递公钥hash
	pubKeyHash := HashPubKey(pubKey)

	//遍历找到所有有转出交易的utxo(input)
	utxos, resValue := bc.FindNeedUTXOs(pubKeyHash, amount)

	if resValue < amount{
		fmt.Println("余额不足，交易失败！")
		return nil
	}

	//储存所有的input，output结构
	var inputs []TXInput
	var ouputs []TXOuput
	//遍历得到相应的output，即交易的TXID，写入新区块的input中
	for id, indexArry := range utxos{
		//fmt.Printf("%s\n",id)
		//fmt.Println(indexArry)
		//fmt.Println("+++++++++")
		for _, i := range indexArry{
			input := TXInput{[]byte(id), int64(i), nil, pubKey}
			inputs = append(inputs, input)
		}
	}

	//创建输出交易
	//ouput := TXOuput{amount, to}
	ouput := NewTXOutput(amount, to)
	ouputs = append(ouputs, *ouput)

	//钱多找零
	if resValue > amount{
		ouput = NewTXOutput(resValue - amount, from)
		ouputs = append(ouputs, *ouput)
		//ouputs = append(ouputs, TXOuput{resValue - amount, from})
	}

	tx := Transaction{[]byte{}, inputs, ouputs}
	tx.setHash()
	//tx.TXID = txid

	//交易创建完后进行签名
	bc.SignTranscation(&tx, private)

	return &tx
}
//签名的具体实验，
//参数为：私钥，inputs里所有的交易  map[string]transaction  (transaction为交易号)
func (tx *Transaction)Sign(private *ecdsa.PrivateKey, prevTXs map[string]Transaction)  {
	if tx.IsCoinbase(){
		return
	}
/*
	签名原理：所签数据为：该交易所引用的output，已经该交易中的output（钱到哪以及钱给谁）
	1.创建一个当前交易的txCopy：TrimmedCopy，其中所有的input的Signature，PubKey设nil
	2.循环遍历txCopy中的inputs，将每个input中的PubKey设置为该input所引用的output的公钥hash
		(prevTXs中为每个所用到的hash)由此可以得到每个input所引用的output的公钥hash
		(input里有引用的交易名，及output的index，所以在prevTXs查找即可知道其对应output的公钥hash)
	3.对交易中的每个input进行签名
	4.放到signature中
*/
	//1.txCopy
	txCopy := tx.TrimmedCopy()
	//2.循环遍历txCopy的inputs
	for i, input := range txCopy.TXInputs{
		txPre := prevTXs[string(input.TXID)]
		//fmt.Println(string(input.TXID))
		//fmt.Println(len(txPre.TXID))
		//fmt.Printf("%x",txPre.TXID)
		if len(txPre.TXID) == 0{
			log.Panic("引用交易无效")
		}
		//for i, ouput := range tx.TXOuputs{
		//	if int64(i) == input.Index{
		//		input.PubKey = ouput.PubKeyHash
		//		continue
		//	}
		//}
		txCopy.TXInputs[i].PubKey = txPre.TXOuputs[input.Index].PubKeyHash

		//对该input进行签名
		//1.设置hash
		txCopy.setHash()
		//2.还原PubKey，不影响后面签名
		txCopy.TXInputs[i].PubKey = nil
		//3.开始签名
		signDataHash := txCopy.TXID
		r, s, err := ecdsa.Sign(rand.Reader, private, signDataHash)
		if err != nil{
			log.Panic("签名错误")
		}
		signature := append(r.Bytes(), s.Bytes()...)
		//4.赋值
		tx.TXInputs[i].Signature = signature
	}
}
func (tx *Transaction)TrimmedCopy() *Transaction {
	var inputs []TXInput
	var ouputs []TXOuput

	for _, input := range tx.TXInputs{
		inputs = append(inputs, TXInput{input.TXID, input.Index, nil, nil})
	}

	for _, ouput := range tx.TXOuputs{
		ouputs = append(ouputs, ouput)
	}

	return &Transaction{tx.TXID, inputs, ouputs}
}

//校验过程
//校验所需：公钥，数据（TrimmedCopy，hash生成），signature
func (tx *Transaction)Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase(){
		return true
	}
	//1.得到签名数据
	txCopy := tx.TrimmedCopy()
	for i, input := range tx.TXInputs{
		txPre := prevTXs[string(input.TXID)]
		if len(txPre.TXID) == 0{
			log.Panic("引用交易无效")
		}
		txCopy.TXInputs[i].PubKey = txPre.TXOuputs[input.Index].PubKeyHash
		txCopy.setHash()
		txCopy.TXInputs[i].PubKey = nil
		dataHash := txCopy.TXID
		//2.通过signature，反推回r，s
		signature := input.Signature
		R := big.Int{}
		S := big.Int{}
		R.SetBytes(signature[:len(signature) / 2])
		S.SetBytes(signature[len(signature) / 2:])
		//3.通过PubKey，反推回验证所需的公钥格式
		pubKey := input.PubKey
		X := big.Int{}
		Y := big.Int{}
		X.SetBytes(pubKey[:len(pubKey) / 2])
		Y.SetBytes(pubKey[len(pubKey) / 2:])
		pubKeyOri := ecdsa.PublicKey{elliptic.P256(), &X,&Y}
		//4.校验
		res := ecdsa.Verify(&pubKeyOri, dataHash[:], &R, &S)
		if !res{
			return false
		}
	}
	return true
}

func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.TXID))

	for i, input := range tx.TXInputs {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.TXID))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Index))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.TXOuputs{
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %f", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}

