package common

import "testing"
import "log"
import "fmt"
import "encoding/json"
import "net"

//import "github.com/stretchr/testify/assert"

func TestParam(t *testing.T) {
	param := TaskParam{
		Type: IsAliveTask,
		Data: TaskParam{},
	}

	if err := param.Validate(); err == nil {
		t.Error("Expected invalid param data")
	}

	param.Type = 50
	if err := param.Validate(); err == nil {
		t.Error("Expected invalid param type")
	}
	param.Type = IsAliveTask
	param.Data = IsAliveParam{}
	if err := param.Validate(); err == nil {
		t.Error("Expected invalid param data")
	}
}

func TestParamSerialization(t *testing.T) {
	param := TaskParam{
		Type: IsAliveTask,
		Data: IsAliveParam{IpRange: IpRange{Start: net.ParseIP("192.168.1.1"), End: net.ParseIP("192.168.1.10")}},
	}

	var isAliveResult []IpResult
	isAliveResult = append(isAliveResult, IpResult{Ip: net.ParseIP("192.168.2.1"), Status: IpAlive})
	isAliveResult = append(isAliveResult, IpResult{Ip: net.ParseIP("192.168.2.1"), Status: IpDead})
	result := CompleteTaskArgs{TaskId: 1, Result: isAliveResult}
	b, err := json.MarshalIndent(param, "", "   ")
	if err != nil {
		log.Println("error:", err)
	}
	fmt.Printf("IsAliveParam: %s\n", b)
	b, err = json.MarshalIndent(result, "", "   ")
	if err != nil {
		log.Println("error:", err)
	}
	fmt.Printf("IsAliveResult: %s\n", b)

	param = TaskParam{
		Type: PortScanTask,
		Data: PortScanParam{Type: SynScan, Ip: net.ParseIP("192.168.2.1"), PortRange: PortRange{100, 180}},
	}
	b, err = json.MarshalIndent(param, "", "   ")
	if err != nil {
		log.Println("error:", err)
	}
	fmt.Printf("PortScanParam: %s\n", b)
}
