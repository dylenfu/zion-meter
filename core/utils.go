package core

import (
	"net"
	"strings"

	"github.com/dylenfu/zion-meter/config"
	"github.com/dylenfu/zion-meter/pkg/sdk"
	"github.com/ethereum/go-ethereum/common"
)

func masterAccount() (*sdk.Account, error) {
	chainID := config.Conf.ChainID
	node := config.Conf.Nodes[0]
	sender, err := sdk.NewSender(node, chainID)
	if err != nil {
		return nil, err
	}
	return sdk.MasterAccount(sender, config.Conf.NodeKey)
}

func singleAccount() (*sdk.Account, error) {
	chainID := config.Conf.ChainID
	node := config.Conf.Nodes[0]
	sender, err := sdk.NewSender(node, chainID)
	if err != nil {
		return nil, err
	}
	account, err := sdk.NewAccount()
	if err != nil {
		return nil, err
	}
	account.SetSender(sender)
	return account, nil
}

// 每组N个账户，总共G组
func generateAccounts() (int, [][]*User, error) {
	chainID := config.Conf.ChainID
	connNum := config.Conf.AccsPerGroup
	nodeNum := len(config.Conf.Nodes)
	senderList := make([]*sdk.Sender, 0)
	for i := 0; i < connNum; i++ {
		node := config.Conf.Nodes[i%nodeNum]
		sender, err := sdk.NewSender(node, chainID)
		if err != nil {
			return 0, nil, err
		}
		senderList = append(senderList, sender)
	}

	groupNo := config.Conf.Groups
	AccNoPerGroup := config.Conf.AccsPerGroup
	total := groupNo * AccNoPerGroup
	list := make([][]*User, 0)
	for i := 0; i < groupNo; i++ {
		group := make([]*User, 0)
		for j := 0; j < AccNoPerGroup; j++ {
			acc, err := sdk.NewAccount()
			if err != nil {
				return 0, nil, err
			}
			index := (i*AccNoPerGroup + j) % connNum
			acc.SetSender(senderList[index])
			user := newUser(acc)
			group = append(group, user)
		}
		list = append(list, group)
	}

	return total, list, nil
}

func resetOrDeployContract(acc *sdk.Account, oldContract common.Address, startTime uint64) (common.Address, error) {
	if _, err := acc.ResetDataStat(oldContract, startTime); err != nil {
		return common.EmptyAddress, err
	} else {
		return oldContract, nil
	}

	if acc.DataStatExist(oldContract) {
		if _, err := acc.ResetDataStat(oldContract, startTime); err != nil {
			return common.EmptyAddress, err
		}
		return oldContract, nil
	}

	newContract, err := acc.DeployDataStat()
	if err != nil {
		return common.EmptyAddress, err
	}
	if _, err := acc.ResetDataStat(newContract, startTime); err != nil {
		return newContract, err
	}
	return newContract, nil
}

func InternalIP() string {
	inters, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, inter := range inters {
		if inter.Flags&net.FlagUp != 0 && !strings.HasPrefix(inter.Name, "lo") {
			addrs, err := inter.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						return ipnet.IP.String()
					}
				}
			}
		}
	}

	return ""
}

// 如果不需要多台机器，则直接返回true，本机需要统计tps;
// 如果需要多台机器，则判断本机是否为第一台机器地址，是则返回true
func IsLeader() bool {
	if !config.Conf.Sharding {
		return true
	}

	local := InternalIP()
	if local == config.Conf.FirstMachine {
		return true
	}

	return false
}
