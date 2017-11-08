package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gizak/termui"
)

var cpuCores float64 = float64(runtime.NumCPU())

func RefreshLoadavgView(p *termui.Par, lcs map[string]*termui.LineChart, chs chan bool) {
	defer func(ch chan bool) {
		ch <- true
	}(chs)

	//读取/proc/meminfo文件内容
	bs, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		return
	}

	//读取前三个字段系统负载信息
	reader := bufio.NewReader(bytes.NewBuffer(bs))
	data, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	fields := strings.Fields(data)
	if len(fields) < 3 {
		return
	}

	Load1min, _ := strconv.ParseFloat(fields[0], 64)
	Load5min, _ := strconv.ParseFloat(fields[1], 64)
	Load15min, _ := strconv.ParseFloat(fields[2], 64)
	Util := Load1min * 100 / cpuCores

	p.Text = fmt.Sprintf("Loadavg:   %.2f   %.2f   %.2f   SystemUtil:   %.2f%%", Load1min, Load5min, Load15min, Util)

	for name, lc := range lcs {
		procStat := lc.Data[1:]
		switch name {
		case "percent":
			procStat = append(procStat, Util)
		case "1min":
			procStat = append(procStat, Load1min)
		case "5min":
			procStat = append(procStat, Load5min)
		case "15min":
			procStat = append(procStat, Load15min)
		}
		lc.Data = procStat
	}

	if bWriteXlsx {
		xlsx.SetCellValue("Loadavg", fmt.Sprintf("A%d", loadXlsxCount+2), time.Now().Format("15:04:05"))
		xlsx.SetCellValue("Loadavg", fmt.Sprintf("B%d", loadXlsxCount+2), Util)
		xlsx.SetCellValue("Loadavg", fmt.Sprintf("C%d", loadXlsxCount+2), Load1min)
		xlsx.SetCellValue("Loadavg", fmt.Sprintf("D%d", loadXlsxCount+2), Load5min)
		xlsx.SetCellValue("Loadavg", fmt.Sprintf("E%d", loadXlsxCount+2), Load15min)
		loadXlsxCount += 1
	}
}

func RefreshLoadavgData() {
	//读取/proc/meminfo文件内容
	bs, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		return
	}

	//读取前三个字段系统负载信息
	reader := bufio.NewReader(bytes.NewBuffer(bs))
	data, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	fields := strings.Fields(data)
	if len(fields) < 3 {
		return
	}

	Load1min, _ := strconv.ParseFloat(fields[0], 64)
	Load5min, _ := strconv.ParseFloat(fields[1], 64)
	Load15min, _ := strconv.ParseFloat(fields[2], 64)
	Util := Load1min * 100 / float64(cpuCores)

	if bWriteXlsx {
		xlsx.SetCellValue("Loadavg", fmt.Sprintf("A%d", loadXlsxCount+2), time.Now().Format("15:04:05"))
		xlsx.SetCellValue("Loadavg", fmt.Sprintf("B%d", loadXlsxCount+2), Util)
		xlsx.SetCellValue("Loadavg", fmt.Sprintf("C%d", loadXlsxCount+2), Load1min)
		xlsx.SetCellValue("Loadavg", fmt.Sprintf("D%d", loadXlsxCount+2), Load5min)
		xlsx.SetCellValue("Loadavg", fmt.Sprintf("E%d", loadXlsxCount+2), Load15min)
		loadXlsxCount += 1
	}
}
