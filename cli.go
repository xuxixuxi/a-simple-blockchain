package main

import (
	"fmt"
	"os"
	"strconv"
)

type CLI struct {
	bc *BlockChain
}

const Usage =
	`
		printRChain						'反向打印区块链'
		printChain						'正向打印区块链'
		getBalance --address ADDRESS				'获取指定地址的余额'
		send FROM TO AMOUNT MINER DATA				'由FROM转入TO AMOUNT金额, 由MINER挖矿, 同时写入DATA'
		newWallet 						'创建一个钱包（公钥私钥对）'
		listAddresses						'列举所有的钱包地址'
	`

//According to args to do things
func (cli *CLI)Run()  {
	//Get all args
	args := os.Args
	if len(args) < 2{
		fmt.Println(Usage)
		return
	}
	//Execute commands
	cmd := args[1]
	switch cmd {
	case "printRChain":
		cli.PrintBlockChain()
	case "printChain":
		cli.PrintRBlockChain()
	case "getBalance":
		if len(args) == 4 && args[2] == "--address" {
			address := args[3]
			cli.getBalance(address)
		}else {
			fmt.Println(Usage)
		}
	case "send":
		if len(args) != 7{
			fmt.Println(Usage)
			return
		}
		from, to, miner, data := args[2], args[3], args[5], args[6]
		amount, _ := strconv.ParseFloat(args[4], 64)
		fmt.Println("转账开始...")
		cli.send(from, to, amount, miner, data)
	case "newWallet":
		cli.NewWallet()
	case "listAddresses":
		cli.ListAddresses()
	default:
		fmt.Println(Usage)
	}
}

