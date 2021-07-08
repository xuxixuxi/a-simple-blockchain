package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
)

//Create the ProofOfWork struct, including block and target
type ProofOfWork struct {
	block *Block
	target *big.Int
}

//
func NewProofOfWork(block *Block)  *ProofOfWork{
	pow := ProofOfWork{
		block: block,
	}

	//The diffculty value specified
	targetStr := "0000100000000000000000000000000000000000000000000000000000000000"
	bigTmp := big.Int{}
	bigTmp.SetString(targetStr,16)
	pow.target = &bigTmp

	return &pow
}

//A function to calculate Hash
func (pow *ProofOfWork)Run() ([]byte, uint64) {
	/*
	The steps of algorithm
	1.assemble data
	2.calculate hash
	3.compare to the exit criteria(that is target)
	*/

	var nonce uint64
	var hash [32]byte
	block := pow.block
	fmt.Println("start mine!")
	for  {
		//1.assemble data
		item := [][]byte{
			Uint64ToBtye(block.Version),
			block.PrevHash,
			block.Merkel,
			Uint64ToBtye(block.TimeStamp),
			Uint64ToBtye(block.Difficulty),
			Uint64ToBtye(nonce),
			//block.Data,
		}
		blockInfo := bytes.Join(item,[]byte{})
		//2.calculate hash
		hash = sha256.Sum256(blockInfo)
		//3.compare to the exit criteria(that is target)
		bigTmp := big.Int{}
		bigTmp.SetBytes(hash[:])
		if (bigTmp.Cmp(pow.target)) == -1{
			fmt.Printf("succeed:%x,",hash)
			fmt.Printf("nonce:%d\n",nonce)
			break
		} else {
			nonce++
		}


	}
	return hash[:], nonce
}