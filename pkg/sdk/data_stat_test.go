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
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// go test -v -count=1 github.com/dylenfu/zion-meter/pkg/sdk -run TestDataStat
func TestDataStat(t *testing.T) {

	// prepare test account
	acc, err := NewAccount()
	if err != nil {
		t.Fatal(err)
	}
	sender, err := NewSender(testUrl, testChainID)
	if err != nil {
		t.Fatal(err)
	}
	acc.SetSender(sender)

	// set balance for test account
	t.Log("start to prepare account balance...")
	balance, _ := new(big.Int).SetString("1000000000000000000", 10)
	if _, err := testMaster.TransferWithConfirm(acc.Address(), balance); err != nil {
		t.Fatal(err)
	}
	t.Log("prepare account balance done!")

	// deploy contract
	t.Log("start to deploy contract...")
	contract, err := testMaster.DeployDataStat()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("contract address %s, nonce before testing %d", contract.Hex(), acc.nonce)

	// test send complex tx
	t.Log("start to test send `costManyGas`")
	n := 1000
	startTime := uint64(time.Now().Unix())
	input, _ := hex.DecodeString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	complexity := uint64(8)
	var txHash common.Hash
	for i := 0; i < n; i++ {
		if txHash, _, err = acc.CostManyGas(contract, input, complexity); err != nil {
			t.Error(err)
		}
	}
	t.Logf("send tx done and waiting for block, the last tx hash %s!", txHash.Hex())

	acc.waitTransaction(txHash)

	total, err := testMaster.TxNum(contract)
	if err != nil {
		t.Fatal(err)
	}
	endTime := uint64(time.Now().Unix())
	t.Logf("end time %d, spent %d, nonce after testing %d, total tx number %d", endTime, endTime-startTime, acc.nonce, total)
}
