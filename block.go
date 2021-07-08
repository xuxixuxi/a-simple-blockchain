package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"log"
	"time"
)

//1.define struct
type Block struct {
	//1.version
	Version uint64
	//2.the previous hash
	PrevHash []byte
	//3.merkel root
	Merkel []byte
	//4.time stamp
	TimeStamp uint64
	//5.diffculty number
	Difficulty uint64
	//nonce
	Nonce uint64
	//6.the current hash
	Hash []byte
	//7.transaction
	//Data []byte
	Transaction []*Transaction
}
//a function uint64 > []byte
func Uint64ToBtye(num uint64) []byte {
	var buffer bytes.Buffer

	err := binary.Write(&buffer, binary.BigEndian, num)
	if err != nil { fmt.Println("binary.Write failed:", err) }
	return buffer.Bytes()
}

//2.create block
func NewBlock(txs []*Transaction, preHash []byte) (*Block){
	block := Block{
		Version: 			00,
		PrevHash:			preHash,
		Merkel: 			[]byte{},
		TimeStamp: 			uint64(time.Now().Unix()),
		Difficulty: 			00,
		Nonce: 				00,
		Transaction:		txs,
		Hash: 				[]byte{},
	}

	block.Merkel = block.MakeMerkelRoot()
	//block.setHash()
	//Replace the simple function by proof of worf
	pow := NewProofOfWork(&block)
	hash, nonce := pow.Run()
	block.Hash = hash
	block.Nonce = nonce
	return &block
}

//3.create hash
func (block *Block)setHash() {
	//1.assemble data
	/*
	//block.PrevHash no change
	blockInfo := append(block.PrevHash, block.Data...)
	blockInfo = append(blockInfo, Uint64ToBtye(block.Version)...)
	blockInfo = append(blockInfo, block.Merkel...)
	blockInfo = append(blockInfo, Uint64ToBtye(block.TimeStamp)...)
	blockInfo = append(blockInfo, Uint64ToBtye(block.Diffculty)...)
	blockInfo = append(blockInfo, Uint64ToBtye(block.Nonce)...)
	*/
	//Use the API(join) to simplify the above codes
	item := [][]byte{
		Uint64ToBtye(block.Version),
		block.PrevHash,
		block.Merkel,
		Uint64ToBtye(block.TimeStamp),
		Uint64ToBtye(block.Difficulty),
		Uint64ToBtye(block.Nonce),
		//block.Data,
	}
	blockInfo := bytes.Join(item,[]byte{})
	//2.sha256
	hash := sha256.Sum256(blockInfo)
	block.Hash = hash[:]
}

// Serialize
func (block *Block)Serialize() []byte {
	//编码数据放到buffer中
	var buffer bytes.Buffer
	//使用gob进行序列化得到字节流
	//定义编码器
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(&block)
	if(err != nil){
		log.Panic("encode failed")
	}
	return buffer.Bytes()
}
// deSerialize
func DeSerialize(data []byte)  Block{
	//使用gob进反序列化得到Person结构
	//定义解码器
	decoder := gob.NewDecoder(bytes.NewReader(data))
	var block Block
	err := decoder.Decode(&block)
	if err != nil{
		log.Panic("decode failed")
	}
	return block
}

//求梅克尔根
func (block *Block)MakeMerkelRoot()  []byte{
	var info []byte
	for _, tx := range block.Transaction{
		info = append(info, tx.TXID...)
	}
	hash := sha256.Sum256(info)
	return hash[:]
}