package bitcoin_runtime

import (
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"github.com/quocky/taproot-asset/bitcoin_runtime/utils"
)

const (
	MockBtcUser      = "admin"
	MockBtcPass      = "admin123"
	MiningAddress    = ""
	MineTime         = 1 * time.Minute
	MiningAddr       = "SSCs9JaNW6AFtsgdwS6xwHnd8yhZaPdXYH"
	WalletPassphrase = "admin"
)

type BitcoinRuntime struct {
	btcdCmd      *exec.Cmd
	btcWalletCmd *exec.Cmd
}

func New() *BitcoinRuntime {
	return &BitcoinRuntime{}
}

func (b *BitcoinRuntime) SetUpRuntime() error {
	if err := b.startBTCD(); err != nil {
		return err
	}

	time.Sleep(3 * time.Second)

	if err := b.startBtcwallet(); err != nil {
		return err
	}

	// defer func() {
	// 	b.stopBtcd()
	// 	time.Sleep(3 * time.Second)
	// 	b.stopBtcwallet()
	// }()

	return nil
}

func (b *BitcoinRuntime) startBtcwallet() error {
	fmt.Println("start btcwallet ")
	// setup wallet running in simnet mode
	// btcwallet --simnet --noclienttls --noservertls -A wallet --btcdusername=admin --btcdpassword=admin123 -u admin -P admin123
	// btcwallet --create --simnet --noclienttls --noservertls -A wallet --btcdusername=admin --btcdpassword=admin123 -u admin -P admin123
	b.btcWalletCmd = exec.Command("btcwallet", "--simnet", "--noclienttls", "--noservertls", "-A", "wallet", "--logdir", "wallet", "--btcdusername", MockBtcUser, "--btcdpassword", MockBtcPass, "-u", MockBtcUser, "-P", MockBtcPass, "&")
	b.btcWalletCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// determine if there is already a running btcwallet process
	if !utils.IsProcessRunning("btcwallet") {
		if err := b.btcWalletCmd.Start(); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("[StartBtcwallet] btcwallet process already running")
	}

	// wait for wallet to start
	time.Sleep(3 * time.Second)

	return nil
}

func (b *BitcoinRuntime) startBTCD() error {
	// setup bitcoin node running in simnet mode
	rpcUser := fmt.Sprintf("--rpcuser=%s", MockBtcUser)
	rpcPass := fmt.Sprintf("--rpcpass=%s", MockBtcPass)
	// btcd --simnet --txindex --notls --datadir simnet/btcd --logdir simnet/btcd/logs --miningaddr SgWABqYDjsugfAbPZmTniuTHxnZjHzxe5Z --rpcuser=admin --rpcpass=admin123
	b.btcdCmd = exec.Command("btcd", "--simnet", "--txindex", "--notls", "--datadir", "simnet/btcd", "--logdir", "simnet/btcd/logs", "--miningaddr", MiningAddr, rpcUser, rpcPass, "&")
	// set child process group id to the same as parent process id, so that KILL signal can kill both parent and child processes
	b.btcdCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// determine if there is already a running btcd process
	if !utils.IsProcessRunning("btcd") {
		if err := b.btcdCmd.Start(); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("[StartBTCD] btcd process already running")
	}

	// wait for wallet to start
	time.Sleep(3 * time.Second)

	go func() {
		for {
			err := exec.Command("btcctl", "--simnet", "--notls", rpcUser, rpcPass, "generate", "100").Run()
			if err != nil {
				panic(err)
			}
			time.Sleep(MineTime)
		}
	}()

	return nil
}

func (b *BitcoinRuntime) stopBtcwallet() {
	if b.btcWalletCmd != nil {
		err := b.btcWalletCmd.Process.Kill()
		if err != nil {
			panic(err)
		}
	}
}

func (b *BitcoinRuntime) stopBtcd() {
	if b.btcdCmd != nil {
		err := b.btcdCmd.Process.Kill()
		if err != nil {
			panic(err)
		}
	}
}

func (b *BitcoinRuntime) GetNewAddress() error {
	rpcUser := fmt.Sprintf("--rpcuser=%s", MockBtcUser)
	rpcPass := fmt.Sprintf("--rpcpass=%s", MockBtcPass)
	err := exec.Command("btcctl", "--simnet", "--notls", rpcUser, rpcPass, "generate", "1").Run()
	if err != nil {
		return err
	}

	return nil
}
