package main

import (
	"encoding/json"
	"fmt"
	"os"
	"server/network"
)

/*
<스터디장님 팁>
localhost :: UINX/DOMAIN 소속, TCP/IP 프로토콜 안 탐. (etc/mysql.sock)
127.0.0.1 :: UNIX/DOMAIN 소속X, TCP/IP 프로토콜 타서 되돌아옴.
0.0.0.0  ::  외부의 모든 주소로 받겠다.
*/
func main() {
	netConfig, err := parseNetConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	startLifeGameServer(netConfig) // [PATH]: [./game.go]
}

func parseNetConfig() (network.NetConfig, error) {
	var netConfig network.NetConfig
	/*
		[PATH]: [./network/network.go]
		type NetConfig struct {
			BindAdress string `json: "bind_address"`
			Port       int    `json: "port"`
		}
	*/
	file, err := os.Open("./config/net_config.json")
	if err != nil {
		return netConfig, err
	}

	defer file.Close()

	jsonParser := json.NewDecoder(file)
	jsonParser.Decode(&netConfig)

	return netConfig, err
}
