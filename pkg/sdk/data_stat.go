/*
 * Copyright (C) 2021 The Zion Authors
 * This file is part of The Zion library.
 *
 * The Zion is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The Zion is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The Zion.  If not, see <http://www.gnu.org/licenses/>.
 */

package sdk

import (
	"fmt"
	"math/big"
	"strings"

	stat "github.com/dylenfu/zion-meter/pkg/go_abi/data_stat_abi"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/contracts/native/utils"
)

var (
	dataStatABI *abi.ABI
)

func init() {
	ab, err := abi.JSON(strings.NewReader(stat.DataStatABI))
	if err != nil {
		panic(fmt.Sprintf("failed to load abi json string: [%v]", err))
	}
	dataStatABI = &ab
}

func (c *Account) DeployDataStat() (common.Address, error) {
	if c.sender == nil {
		return common.EmptyAddress, ErrNoSender
	}

	auth := c.makeDeployAuth()
	addr, tx, _, err := stat.DeployDataStat(auth, c.sender.client)
	if err != nil {
		return common.EmptyAddress, err
	}
	if err := c.waitTransaction(tx.Hash()); err != nil {
		return common.EmptyAddress, err
	}
	return addr, nil
}

func (c *Account) ResetDataStat(contract common.Address, startTime uint64) (common.Hash, error) {
	if c.sender == nil {
		return common.EmptyHash, ErrNoSender
	}
	st, err := stat.NewDataStat(contract, c.sender.client)
	if err != nil {
		return common.EmptyHash, err
	}
	auth := c.makeAuth()
	auth.GasLimit = 500000
	tx, err := st.Reset(auth, startTime)
	if err != nil {
		return common.EmptyHash, err
	}
	if err := c.waitTransaction(tx.Hash()); err != nil {
		return common.EmptyHash, err
	}
	return tx.Hash(), nil
}

func (c *Account) CostManyGas(contract common.Address, callData []byte, complexity uint64) (common.Hash, uint64, error) {
	if c.sender == nil {
		return common.EmptyHash, c.nonce, ErrNoSender
	}

	payload, err := utils.PackMethod(dataStatABI, "costManyGas", callData, complexity)
	if err != nil {
		return common.EmptyHash, c.nonce, err
	}

	originNonce := c.nonce

	tx, err := c.newSignedTx(contract, big.NewInt(0), payload)
	if err != nil {
		return common.EmptyHash, originNonce, err
	}

	if err := c.SendTx(tx); err != nil {
		return common.EmptyHash, originNonce, err
	}

	return tx.Hash(), originNonce, nil
}

func (c *Account) DataStatTxNum(contract common.Address) (uint64, error) {
	if c.sender == nil {
		return 0, ErrNoSender
	}

	st, err := stat.NewDataStat(contract, c.sender.client)
	if err != nil {
		return 0, err
	}
	if num, err := st.TxNum(nil); err != nil {
		return 0, err
	} else {
		return num.Uint64(), nil
	}
}

func (c *Account) DataStatExist(contract common.Address) bool {
	if c.sender == nil {
		return false
	}

	st, err := stat.NewDataStat(contract, c.sender.client)
	if err != nil {
		return false
	}

	if startTime, _ := st.StartTime(nil); startTime > 0 {
		return true
	} else {
		return false
	}
}
