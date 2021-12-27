package method

import (
	"awesomeProject2/config"
	"awesomeProject2/contractConfig"
	"awesomeProject2/dbConfig"
	do "awesomeProject2/do"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Conflux-Chain/go-conflux-sdk/types"
	"github.com/Conflux-Chain/go-conflux-sdk/types/cfxaddress"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
	"hash"
	"log"
	"net/http"
	"net/url"
)

type DataOfFix struct {
	UserId  string
	TokenId string
	Address string
	NftID   int
}

func GiveMoney() {
	var dataOfFix []DataOfFix
	// 1、 nfts 和 user_nfts 联表
	err := dbConfig.NaixueDB.Table("nfts_copy1").Distinct("address").Select("user_id, tokenid, address, nft_id").Joins("left join user_nfts on user_nfts.nft_id = nfts_copy1.id").Scan(&dataOfFix).Error
	if err != nil {
		fmt.Println(err)
		return
	}
	// 每个地址打gas费
	value := types.NewBigInt(200000000000000000)
	for _, data := range dataOfFix {
		utx, err := contractConfig.Client.CreateUnsignedTransaction(cfxaddress.MustNewFromBase32("cfxtest:aarjy3cwmayh8h5j3p2fmy56d8dn1178mepu9y1ue6"), cfxaddress.MustNewFromBase32(data.Address), value, nil)
		if err != nil {
			panic(err)
		}
		txhash, err := contractConfig.Client.SendTransaction(utx)
		if err != nil {
			panic(err)
		}
		fmt.Printf("send transaction hash: %v\n\n", txhash)

	}

}

// 数据库修复
func fix() {

	var dataOfFix []DataOfFix
	// 1、 nfts 和 user_nfts 联表
	err := dbConfig.NaixueDB.Table("nfts").Select("user_id, tokenid, address, nft_id").Joins("left join user_nfts on user_nfts.nft_id = nfts.id").Scan(&dataOfFix).Error
	if err != nil {
		fmt.Println(err)
		return
	}
	userIdArr := make([]string, 0)
	for _, data := range dataOfFix {
		userIdArr = append(userIdArr, data.UserId)
	}

	// 2、 中台用户数据
	var users []do.UserInfo
	err = dbConfig.MiddleDB.Debug().Table("user_real").Select("userid", "address").Where("userid IN ?", userIdArr).Scan(&users).Error
	if err != nil {
		fmt.Println(err)
		return
	}
	// 中台数据放入map
	m := make(map[string]string, 300)
	for _, user := range users {
		m[user.UserId] = user.Address
	}
	count := 0
	// 3、 对比数据
	for _, data := range dataOfFix {
		address, ok := m[data.UserId]
		if !ok {
			continue
		}
		if address != data.Address {
			fmt.Println(data.UserId)
			fmt.Println(address)
			fmt.Println(data.Address)
			fmt.Println("=================================")
			// 更新奈雪数据库
			dbConfig.NaixueDB.Debug().Table("nfts").Where("id = ?", data.NftID).Update("address", address)
			count++
		}
	}
	fmt.Println(count)

}

// 给地址授权
func approveForAll() {

	// 1.2、 从奈雪数据库拿出数据
	var nfts []do.NFTS
	err := dbConfig.NaixueDB.Table("nfts_copy1").Select("tokenid", "address").Scan(&nfts).Error
	if err != nil {
		fmt.Println(err)
		return
	}
	var nftAddress []string
	for _, nft := range nfts {
		nftAddress = append(nftAddress, nft.Address)
	}
	// 1.3、 从中台数据库取path
	var users []do.UserInfo
	err = dbConfig.MiddleDB.Debug().Table("user_copy1").Select("path", "address").Where("address IN ?", nftAddress).Scan(&users).Error
	if err != nil {
		fmt.Println(err)
		return
	}
	// 2、 合约操作
	// 2.1、 拿到要被授权的地址
	approvedAddress, err := cfxaddress.NewFromBase32("cfxtest:aarjy3cwmayh8h5j3p2fmy56d8dn1178mepu9y1ue6")
	if err != nil {
		fmt.Println(err)
	}
	//list := account.Am.List()
	//for _, address := range list {
	//	fmt.Println(address)
	//	approvedAddress = address
	//}
	//clientTest.SetAccountManager(account.Am)
	contractAddress := cfxaddress.MustNew(contractConfig.ContractAddress)

	//  循环调用数据库中的data， 给approvedAddress 授权
	for _, nft := range nfts {
		from, err := cfxaddress.NewFromBase32(nft.Address)
		//to, err := cfxaddress.NewFromBase32("cfxtest:aarjy3cwmayh8h5j3p2fmy56d8dn1178mepu9y1ue6")
		if err != nil {
			fmt.Println(err)
		}

		methodData, err := contractConfig.Contract.GetData("setApprovalForAll", approvedAddress.MustGetCommonAddress(), true)
		if err != nil {
			fmt.Println(err)
			return
		}
		// 2、 unsignedTransaction
		tx := new(types.UnsignedTransaction)
		tx.Data = methodData
		tx.To = &contractAddress
		tx.From = &from
		err = contractConfig.Client.ApplyUnsignedTransactionDefault(tx)
		if err != nil {
			fmt.Println(err)
		}
		var path string
		var rawData []byte
		// 调vault sign
		for _, user := range users {
			if user.Address == nft.Address {
				path = user.Path
				break
			}
		}
		sig, err := SignTxByVault(tx, path)
		if err != nil {
			fmt.Println(err)
		}
		rawData, err = tx.EncodeWithSignature(sig[64], sig[0:32], sig[32:64])
		if err != nil {
			fmt.Println(err)
			return
		}

		// 发送
		transaction, err := contractConfig.Client.SendRawTransaction(rawData)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(transaction)
	}

}

func SignTxByVault(tx *types.UnsignedTransaction, id string) ([]byte, error) {
	hash, err := tx.Hash()
	if err != nil {
		return nil, err
	}
	httpurl := config.Config().GetString("vault.url")
	sk := config.Config().GetString("vault.sk")
	skdata, err := hexutil.Decode(sk)
	if err != nil {
		return nil, err
	}
	key, err := crypto.ToECDSA(skdata)
	if err != nil {
		return nil, err
	}
	method := "/generateSig"
	//index := strconv.Itoa(id)
	index := id
	dgst := Keccak256([]byte(index), []byte(hexutil.Encode(hash)))
	mac, err := crypto.Sign(dgst, key)
	data := url.Values{
		"path": {index},
		"msg":  {hexutil.Encode(hash)},
		"mac":  {hexutil.Encode(mac)},
	}
	resp, err := http.PostForm(httpurl+method, data)
	if err != nil {
		return nil, err
	}
	var res map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&res)
	sig, ok := res["data"].(string)
	if !ok {
		log.Println("Assertion failed")
		return nil, errors.New("assertion failed")
	}
	result, err := hexutil.Decode(sig)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func Keccak256(data ...[]byte) []byte {
	b := make([]byte, 32)
	d := NewKeccakState()
	for _, b := range data {
		d.Write(b)
	}
	d.Read(b)
	return b
}

func NewKeccakState() KeccakState {
	return sha3.NewLegacyKeccak256().(KeccakState)
}

//SHA3
type KeccakState interface {
	hash.Hash
	Read([]byte) (int, error)
}
