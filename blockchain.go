package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"time"
)

const BlockChainDB = "blockchain.db"
const BlockBucket = "blockbucket"

// BlockClain 4.define blockchain struct
type BlockChain struct {
	//blocks [] *Block
	db *bolt.DB
	//storage the last block's hash
	tail []byte
}

// NewBlockChain 5. Define a block chain
func NewBlockChain(address string) *BlockChain  {
	//return 	&BlockClain{
	//	[]*Block{genesisBlock},
	//}

	var lastHash []byte
	db, err := bolt.Open(BlockChainDB, 0600, nil)
	//defer db.Close()
	if err != nil {
		log.Fatal("create database failed")
	}
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BlockBucket))
		if bucket == nil{
			bucket,err = tx.CreateBucket([]byte(BlockBucket))
			if err != nil{
				log.Fatal("create bucket failed")
			}

			//Create genesis block
			genesisBlock := GenesisBlock(address)

			//Write message into database
			bucket.Put(genesisBlock.Hash,genesisBlock.Serialize())
			bucket.Put([]byte("LastHashKey"),genesisBlock.Hash)
			lastHash = genesisBlock.Hash
		}else{
			lastHash = bucket.Get([]byte("LastHashKey"))
		}

		return nil
	})
	return &BlockChain{db,lastHash}
}

// GenesisBlock create a genesisiBlock
func GenesisBlock(address string) *Block {
	coinBase := NewCoinbaseTX(address, "创世块")
	coinBases := []*Transaction{coinBase}
	return NewBlock(coinBases, []byte{})
}

// AddBlock 6.add a new block
func (bc *BlockChain)AddBlock(txs []*Transaction)  {
	//found the last block's hash
	lastHash := bc.tail
	db := bc.db
	//create a new block
	//send the new block into the blockchain
	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BlockBucket))
		if bucket == nil{
			log.Fatal("no bucket")
		}else{
			//Write message into database
			block := NewBlock(txs, lastHash)
			bucket.Put(block.Hash,block.Serialize())
			bucket.Put([]byte("LastHashKey"),block.Hash)

			//update the last hash
			bc.tail = block.Hash

		}
		return nil
	})
}

//正向打印区块链
func (bc *BlockChain) Printchain() {

	bcI := bc.NewIterator()
	var blockHeight int
	var blocks []*Block

	for {
		block := bcI.Next()
		blocks = append(blocks, block)
		if block.PrevHash == nil {
			break
		}
	}
	for i := len(blocks) - 1; i > -1; i--{
		timeFormat := time.Unix(int64(blocks[i].TimeStamp), 0).Format("2006-01-02 15:04:05")
		fmt.Printf("=============== 区块高度: %d ==============\n", blockHeight)
		fmt.Printf("版本号: %d\n", blocks[i].Version)
		fmt.Printf("前区块哈希值: %x\n", blocks[i].PrevHash)
		fmt.Printf("梅克尔根: %x\n", blocks[i].Merkel)
		fmt.Printf("时间戳: %s\n", timeFormat)
		fmt.Printf("难度值: %d\n", blocks[i].Difficulty)
		fmt.Printf("随机数 : %d\n", blocks[i].Nonce)
		fmt.Printf("当前区块哈希值: %x\n", blocks[i].Hash)
		fmt.Printf("区块数据 :%s\n", blocks[i].Transaction[0].TXInputs[0].Sig)
		blockHeight++
	}
}

//找到指定地址所有的UTXO，即未消费的
//func (bc *BlockChain)FindUTXOs(address string) []TXOuput {
//	var UTXO []TXOuput
//	//定义一个map来保存消费过的output, key为这个消费过的output的交易id，value值为这个交易中索引的数组
//	spentOutput := make(map[string][]int64)
//
//
//	// 遍历input，找到自己花费过的utxo的集合
//
//	//创建迭代器
//	it := bc.NewIterator()
//
//	//遍历区块
//	for  {
//		block := it.Next()
//
//		//遍历区块中的每笔交易
//		for _, transaction := range block.Transaction{
//			//遍历output，添加该地址有关的到返回的utxo中
//			//这里的i为outputs的下标
//			OUTPUT:
//			for i, output := range transaction.TXOuputs{
//				//过滤，已经消费过的output不用添加进去
//				if spentOutput[string(transaction.TXID)] != nil{
//					for _, j := range spentOutput[string(transaction.TXID)]{
//						/*
//							//找错误, continue只能跳出最近的for循环
//							fmt.Println(j)
//							fmt.Println(i)
//							var a bool
//							a = int64(i) == j
//							fmt.Println(a)
//						*/
//						//标识过下标和循环中的下标对比, 过滤到已经消费的output
//						if int64(i) == j{
//							continue OUTPUT
//						}
//					}
//				}
//
//				if output.PuKKeyHash == address{
//					//fmt.Println(output)
//					UTXO = append(UTXO, output)
//				}
//			}
//			//挖矿交易没有input
//			if !transaction.IsCoinbase(){
//				//遍历input，找到花费过的utxo的集合
//				for _, input := range transaction.TXInputs{
//					if input.Sig == address{
//						//key为签名的那个交易
//						//indexArray := spentOutput[string(input.TXID)]
//						//	//这个数组为签名的那个交易中 已经消费过的output的index值
//						//indexArray = append(indexArray, input.Index)
//						spentOutput[string(input.TXID)] = append(spentOutput[string(input.TXID)], input.Index)
//						//fmt.Println("===========")
//						//fmt.Printf("%x\n", input.TXID)
//						//fmt.Println(spentOutput[string(input.TXID)])
//						//fmt.Println("===========")
//					}
//				}
//			}
//		}
//
//		if len(block.PrevHash) == 0 {
//			//fmt.Println("遍历结束")
//			break
//		}
//	}
//
//
//	return UTXO
//}

//找到指定地址所有的UTXO，即未消费的,优化上面函数
func (bc *BlockChain)FindUTXOs(address string) []TXOuput {
	var UTXO []TXOuput
	txs := bc.FindUTXOsBased(address)
	for _, tx := range txs{
		for _, output := range tx.TXOuputs{
			if address == output.PuKKeyHash{
				UTXO = append(UTXO, output)
			}
		}
	}
	return UTXO
}

// FindNeedUTXOs 根据需求找到合理的utxo集合返回，格式 map[string][]in64 即map[Transaction.TXID] {合适的output的index}
//func (bc *BlockChain)FindNeedUTXOs(from string, amount float64) (map[string][]int64, float64){
//	//合理utxo集合
//	utxos := make(map[string][]int64)
//	//找到钱的总数
//	var cacl float64
//
//
//	//=================================
//	//定义一个map来保存消费过的output, key为这个消费过的output的交易id，value值为这个交易中索引的数组
//	spentOutput := make(map[string][]int64)
//
//
//	// 遍历input，找到自己花费过的utxo的集合
//	//创建迭代器
//	it := bc.NewIterator()
//	//遍历区块
//	for  {
//		block := it.Next()
//
//		//遍历区块中的每笔交易
//		for _, transaction := range block.Transaction{
//			//遍历output，添加该地址有关的到返回的utxo中
//			//这里的i为outputs的下标
//			OUTPUT:
//			for i, output := range transaction.TXOuputs{
//				//过滤，已经消费过的output不用添加进去
//				if spentOutput[string(transaction.TXID)] != nil{
//					for _, j := range spentOutput[string(transaction.TXID)]{
//						//标识过下标和循环中的下标对比, 过滤到已经消费的output
//						if int64(i) == j{
//							continue OUTPUT
//						}
//					}
//				}
//
//				if output.PuKKeyHash == from{
//					//将utxo加进来，统计总额，比较是否是否满足转账需求
//					// 满足则退出并返回
//					//fmt.Println(output)
//					if cacl < amount {
//						//统计金额
//						cacl += output.Value
//						//将对应交易号及output的index添加进map
//						//array := utxos[string(transaction.TXID)]
//						//array = append(array, int64(i))
//						utxos[string(transaction.TXID)] = append(utxos[string(transaction.TXID)], int64(i))
//						if cacl >= amount{
//							fmt.Printf("找到满足的金额%f\n", cacl)
//							return utxos, cacl
//						}
//					}
//				}
//			}
//			//挖矿交易没有input
//			if !transaction.IsCoinbase(){
//				//遍历input，找到花费过的utxo的集合
//				for _, input := range transaction.TXInputs{
//					if input.Sig == from{
//						/*
//						//key为签名的那个交易
//						indexArray := spentOutput[string(input.TXID)]
//						//这个数组为签名的那个交易中 已经消费过的output的index值
//						indexArray = append(indexArray, input.Index)
//						*/
//						spentOutput[string(input.TXID)] = append(spentOutput[string(input.TXID)], input.Index)
//					}
//				}
//			}
//		}
//
//		if len(block.PrevHash) == 0 {
//			//fmt.Println("遍历结束")
//			break
//		}
//	}
//	//=================================
//
//
//
//
//
//	return utxos, cacl
//}

func (bc *BlockChain)FindNeedUTXOs(from string, amount float64) (map[string][]int64, float64){
	//合理utxo集合
	utxos := make(map[string][]int64)
	//找到钱的总数
	var cacl float64

	txs := bc.FindUTXOsBased(from)
	for _, tx := range txs{
		for i, output := range tx.TXOuputs{
			if from == output.PuKKeyHash{
				//将utxo加进来，统计总额，比较是否是否满足转账需求
					// 满足则退出并返回
					//fmt.Println(output)
					if cacl < amount {
						//统计金额
						cacl += output.Value
						//将对应交易号及output的index添加进map
						//array := utxos[string(transaction.TXID)]
						//array = append(array, int64(i))
						utxos[string(tx.TXID)] = append(utxos[string(tx.TXID)], int64(i))
						if cacl >= amount {
							fmt.Printf("找到满足的金额%f\n", cacl)
							return utxos, cacl
						}
					}
			}
		}
	}
	return utxos, cacl
}
//提炼公共基础函数
//有问题，因为该链上每个区块只有一个交易，所有才能用，如果一个连上有多个交易则可能出错！
func (bc *BlockChain)FindUTXOsBased(address string) []*Transaction {
	//var UTXO []TXOuput
	var txs []*Transaction
	//定义一个map来保存消费过的output, key为这个消费过的output的交易id，value值为这个交易中索引的数组
	spentOutput := make(map[string][]int64)


	// 遍历input，找到自己花费过的utxo的集合

	//创建迭代器
	it := bc.NewIterator()

	//遍历区块
	for  {
		block := it.Next()

		//遍历区块中的每笔交易
		for _, transaction := range block.Transaction{
			//遍历output，添加该地址有关的到返回的utxo中
			//这里的i为outputs的下标
		OUTPUT:
			for i, output := range transaction.TXOuputs{
				//过滤，已经消费过的output不用添加进去
				if spentOutput[string(transaction.TXID)] != nil{
					for _, j := range spentOutput[string(transaction.TXID)]{
						/*
							//找错误, continue只能跳出最近的for循环
							fmt.Println(j)
							fmt.Println(i)
							var a bool
							a = int64(i) == j
							fmt.Println(a)
						*/
						//标识过下标和循环中的下标对比, 过滤到已经消费的output
						if int64(i) == j{
							continue OUTPUT
						}
					}
				}

				if output.PuKKeyHash == address{
					//fmt.Println(output)
					txs = append(txs, transaction)
				}
			}
			//挖矿交易没有input
			if !transaction.IsCoinbase(){
				//遍历input，找到花费过的utxo的集合
				for _, input := range transaction.TXInputs{
					if input.Sig == address{
						//key为签名的那个交易
						//indexArray := spentOutput[string(input.TXID)]
						//	//这个数组为签名的那个交易中 已经消费过的output的index值
						//indexArray = append(indexArray, input.Index)
						spentOutput[string(input.TXID)] = append(spentOutput[string(input.TXID)], input.Index)
						//fmt.Println("===========")
						//fmt.Printf("%x\n", input.TXID)
						//fmt.Println(spentOutput[string(input.TXID)])
						//fmt.Println("===========")
					}
				}
			}
		}

		if len(block.PrevHash) == 0 {
			//fmt.Println("遍历结束")
			break
		}
	}


	return txs
}
