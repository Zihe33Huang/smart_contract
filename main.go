package main

import (
	"awesomeProject2/contractConfig"
	"awesomeProject2/dbConfig"
	do "awesomeProject2/do"
	"awesomeProject2/method"
	"fmt"
	"github.com/Conflux-Chain/go-conflux-sdk/types"
	"github.com/Conflux-Chain/go-conflux-sdk/types/cfxaddress"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

// 合约迁移到测试网
func main() {
	//approveForAll()   // 所有地址给operator授权
	transferForAll() // 把nft transfer给正确的主人
	//fix()				// 修复数据库
	//method.GiveMoney()			//给每个账号打gas费
}

// transfer
func transferForAll() {
	// 合约地址
	contractAddress := cfxaddress.MustNew(contractConfig.ContractAddress)
	// 1、 从奈雪数据库拿出nft数据
	var nfts []do.NFTS
	err := dbConfig.NaixueDB.Table("nfts").Select("tokenid", "address").Scan(&nfts).Error
	if err != nil {
		fmt.Println(err)
		return
	}
	var nftAddress []string
	for _, nft := range nfts {
		nftAddress = append(nftAddress, nft.Address)
	}
	// 2、 从中台数据库取出对应的用户
	var users []do.UserInfo
	err = dbConfig.MiddleDB.Debug().Table("user").Select("path", "address").Where("address IN ?", nftAddress).Scan(&users).Error
	if err != nil {
		fmt.Println(err)
		return
	}
	//addressToPath := make(map[string]string, 300)
	//for _, user := range users {
	//	addressToPath[user.Address] = user.Path
	//}

	// 3、遍历每个nft
	count := 0
	for _, nft := range nfts {
		// 3.1、 调用ownerOf, 找到nft当前的owner
		tokenId, _ := new(big.Int).SetString(nft.TokenId, 10)
		res := &struct {
			common.Address
		}{}
		err = contractConfig.Contract.Call(nil, res, "ownerOf", tokenId)
		if err != nil {
			fmt.Println(err)
		}

		var networkID uint32 = 1
		// 3.2、 构建from 和 to 的地址
		// from就是当前合约中，该nft的owner
		from, _ := cfxaddress.New(res.Address.String(), networkID)
		// to就是数据库中，该nft本应属于的owner
		to, _ := cfxaddress.NewFromBase32(nft.Address)
		// ===================================

		// 3.3、 如果地址相同, 则不用操作
		if from.String() == to.String() {
			count++
		} else {
			// 3.4、 若地址不同，则先构建transaction
			methodData, err := contractConfig.Contract.GetData("transferFrom", from.MustGetCommonAddress(), to.MustGetCommonAddress(), tokenId)
			if err != nil {
				fmt.Println(err)
			}
			tx := new(types.UnsignedTransaction)
			tx.Data = methodData
			tx.To = &contractAddress
			tx.From = &from
			fmt.Println("from: ", from)
			fmt.Println("to: ", to)
			fmt.Println("tokenId: ", tokenId)
			err = contractConfig.Client.ApplyUnsignedTransactionDefault(tx)
			if err != nil {
				fmt.Println(err)
			}
			// 3.5、 将unsignedTransaction给covault签名
			var path string
			// 调vault sign
			for _, user := range users {
				if user.Address == from.MustGetBase32Address() {
					path = user.Path
					break
				}
			}
			sig, err := method.SignTxByVault(tx, path)
			if err != nil {
				fmt.Println(err)
			}
			var rawData []byte
			rawData, err = tx.EncodeWithSignature(sig[64], sig[0:32], sig[32:64])
			// 3.6、发送transaction
			transaction, err := contractConfig.Client.SendRawTransaction(rawData)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(transaction)
		}
	}
	fmt.Println("地址没问题的nft数量: ", count)
}
