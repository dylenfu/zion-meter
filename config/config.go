/*
 * Copyright (C) 2020 The poly network Authors
 * This file is part of The poly network library.
 *
 * The  poly network  is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The  poly network  is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 * You should have received a copy of the GNU Lesser General Public License
 * along with The poly network .  If not, see <http://www.gnu.org/licenses/>.
 */

package config

import (
	"encoding/json"

	"github.com/dylenfu/zion-meter/pkg/files"
)

const (
	MinTPS   = 20
	MinGroup = 2
)

var (
	Conf *Config
)

type Config struct {
	Workspace    string
	ChainID      uint64
	Groups       int    // 账户分组，总组数为`Groups`，每秒切换不同组用户发送交易
	AccsPerGroup int    // 每组账户数量
	Sharding     bool   // 是否需要多台机器测试
	FirstMachine string // 第一台机器内网地址，只需第一台机器统计tps，其他的不需要
	LastTime     string
	Contract     string
	NodeKey      string
	Nodes        []string
}

func LoadConfig(filepath string, group, account int, sLastTime string) {
	if account < MinTPS || group < MinGroup {
		panic("user per group should > 20 and group number should > 2")
	}

	data, err := files.ReadFile(filepath)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(data, &Conf); err != nil {
		panic(err)
	}

	if group > 0 {
		Conf.Groups = group
	}
	if account > 0 {
		Conf.AccsPerGroup = account
	}
	if sLastTime != "" {
		Conf.LastTime = sLastTime
	}
}
