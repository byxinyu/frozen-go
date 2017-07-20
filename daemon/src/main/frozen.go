package main

import (
	"encoding/json"
	"io/ioutil"
	"fmt"
	"conf"
	"os"
	"time"
	"io"
	"net"
	"strconv"
	"os/exec"
	"errors"
	"path/filepath"
	"crypto/sha256"
)

const VERSION string = "v0.0 Alpha 0x00d"
const FILE_CONFIGURATION string = "../conf/fg.json"

var serverSaved []ServerLocal
var config conf.Config
var servers []ServerRun

// SSC: ServerLocal Self Checking
// Status CODE
// 00
// 二进制表示
const SSC_NO_CONFIG_FILE int = -1
const SSC_NO_SERVER_DIR = -2

// 服务器状态码

// 已经关闭
const SERVER_STATUS_CLOSED = 0

// 正在运行
const SERVER_STATUS_RUNNING = 1

type ServerLocal struct {
	ID         int
	Name       string
	Executable string
	Status     int
}

type ExecConf struct {
	Command string
	Args    []string
}

type ServerRun struct {
	ID     int
	Cmd    *exec.Cmd
	Stdin  io.WriteCloser
	Stdout io.ReadCloser
}

type Request struct {
	Method    string
	OperateID int
	Message   string
}

type Response struct {
	Status  int
	Message string
}

type InterfaceRequest struct {
	Auth string
	Req  Request
}

/*
Command : List / Start / getStatus /
 */
func main() {
	if !(len(os.Args) > 1 && os.Args[1] == "-jump") {
		printInfo()
	}
	config, _ = conf.GetConfig(FILE_CONFIGURATION)
	b, _ := ioutil.ReadFile(config.Smc.Servers)
	json.Unmarshal(b, &serverSaved)
	fmt.Println("Started Server Manager.")
	fmt.Println("Online...")
	handleRequest(Request{"Start",0,""})
	go StartDaemonServer()
	fmt.Println("Done,type \"?\" for help. ")
	for {
		var s string
		fmt.Scanf("%s", &s)
		processLocalCommand(s)
	}
}

// 命令处理器
func handleRequest(request Request) Response {
	switch request.Method {

	case "List":
		return outputListOfServers()
	case "Create":
		serverSaved = append(serverSaved, ServerLocal{len(serverSaved), request.Message, "", 0})
		serverSaved[len(serverSaved)-1].EnvPrepare()
		// 序列化b来储存。
		b, err := json.MarshalIndent(serverSaved, "", "\t")

		// 新创建的服务器写入data文件
		err2 := ioutil.WriteFile(config.Smc.Servers, b, 0666)
		if err2 != nil {
			return Response{
				-1,
				err2.Error(),
			}
		}
		if err != nil {
			return Response{
				-1,
				err.Error(),
			}
		}
		return Response{
			0,
			"OK",
		}
	case "Start":
		// 运行这个服务器
		if request.OperateID > len(serverSaved)-1 {
			return Response{
				-1, "Invalid server id",
			}
		}
		err := serverSaved[request.OperateID].Start()
		if err == nil {
			return Response{
				0, "OK",
			}
		} else {
			return Response{-1, err.Error()}
		}

	case "Stop":
		if request.OperateID > len(servers)-1 {
			return Response{0, "Invalid serverid"}
		}
		servers[request.OperateID].Close()

		return Response{
			0, "OK",
		}

	case "SetExecutable":
		serverSaved[request.OperateID].Executable = request.Message
		return Response{
			0, "OK",
		}
	}
	return Response{
		-1, "Unexpected err",
	}
}

func outputListOfServers() Response {
	b, _ := json.Marshal(serverSaved)
	return Response{0, string(b)}
}

func printInfo() {
	fmt.Println("  _____                        ____")
	fmt.Println("|  ___| __ ___ _______ _ __  / ___| ___")
	fmt.Println("| |_ | '__/ _\\_  / _ \\ '_ \\| |  _ / _ \\")
	fmt.Println("|  _|| | | (_) / /  __/ | | | |_| | (_) |")
	fmt.Println("|_|  |_|\\___/___\\___|_| |_|\\____|\\___/")
	time.Sleep(2 * time.Second)
	fmt.Println("---------------------")
	time.Sleep(100 * time.Microsecond)
	fmt.Print("Powered by ")
	for _, v := range []byte("Axoford12") {
		time.Sleep(240 * time.Millisecond)
		fmt.Print(string(v))
	}
	fmt.Println()
	time.Sleep(1000 * time.Millisecond)
	time.Sleep(100 * time.Microsecond)
	fmt.Println("---------------------")
	time.Sleep(300 * time.Millisecond)
	fmt.Println("version:" + VERSION)
	time.Sleep(1 * time.Second)
}

func StartDaemonServer() {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(config.Dsc.Port)) // 默认使用tcp连接
	if err != nil {
		panic(err)
	} else {
		for {
			conn, err := ln.Accept()
			fmt.Println("New Client connected.")
			if err != nil {
				continue
			}
			go handleConnection(conn)
		}
	}

}

func auth(src InterfaceRequest) bool {
	dst := sha256.Sum256([]byte(config.Dsc.VerifyCode))
	auth := sha256.Sum256([]byte(src.Auth))
	return dst == auth
}

func handleConnection(c net.Conn) {
	buf := make([]byte, config.DefaultBufLength)
	length, _ := c.Read(buf)
	request := InterfaceRequest{}
	err := json.Unmarshal(buf[:length], &request)
	if err != nil {
		res, _ := json.Marshal(Response{-1, err.Error(), })
		c.Write(res)
		c.Close()
	}
	if auth(request) {
		fmt.Fprintln(c, "Auth succeeded!")
		if request.Req.Method == "GetInput" {
			// 判断输入的有效性
			if request.Req.OperateID > len(serverSaved)-1 {
				res, _ := json.Marshal(Response{-1, "Invalid argument : OperateID", })
				c.Write(res)
				c.Close()
			} else if serverSaved[request.Req.OperateID].Status != 1 {
				res, _ := json.Marshal(Response{-1, "Server not started", })
				c.Write(res)
				c.Close()
			} else {
				for {
					io.Copy(servers[request.Req.OperateID].Stdin,c)
				}
			}
		} else if request.Req.Method == "GetOutput" {

			if request.Req.OperateID > len(serverSaved)-1 {
				res, _ := json.Marshal(Response{-1, "Invalid argument : OperateID", })
				c.Write(res)
				c.Close()
			} else if serverSaved[request.Req.OperateID].Status != 1 {
				res, _ := json.Marshal(Response{-1, "Server not started", })
				c.Write(res)
				c.Close()
			} else {
				for {
					io.Copy(c, servers[request.Req.OperateID].Stdout)
				}
			}

		} else {
			response := handleRequest(request.Req)
			res, _ := json.Marshal(response)
			c.Write(res)
			c.Close()
		}

	} else {
		res, _ := json.Marshal(Response{-1, "Auth failed", })
		c.Write(res)
		c.Close()
		return
	}

}

func (server *ServerLocal) Start() error {
	if server.Status == 1 {
		return errors.New("Server already started.")
	}
	server.EnvPrepare()
	serverRC, err := server.loadExecutableConfig()
	if err != nil {
		// 环境准备失败
		return errors.New("Cannot prepare server env")
	} else {
		// 如果Command就是一个绝对路径，不再寻找。
		execPath := serverRC.Command
		if !filepath.IsAbs(serverRC.Command) {
			var isNoFound error
			execPath, isNoFound = exec.LookPath(serverRC.Command)
			if isNoFound != nil {
				return isNoFound // 没找到抛err
			}
		}
		// 根据提供的EXEC名，搜寻绝对目录

		nowPath, err := filepath.Abs(".")
		if err != nil {
			return errors.New(err.Error())
		}
		// 取得服务器目录
		serverRunPath := filepath.Clean(nowPath + "/../servers/server" + strconv.Itoa(server.ID))
		cmd := exec.Command(execPath, serverRC.Args...)
		cmd.Dir = serverRunPath
		stdout, err := cmd.StdoutPipe()
		stdin, err2 := cmd.StdinPipe()
		if err2 != nil {
			panic(err2)
		}
		if err != nil {
			panic(err)
		}

		err3 := cmd.Start()
		if err3 != nil {
			panic(err3)
		}
		newRunningServer := ServerRun{
			ID:     server.ID,
			Cmd:    cmd,
			Stdout: stdout,
			Stdin:  stdin,
		}
		server.Status = SERVER_STATUS_RUNNING
		servers = append(servers, newRunningServer)
		return nil
	}
}

// 检查运行环境
func (server *ServerLocal) selfChecking() int {
	var status int = 0
	_, err := os.Stat("../exec/" + server.Executable + ".json")
	_, err2 := os.Stat("../servers/server" + strconv.Itoa(server.ID))
	if err != nil {

		status += SSC_NO_CONFIG_FILE

	}
	if err2 != nil {
		status += SSC_NO_SERVER_DIR
	}
	return status
}

// 按照错误码修复环境
func (server *ServerLocal) EnvPrepare() bool {
	statusCode := server.selfChecking()
	switch statusCode {
	case SSC_NO_SERVER_DIR:
		err := os.MkdirAll("../servers/server"+strconv.Itoa(server.ID), 0666)
		return err == nil
	case SSC_NO_CONFIG_FILE:
		defaultExec := ExecConf{"java", []string{"-jar", "Minecraft.jar"}}
		//defaultExec := ExecConf{"ping",nil}
		file, err := os.Create("../exec/Minecraft.json")
		defer file.Close()
		b, err2 := json.MarshalIndent(defaultExec, "", "\t")
		io.WriteString(file, string(b)) // 写入文件
		return err == nil && err2 == nil
	case SSC_NO_CONFIG_FILE + SSC_NO_SERVER_DIR:
		// 两路一起执行
		err3 := os.MkdirAll("../servers/server"+strconv.Itoa(server.ID), 0666)
		defaultExec := ExecConf{"java", []string{"-jar", "Minecraft.jar"}}
		//defaultExec := ExecConf{"ping",nil}
		file, err := os.Create("../exec/Minecraft.json")
		defer file.Close()
		b, err2 := json.MarshalIndent(defaultExec, "", "\t")
		io.WriteString(file, string(b)) // 写入文件
		return err == nil && err2 == nil && err3 == nil
	}
	return false
}
func (server *ServerLocal) loadExecutableConfig() (ExecConf, error) {
	var newServerRuntimeConf ExecConf
	b, err := ioutil.ReadFile("../exec/" + server.Executable + ".json") // 将配置文件读入
	if err != nil {
		// 若在读文件时就有异常则停止反序列化
		return newServerRuntimeConf, err
	}
	err2 := json.Unmarshal(b, &newServerRuntimeConf) //使用自带的json库对读入的东西反序列化
	if err2 != nil {
		return newServerRuntimeConf, err
	}
	return newServerRuntimeConf, nil // 返回结果
}

func (s *ServerRun) Close() {
	s.Cmd.Process.Release()
	s.Cmd.Process.Kill()
	serverSaved[s.ID].Status = SERVER_STATUS_CLOSED
}

// 保存服务器信息
func saveServerInfo() error {

	for i := 0; i < len(serverSaved); i++ {
		serverSaved[i].Status = 0
	}
	b, err := json.Marshal(serverSaved)
	if err != nil {
		return err
	}
	ioutil.WriteFile(config.Smc.Servers, b, 0664)
	return nil
}

// 处理本地命令
func processLocalCommand(c string) {
	switch c {
	case "stop":

		fmt.Println("Stopping")
		for i:=0;i<len(serverSaved);i++{
			if serverSaved[i].Status == 1 {
				servers[i].Cmd.Process.Kill()
			}
		}
		saveServerInfo()
		os.Exit(0)
	case "?":
		fmt.Println("FrozenGo" + VERSION + " Help Manual -- by Axoford12")
		fmt.Println("stop: Stop the daemon.save server changes.")
		fmt.Println("status: Echo server status.")
		return
	case "status":
		spaceH := "|--"
		switch len(serverSaved) {
		case 0:
			fmt.Println(spaceH + "There is no server.")
		case 1:
			fmt.Println(spaceH + "There is 1 server")
		default:
			fmt.Println(spaceH + "There are " + strconv.Itoa(len(serverSaved)) + " servers")
		}
		for i := 0; i < len(serverSaved); i ++ {
			fmt.Println(spaceH + spaceH + "ID:" + strconv.Itoa(i))
			fmt.Println(spaceH + spaceH + serverSaved[i].Name)
			var status string
			switch serverSaved[i].Status {
			case 0:
				status = "Stopped"
			case 1:
				status = "Running"
			}
			fmt.Println(spaceH + spaceH + "Status:" + status)

		}
		return

	}
}
