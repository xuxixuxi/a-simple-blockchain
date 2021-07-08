package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
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
	Sig string
}

type TXOuput struct {
	//金额
	Value float64
	//锁定脚本
	PuKKeyHash string
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
	input := tx.TXInputs[0]
	if len(tx.TXInputs) == 1 && len(input.TXID) == 0 && input.Index == -1{
		return true
	}
	return false
}

//挖矿交易
//只有一个input 无需指定签名、id等
func NewCoinbaseTX(address string, data string) (*Transaction) {
	input := TXInput{[]byte{}, -1, data}
	output := TXOuput{reward, address}
	tx := Transaction{[]byte{}, []TXInput{input}, []TXOuput{output}}
	tx.setHash()
	return &tx
}

//普通的转账交易
func NewTransaction(from, to string, amount float64, bc *BlockChain) *Transaction {
	//遍历找到所有有转出交易的utxo(input)
	utxos, resValue := bc.FindNeedUTXOs(from, amount)

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
			input := TXInput{[]byte(id), int64(i), from}
			inputs = append(inputs, input)
		}
	}

	ouput := TXOuput{amount, to}
	ouputs = append(ouputs, ouput)

	//钱多找零
	if resValue > amount{
		ouputs = append(ouputs, TXOuput{resValue - amount, from})
	}

	tx := Transaction{[]byte{}, inputs, ouputs}
	tx.setHash()
	//tx.TXID = txid
	return &tx
}