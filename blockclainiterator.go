package main

import (
	"github.com/boltdb/bolt"
	"log"
)

type BlockChainIterator struct {
	db *bolt.DB
	currentHashPointer []byte
}

func (bc *BlockChain)NewIterator() *BlockChainIterator {
	return &BlockChainIterator{bc.db, bc.tail}
}

//iterator:loop shift to left
func (it *BlockChainIterator)Next() *Block{
	var block Block
	it.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BlockBucket))
		if bucket == nil{
			log.Fatal("no bucket")
		}else {
			blockTmp := bucket.Get(it.currentHashPointer)
			//deserialize
			block = DeSerialize(blockTmp)
			it.currentHashPointer = block.PrevHash
		}
		return nil
	})
	return &block
}
