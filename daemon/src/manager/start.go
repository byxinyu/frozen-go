/*
Author Axoford12
Team Freeze
Org Rubiginosu
  _____                        ____
|  ___| __ ___ _______ _ __  / ___| ___
| |_ | '__/ _ \_  / _ \ '_ \| |  _ / _ \
|  _|| | | (_) / /  __/ | | | |_| | (_) |
|_|  |_|  \___/___\___|_| |_|\____|\___/

 */

// 对服务器进行管理的一个程序模块
package manager

type Server struct {
	Name  string
	Owner string
	Path  string
	ID    int
}
type ServerManager struct {
	servers map[string]Server
}

func Start() {

}
