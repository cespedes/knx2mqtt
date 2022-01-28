package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/vapourismo/knx-go/knx/cemi"
)

/* Syntax of KNXweb config file:

logdir /var/log/knx
port 8001
mqtt-server 127.0.0.1
mqtt-prefix1 control/knx
mqtt-prefix2 control/rooms
gateway 192.168.1.11 1/ 2/5/
	...
device 1.1.10 myroom.thermostat
	...
address 2/5/7 9.001 myroom/temperature
	...
*/
type addrNameType struct {
	Name string
	DPT  string
}

type Gateway struct {
	Address string
	Groups  []string
}

type Config struct {
	MQTTServer  string
	MQTTPrefix1 string
	MQTTPrefix2 string
	Logdir      string                          // Where to store packet logs
	Port        int                             // TCP port to listen HTTP requests
	Devices     map[cemi.IndividualAddr]string  // List of KNX devices
	Addresses   map[cemi.GroupAddr]addrNameType // List of KNX group addresses
}

type UnknownDPT []byte

func (d UnknownDPT) Pack() []byte {
	return []byte(d)
}

func (d *UnknownDPT) Unpack(data []byte) error {
	tmp := make([]byte, len(data))
	copy(tmp, data)
	*d = UnknownDPT(tmp)
	return nil
}

func (d UnknownDPT) String() string {
	return fmt.Sprintf("%v", []byte(d))
}

func (d UnknownDPT) Unit() string {
	return ""
}

func ReadConfig(filename string) (*Config, error) {
	var c Config
	c.Devices = make(map[cemi.IndividualAddr]string)
	c.Addresses = make(map[cemi.GroupAddr]addrNameType)
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	s := bufio.NewScanner(f)
	lineNum := 0
	for s.Scan() {
		lineNum++
		line := strings.TrimSpace(s.Text())
		if i := strings.IndexByte(line, '#'); i >= 0 {
			// strip comments
			line = strings.TrimSpace(line[0:i])
		}
		if len(line) == 0 {
			// empty line
			continue
		}
		tokens := strings.Fields(line)
		switch tokens[0] {
		case "mqtt-server":
			if len(tokens) != 2 {
				return nil, fmt.Errorf("syntax error in %s line %d", filename, lineNum)
			}
			c.MQTTServer = tokens[1]
		case "mqtt-prefix1":
			if len(tokens) != 2 {
				return nil, fmt.Errorf("syntax error in %s line %d", filename, lineNum)
			}
			c.MQTTPrefix1 = tokens[1]
		case "mqtt-prefix2":
			if len(tokens) != 2 {
				return nil, fmt.Errorf("syntax error in %s line %d", filename, lineNum)
			}
			c.MQTTPrefix2 = tokens[1]
		case "logdir":
			if len(tokens) != 2 {
				return nil, fmt.Errorf("syntax error in %s line %d", filename, lineNum)
			}
			c.Logdir = tokens[1]
		case "port":
			if len(tokens) != 2 {
				return nil, fmt.Errorf("syntax error in %s line %d", filename, lineNum)
			}
			c.Port, err = strconv.Atoi(tokens[1])
			if err != nil {
				return nil, fmt.Errorf("%s line %d: %w", filename, lineNum, err)
			}
		case "device":
			if len(tokens) != 3 {
				return nil, fmt.Errorf("syntax error in %s line %d", filename, lineNum)
			}
			// fmt.Printf("line %d: new device: %v\n", lineNum, tokens)
			addr, err := cemi.NewIndividualAddrString(tokens[1])
			if err != nil {
				return nil, fmt.Errorf("error in %s line %d: %w", filename, lineNum, err)
			}
			c.Devices[addr] = tokens[2]
		case "address":
			if len(tokens) != 4 {
				return nil, fmt.Errorf("syntax error in %s line %d", filename, lineNum)
			}
			aAddr := tokens[1]
			aDPT := tokens[2]
			aName := tokens[3]
			// fmt.Printf("line %d: new address: %v\n", lineNum, tokens)
			addr, err := cemi.NewGroupAddrString(aAddr)
			if err != nil {
				return nil, fmt.Errorf("error in %s line %d: %w", filename, lineNum, err)
			}
			c.Addresses[addr] = addrNameType{Name: aName, DPT: aDPT}
		default:
			return nil, fmt.Errorf("syntax error in %s line %d: unrecognized token %s", filename, lineNum, tokens[0])
		}
	}
	return &c, nil
}
