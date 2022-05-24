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
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

var (
	testMaster      *Account
	testAmount      = big.NewInt(1000000000000000) //0.001 eth
	testChainID     = uint64(60801)
	testUrl         = "http://127.0.0.1:22000"
	testMainNodeKey = "4b0c9b9d685db17ac9f295cb12f9d7d2369f5bf524b3ce52ce424031cafda1ae"
)

func TestMain(m *testing.M) {
	sender, _ := NewSender(testUrl, testChainID)
	testMaster, _ = MasterAccount(sender, testMainNodeKey)
	os.Exit(m.Run())
}

// go test -v github.com/dylenfu/zion-meter/pkg/sdk -run TestTransfer
func TestTransfer(t *testing.T) {
	n := 100
	to := common.HexToAddress("0x67CDE763bD045B14898d8B044F8afC8695ae8608")

	balance, _ := testMaster.BalanceOf(to, nil)
	t.Logf("balance before transfer %s, nonce before transfer %d", balance.String(), testMaster.nonce)

	for i := 0; i < n; i++ {
		if _, err := testMaster.Transfer(to, testAmount); err != nil {
			t.Fatal(err)
		}
	}

	time.Sleep(15 * time.Second)
	balance, _ = testMaster.BalanceOf(to, nil)
	t.Logf("balance after transfer %s, nonce after transfer %d", balance.String(), testMaster.nonce)
}

// go test -v github.com/dylenfu/zion-meter/pkg/sdk -run TestStat
func TestStat(t *testing.T) {
	n := 10000
	startTime := uint64(time.Now().Unix())

	acc, err := NewAccount()
	if err != nil {
		t.Fatal(err)
	}

	balance, _ := new(big.Int).SetString("1000000000000000000", 10)
	if _, err := testMaster.TransferWithConfirm(acc.Address(), balance); err != nil {
		t.Fatal(err)
	}

	contract, err := testMaster.Deploy()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("contract address %s, start time %d, nonce before testing %d", contract.Hex(), startTime, acc.nonce)

	sender, err := NewSender(testUrl, testChainID)
	acc.SetSender(sender)
	for i := 0; i < n; i++ {
		if _, _, err := acc.Add(contract); err != nil {
			t.Error(err)
		}
	}

	time.Sleep(15 * time.Second)

	total, err := testMaster.TxNum(contract)
	if err != nil {
		t.Fatal(err)
	}
	endTime := uint64(time.Now().Unix())
	t.Logf("end time %d, spent %d, nonce after testing %d, total tx number %d", endTime, endTime-startTime, acc.nonce, total)
}
