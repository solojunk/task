package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/gizak/termui"
)

type NetworkInfo struct {
	Ip  string
	Mac string
}

type NetworkStat struct {
	RxBytes uint64
	TxBytes uint64
	RxSpeed uint64
	TxSpeed uint64
}

var (
	netInfo map[string]*NetworkInfo = map[string]*NetworkInfo{}
	netMap  map[string]*NetworkStat = map[string]*NetworkStat{}
)

func init() {
	GetFirstNetworkData()
}

func GetFirstNetworkData() {
	//读取/proc/net/dev文件内容
	bs, err := ioutil.ReadFile("/proc/net/dev")
	if err != nil {
		return
	}

	//读取网卡流量信息
	reader := bufio.NewReader(bytes.NewBuffer(bs))

	//保存最新数据到数组
	for {
		data, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return
		}

		idx := strings.Index(data, ":")
		if idx < 0 {
			continue
		}

		name := strings.TrimSpace(data[0:idx])

		fields := strings.Fields(data[idx+1:])
		if len(fields) < 16 {
			continue
		}

		var lastNetStat NetworkStat
		lastNetStat.RxBytes, _ = strconv.ParseUint(fields[0], 10, 64)
		lastNetStat.TxBytes, _ = strconv.ParseUint(fields[8], 10, 64)

		if lastNetStat.RxBytes > 0 || lastNetStat.TxBytes > 0 {
			netMap[name] = &lastNetStat
		}

		ifcfg := fmt.Sprintf("/etc/sysconfig/network-scripts/ifcfg-%s", name)
		//读取ifcfg-ethX文件内容
		ifData, err := ioutil.ReadFile(ifcfg)
		if err != nil {
			continue
		}

		//读取网卡信息
		var net NetworkInfo
		ifReader := bufio.NewReader(bytes.NewBuffer(ifData))
		for {
			ifLine, err := ifReader.ReadString('\n')
			if err == io.EOF {
				break
			} else if err != nil {
				break
			}

			fields := strings.Split(ifLine, "=")
			if len(fields) != 2 {
				continue
			}

			fields[1] = strings.TrimRight(fields[1], "\n")
			if fields[0] == "IPADDR" {
				net.Ip = strings.Trim(fields[1], "\"")
			} else if fields[0] == "HWADDR" {
				net.Mac = strings.Trim(fields[1], "\"")
			}
			if len(net.Ip) > 0 && len(net.Mac) > 0 {
				break
			}
		}
		if len(net.Ip) == 0 {
			net.Ip = "0.0.0.0"
		}
		if len(net.Mac) == 0 {
			net.Mac = "00:00:00:00:00:00"
		}
		netInfo[name] = &net
	}
}

func RefreshNetworkView(interval uint64, p *termui.Par, lcs map[string]*termui.LineChart, chs chan bool) {
	defer func(ch chan bool) {
		ch <- true
	}(chs)

	//读取/proc/net/dev文件内容
	bs, err := ioutil.ReadFile("/proc/net/dev")
	if err != nil {
		return
	}

	//读取网卡流量信息
	var parText string
	reader := bufio.NewReader(bytes.NewBuffer(bs))

	//保存最新数据到数组
	for {
		data, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return
		}

		idx := strings.Index(data, ":")
		if idx < 0 {
			continue
		}

		name := strings.TrimSpace(data[0:idx])

		fields := strings.Fields(data[idx+1:])
		if len(fields) < 16 {
			continue
		}

		lastStat, exist := netMap[name]
		if !exist {
			continue
		}

		RxBytes, _ := strconv.ParseUint(fields[0], 10, 64)
		TxBytes, _ := strconv.ParseUint(fields[8], 10, 64)
		lastStat.RxSpeed = (RxBytes - lastStat.RxBytes) / interval / 1024
		lastStat.TxSpeed = (TxBytes - lastStat.TxBytes) / interval / 1024
		lastStat.RxBytes = RxBytes
		lastStat.TxBytes = TxBytes

		net, exist := netInfo[name]
		if !exist {
			continue
		}

		if len(parText) > 0 {
			parText = fmt.Sprintf("%s\nName: %4s Ip: %15s Mac: %17s RXBytes: %16d TXBytes: %16d",
				parText, name, net.Ip, net.Mac, RxBytes, TxBytes)
		} else {
			parText = fmt.Sprintf("Name: %4s Ip: %15s Mac: %17s RXBytes: %16d TXBytes: %16d",
				name, net.Ip, net.Mac, RxBytes, TxBytes)
		}
	}

	if len(parText) > 0 {
		p.Text = parText
	}

	if bWriteXlsx {
		xlsx.SetCellValue("Network", fmt.Sprintf("A%d", netXlsxCount+2), time.Now().Format("15:04:05"))
	}

	var netCount int = 0
	for name, stat := range netMap {
		if lc, exist := lcs[fmt.Sprintf("%s-Rx", name)]; exist {
			lc.BorderLabel = fmt.Sprintf("Network RX Speed %s %dKB/s", name, stat.RxSpeed)
			procStat := lc.Data[1:]
			procStat = append(procStat, float64(stat.RxSpeed))
			lc.Data = procStat
		}

		if lc, exist := lcs[fmt.Sprintf("%s-Tx", name)]; exist {
			lc.BorderLabel = fmt.Sprintf("Network TX Speed %s %dKB/s", name, stat.TxSpeed)
			procStat := lc.Data[1:]
			procStat = append(procStat, float64(stat.TxSpeed))
			lc.Data = procStat
		}

		if bWriteXlsx {
			xlsx.SetCellValue("Network", fmt.Sprintf("%c%d", 'B'+netCount*2, netXlsxCount+2), stat.RxSpeed)
			xlsx.SetCellValue("Network", fmt.Sprintf("%c%d", 'B'+netCount*2+1, netXlsxCount+2), stat.TxSpeed)
			netCount += 1
		}
	}
	netXlsxCount += 1
}

func RefreshNetworkData(interval uint64) {
	//读取/proc/net/dev文件内容
	bs, err := ioutil.ReadFile("/proc/net/dev")
	if err != nil {
		return
	}

	//读取网卡流量信息
	reader := bufio.NewReader(bytes.NewBuffer(bs))

	//保存最新数据到数组
	for {
		data, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return
		}

		idx := strings.Index(data, ":")
		if idx < 0 {
			continue
		}

		name := strings.TrimSpace(data[0:idx])

		fields := strings.Fields(data[idx+1:])
		if len(fields) < 16 {
			continue
		}

		lastStat, exist := netMap[name]
		if !exist {
			continue
		}

		RxBytes, _ := strconv.ParseUint(fields[0], 10, 64)
		TxBytes, _ := strconv.ParseUint(fields[8], 10, 64)
		lastStat.RxSpeed = (RxBytes - lastStat.RxBytes) / interval / 1024
		lastStat.TxSpeed = (TxBytes - lastStat.TxBytes) / interval / 1024
		lastStat.RxBytes = RxBytes
		lastStat.TxBytes = TxBytes
	}

	if bWriteXlsx {
		var netCount int = 0
		xlsx.SetCellValue("Network", fmt.Sprintf("A%d", netXlsxCount+2), time.Now().Format("15:04:05"))
		for _, stat := range netMap {
			xlsx.SetCellValue("Network", fmt.Sprintf("%c%d", 'B'+netCount*2, netXlsxCount+2), stat.RxSpeed)
			xlsx.SetCellValue("Network", fmt.Sprintf("%c%d", 'B'+netCount*2+1, netXlsxCount+2), stat.TxSpeed)
			netCount += 1
		}
		netXlsxCount += 1
	}
}
