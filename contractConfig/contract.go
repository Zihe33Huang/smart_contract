package contractConfig

import (
	"awesomeProject2/account"
	"awesomeProject2/config"
	"fmt"
	conflux "github.com/Conflux-Chain/go-conflux-sdk"
	"github.com/Conflux-Chain/go-conflux-sdk/types/cfxaddress"
	"io/ioutil"
)

var Contract *conflux.Contract
var Client *conflux.Client
var ContractAddress string

func init() {
	ContractAddress = config.Config().GetString("contract.contractAddress")
	net := config.Config().GetString("contract.net")
	var err error
	Client, err = conflux.NewClient(net)
	Client.SetAccountManager(account.Am)
	contractAddress := cfxaddress.MustNewFromBase32(ContractAddress)
	abiTest, err := ioutil.ReadFile("abiTest.abi")
	if err != nil {
		panic(err)
	}

	Contract, err = Client.GetContract(abiTest, &contractAddress)
	if err != nil {
		fmt.Println(err)
	}
}
