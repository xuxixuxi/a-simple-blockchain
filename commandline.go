package main

import (
	"fmt"
	"time"
)

func (cli *CLI)PrintBlockChain()  {

	bc := cli.bc
	bcI := bc.NewIterator()

	for {
		block := bcI.Next()
		timeFormat := time.Unix(int64(block.TimeStamp), 0).Format("2006-01-02 15:04:05")
		fmt.Println("=====================")
		fmt.Printf("version is: %d\n", block.Version)
		fmt.Printf("prevHash is: %x\n", block.PrevHash)
		fmt.Printf("merkel is: %x\n", block.Merkel)
		fmt.Printf("timeStamp is: %s\n", timeFormat)
		fmt.Printf("difficulty is: %x\n", block.Difficulty)
		fmt.Printf("nonce is: %d\n", block.Nonce)
		fmt.Printf("currentHash is: %x\n", block.Hash)
		fmt.Printf("data is: %s\n", block.Transaction[0].TXInputs[0].Sig)
		//fmt.Printf("txHash is: %x\n", block.Transaction[0].TXID)
		//fmt.Printf("txHash is: %x\n", block.Transaction[1].TXID)
		//fmt.Printf("txHash is: %x\n", block.Transaction[1].TXInputs[0].TXID)

		if len(block.PrevHash) == 0 {
			return
		}
	}
}

func (cli *CLI)PrintRBlockChain()  {
	cli.bc.Printchain()
}

func (cli *CLI)getBalance(address string)  {
	utxos := cli.bc.FindUTXOs(address)
	//fmt.Println(utxos)
	total := 0.0

	for _,utxo := range utxos{
		total += utxo.Value
	}

	fmt.Printf(" `%s` 的余额为%f\n", address, total)
}

func (cli *CLI)send(from, to string, amount float64, miner string, data string)  {
	//创建挖矿交易
	coinbase := NewCoinbaseTX(miner, data)
	//创建普通交易
	tx := NewTransaction(from, to, amount, cli.bc)
	if tx == nil{
		return
	}
	//添加到区块中
	cli.bc.AddBlock([]*Transaction{coinbase, tx})
	fmt.Printf("转账成功！\n")
}

func (cli *CLI)NewWallet() {
	wallets := NewWallets()
	wallets.CreateWallet()
}

func (cli *CLI)ListAddresses()  {
	ws := NewWallets()
	addresses := ws.GetAllAddresses()
	for _, address := range addresses{
		fmt.Printf("%s\n", address)
	}
}

