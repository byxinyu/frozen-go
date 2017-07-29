package dmserver

import (
	"auth"
	"conf"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
)

var config conf.Config
var serverSaved []ServerLocal
var servers []ServerRun

func connErrorToExit(errorInfo string, c net.Conn) {
	res, _ := json.Marshal(Response{-1, errorInfo})
	c.Write(res)
	c.Close()
}

// 保存服务器信息
func saveServerInfo() error {
	b, err := json.Marshal(serverSaved)
	if err != nil {
		return err
	}
	ioutil.WriteFile(config.ServerManager.Servers, b, 0664)
	return nil
}

// 处理本地命令

func handleConnection(c net.Conn) {
	buf := make([]byte, config.DaemonServer.DefaultBufLength)
	length, _ := c.Read(buf)
	request := InterfaceRequest{}
	err := json.Unmarshal(buf[:length], &request)
	if err != nil {
		connErrorToExit(err.Error(), c)
	}
	if request.Req.Method == "GetInput" {
		if ioCheck(request, c) {
			for {
				io.Copy(servers[searchRunningServerByID(request.Req.OperateID)].Stdin, c)
			}
		}

	} else if request.Req.Method == "GetOutput" {
		if ioCheck(request, c) {
			for {
				io.Copy(c, servers[searchRunningServerByID(request.Req.OperateID)].Stdout)
			}
		}
	} else if request.Auth == config.DaemonServer.VerifyCode {
		res, _ := json.Marshal(handleRequest(request.Req))
		c.Write(res)
		c.Close()
	} else {
		connErrorToExit("No command.", c)
	}

}

// 命令处理器
func handleRequest(request Request) Response {
	pairs := auth.GetValidationKeyPairs()
	index := searchServerByID(request.OperateID)
	switch request.Method {

	case "List":
		return outputListOfServers()
	case "Create":
		serverId := 0
		if len(serverSaved) != 0 {
			serverId = serverSaved[len(serverSaved)-1].ID + 1
		}

		serverSaved = append(serverSaved, ServerLocal{serverId, request.Message, "", 0, 0})
		serverSaved[len(serverSaved)-1].EnvPrepare()
		// 序列化b来储存。
		b, err := json.MarshalIndent(serverSaved, "", "\t")

		// 新创建的服务器写入data文件
		err2 := ioutil.WriteFile(config.ServerManager.Servers, b, 0666)
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
	case "Delete":
		if !IsServerAvaible(request.OperateID) {
			return Response{-1, "Invalid server id"}
		}
		serverSaved[searchServerByID(request.OperateID)].Delete()
		return Response{0, "OK"}
	case "Start":

		// 运行这个服务器
		if index < 0 {
			return Response{
				-1, "Invalid server id",
			}
		}
		err := serverSaved[index].Start()
		if err == nil {
			return Response{
				0, "OK",
			}
		} else {
			return Response{-1, err.Error()}
		}

	case "Stop":
		if index := searchRunningServerByID(request.OperateID); index < 0 {
			return Response{0, "Invalid serverid"}
		} else {
			servers[index].Close()
		}


		return Response{
			0, "OK",
		}

	case "SetExecutable":
		serverSaved[index].Executable = request.Message
		return Response{
			0, "OK",
		}

	case "GetPairs":
		if !IsServerAvaible(request.OperateID) {
			return Response{
				-1, "Invalid server id",
			}
		}
		for i := 0; i < len(pairs); i++ {
			if pairs[i].ValidationKeyPair.ID == request.OperateID {
				responseData, _ := json.Marshal(pairs[i])
				return Response{
					0, string(responseData),
				}
			}
		}
		// 未找到已经存在的ValidationKey
		// 为请求者生成ValidationKey
		pair := auth.ValidationKeyGenerate(request.OperateID)
		responseData, _ := json.Marshal(pair)
		auth.ValidationKeyPairs = append(auth.ValidationKeyPairs, pair)
		return Response{0, string(responseData)}
	}
	return Response{
		-1, "Unexpected err",
	}
}
