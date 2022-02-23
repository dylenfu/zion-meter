package sdk

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"
	"time"

	stat "github.com/dylenfu/zion-meter/pkg/go_abi/stat_abi"
	"github.com/dylenfu/zion-meter/pkg/log"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type Account struct {
	pk      *ecdsa.PrivateKey
	address common.Address
	nonce   uint64
	nonceMu sync.Mutex

	url       string
	client    *ethclient.Client
	rpcClient *rpc.Client
	signer    types.EIP155Signer
}

func MasterAccount(url string, chainID uint64, hexkey string) (*Account, error) {
	pk, err := crypto.HexToECDSA(hexkey)
	if err != nil {
		return nil, err
	}

	addr := crypto.PubkeyToAddress(pk.PublicKey)

	signer := types.NewEIP155Signer(new(big.Int).SetUint64(chainID))
	rpcclient, err := rpc.Dial(url)
	if err != nil {
		return nil, err
	}
	client := ethclient.NewClient(rpcclient)

	nonce, err := client.NonceAt(context.Background(), addr, nil)
	if err != nil {
		return nil, err
	}

	return &Account{
		pk:        pk,
		address:   addr,
		nonce:     nonce,
		url:       url,
		signer:    signer,
		client:    client,
		rpcClient: rpcclient,
	}, nil
}

func NewAccount(url string, chainID uint64) (*Account, error) {
	pk, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	addr := crypto.PubkeyToAddress(pk.PublicKey)

	signer := types.NewEIP155Signer(new(big.Int).SetUint64(chainID))
	rpcclient, err := rpc.Dial(url)
	if err != nil {
		return nil, err
	}
	client := ethclient.NewClient(rpcclient)

	nonce, err := client.NonceAt(context.Background(), addr, nil)
	if err != nil {
		return nil, err
	}

	return &Account{
		pk:        pk,
		address:   addr,
		nonce:     nonce,
		url:       url,
		signer:    signer,
		client:    client,
		rpcClient: rpcclient,
	}, nil
}

func (c *Account) Address() common.Address {
	return c.address
}

func (c *Account) Balance(blockNum *big.Int) (*big.Int, error) {
	return c.client.BalanceAt(context.Background(), c.address, blockNum)
}

func (c *Account) BalanceOf(addr common.Address, blockNum *big.Int) (*big.Int, error) {
	return c.client.BalanceAt(context.Background(), addr, blockNum)
}

func (c *Account) TransferWithConfirm(to common.Address, amount *big.Int) (common.Hash, error) {
	hash, err := c.Transfer(to, amount)
	if err != nil {
		return common.EmptyHash, err
	}
	if err := c.waitTransaction(hash); err != nil {
		return common.EmptyHash, err
	}
	return hash, nil
}

func (c *Account) Transfer(to common.Address, amount *big.Int) (common.Hash, error) {
	signedTx, err := c.newSignedTx(to, amount, nil)
	if err != nil {
		return common.EmptyHash, err
	}
	if err := c.SendTx(signedTx); err != nil {
		return common.EmptyHash, err
	} else {
		return signedTx.Hash(), nil
	}
}

func (c *Account) Deploy(startTime uint64) (common.Address, error) {
	auth := c.makeDeployAuth()
	addr, tx, _, err := stat.DeployStat(auth, c.backend(), startTime)
	if err != nil {
		return common.EmptyAddress, err
	}
	if err := c.waitTransaction(tx.Hash()); err != nil {
		return common.EmptyAddress, err
	}
	return addr, nil
}

func (c *Account) Add(contract common.Address) (common.Hash, error) {
	auth := c.makeAuth()
	st, err := stat.NewStat(contract, c.backend())
	if err != nil {
		return common.EmptyHash, err
	}
	if tx, err := st.Add(auth); err != nil {
		return common.EmptyHash, err
	} else {
		return tx.Hash(), nil
	}
}

func (c *Account) TxNum(contract common.Address) (uint64, error) {
	st, err := stat.NewStat(contract, c.backend())
	if err != nil {
		return 0, err
	}
	if num, err := st.TxNum(nil); err != nil {
		return 0, err
	} else {
		return num.Uint64(), nil
	}
}

func (c *Account) Nonce() uint64 {
	defer func() {
		c.nonce += 1
	}()

	return c.nonce
}

func (c *Account) SendTx(signedTx *types.Transaction) error {
	return c.client.SendTransaction(context.Background(), signedTx)
}

func (c *Account) backend() *ethclient.Client {
	return c.client
}

func (c *Account) makeDeployAuth() *bind.TransactOpts {
	auth := bind.NewKeyedTransactor(c.pk)
	auth.GasLimit = 1e7
	auth.Nonce = new(big.Int).SetUint64(c.Nonce())
	return auth
}

func (c *Account) makeAuth() *bind.TransactOpts {
	auth := bind.NewKeyedTransactor(c.pk)
	auth.GasLimit = 50000
	auth.Nonce = new(big.Int).SetUint64(c.Nonce())
	return auth
}

func (c *Account) signAndSendTx(payload []byte, contract common.Address) (common.Hash, error) {
	return c.signAndSendTxWithValue(payload, big.NewInt(0), contract)
}

func (c *Account) signAndSendTxWithValue(payload []byte, amount *big.Int, contract common.Address) (common.Hash, error) {
	hash := common.EmptyHash
	tx, err := c.newSignedTx(contract, amount, payload)
	if tx != nil {
		hash = tx.Hash()
	}
	if err != nil {
		return hash, fmt.Errorf("sign tx failed, err: %v", err)
	}

	if err := c.SendTx(tx); err != nil {
		return hash, err
	}
	if err := c.waitTransaction(tx.Hash()); err != nil {
		return hash, err
	}
	return hash, nil
}

func (c *Account) newSignedTx(to common.Address, amount *big.Int, data []byte) (*types.Transaction, error) {
	unsignedTx, err := c.newUnsignedTx(to, amount, data)
	if err != nil {
		return nil, err
	}
	return types.SignTx(unsignedTx, c.signer, c.pk)
}

func (c *Account) newUnsignedTx(to common.Address, amount *big.Int, data []byte) (*types.Transaction, error) {
	nonce := c.Nonce()
	gasPrice, err := c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	callMsg := ethereum.CallMsg{
		From:     c.Address(),
		To:       &to,
		Gas:      0,
		GasPrice: gasPrice,
		Value:    amount,
		Data:     data,
	}
	gasLimit, err := c.client.EstimateGas(context.Background(), callMsg)
	if err != nil {
		return nil, fmt.Errorf("estimate gas limit error: %s", err.Error())
	}

	return types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &to,
		Value:    amount,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     data,
	}), nil
}

func (c *Account) waitTransaction(hash common.Hash) error {
	for {
		time.Sleep(time.Second * 1)
		_, ispending, err := c.client.TransactionByHash(context.Background(), hash)
		if err != nil {
			log.Errorf("failed to call TransactionByHash: %v", err)
			continue
		}
		if ispending == true {
			continue
		}

		if err := c.dumpEventLog(hash); err != nil {
			return err
		}
		break
	}
	return nil
}

func (c *Account) dumpEventLog(hash common.Hash) error {
	raw, err := c.getReceipt(hash)
	if err != nil {
		return fmt.Errorf("faild to get receipt %s", hash.Hex())
	}

	if raw.Status == 0 {
		return fmt.Errorf("receipt failed %s", hash.Hex())
	}

	log.Infof("txhash %s, block height %d", hash.Hex(), raw.BlockNumber.Uint64())
	for _, event := range raw.Logs {
		log.Infof("eventlog address %s", event.Address.Hex())
		log.Infof("eventlog data %s", hexutil.Encode(event.Data))
		for i, topic := range event.Topics {
			log.Infof("eventlog topic[%d] %s", i, topic.String())
		}
	}
	return nil
}

func (c *Account) getReceipt(hash common.Hash) (*types.Receipt, error) {
	raw := &types.Receipt{}
	if err := c.rpcClient.Call(raw, "eth_getTransactionReceipt", hash.Hex()); err != nil {
		return nil, err
	}
	return raw, nil
}
