package account

import (
	"fmt"
	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	"time"
)

var Am *sdk.AccountManager

func init() {
	//initAccountManager()
	fmt.Println("init account manager done")
	keydir := "./keystore"
	Am = sdk.NewAccountManager(keydir, 1)
	Am.TimedUnlockDefault("huangzihe33", 300*time.Second)
	Am.UnlockDefault("huangzihe33")
}
