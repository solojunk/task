package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/gizak/termui"
)

type CpuClocks struct {
	Total   uint64
	User    uint64
	System  uint64
	Idle    uint64
	Wa      uint64
	Hi      uint64
	Si      uint64
	St      uint64
	Percent uint64
}

var (
	lastCpuTime CpuClocks
)

func init() {
	GetFirstCpuData()
}

func GetFirstCpuData() {
	//读取/proc/stat文件内容
	bs, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}

	//读取第一行CPU状态信息
	reader := bufio.NewReader(bytes.NewBuffer(bs))
	data, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	//分割CPU状态信息行
	fields := strings.Fields(data)
	if len(fields) < 9 {
		return
	}

	lastCpuTime.User, _ = strconv.ParseUint(fields[1], 10, 64)
	lastCpuTime.System, _ = strconv.ParseUint(fields[3], 10, 64)
	lastCpuTime.Idle, _ = strconv.ParseUint(fields[4], 10, 64)
	lastCpuTime.Wa, _ = strconv.ParseUint(fields[5], 10, 64)
	lastCpuTime.Hi, _ = strconv.ParseUint(fields[6], 10, 64)
	lastCpuTime.Si, _ = strconv.ParseUint(fields[7], 10, 64)
	lastCpuTime.St, _ = strconv.ParseUint(fields[8], 10, 64)
	lastCpuTime.Total = lastCpuTime.User + lastCpuTime.System + lastCpuTime.Idle + lastCpuTime.Wa + lastCpuTime.Hi + lastCpuTime.Si + lastCpuTime.St
}

func RefreshCpuView(p *termui.Par, lc *termui.LineChart, chs chan bool) {
	defer func(ch chan bool) {
		ch <- true
	}(chs)

	//读取/proc/stat文件内容
	bs, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}

	//读取第一行CPU状态信息
	reader := bufio.NewReader(bytes.NewBuffer(bs))
	data, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	//分割CPU状态信息行
	fields := strings.Fields(data)
	if len(fields) < 9 {
		return
	}

	user, _ := strconv.ParseUint(fields[1], 10, 64)
	system, _ := strconv.ParseUint(fields[3], 10, 64)
	idle, _ := strconv.ParseUint(fields[4], 10, 64)
	wa, _ := strconv.ParseUint(fields[5], 10, 64)
	hi, _ := strconv.ParseUint(fields[6], 10, 64)
	si, _ := strconv.ParseUint(fields[7], 10, 64)
	st, _ := strconv.ParseUint(fields[8], 10, 64)
	total := user + system + idle + wa + hi + si + st
	deltaTotal := float64(total - lastCpuTime.Total)

	userPercent := float64(user-lastCpuTime.User) * 100 / deltaTotal
	systemPercent := float64(system-lastCpuTime.System) * 100 / deltaTotal
	idlePercent := float64(idle-lastCpuTime.Idle) * 100 / deltaTotal
	waPercent := float64(wa-lastCpuTime.Wa) * 100 / deltaTotal
	hiPercent := float64(hi-lastCpuTime.Hi) * 100 / deltaTotal
	siPercent := float64(si-lastCpuTime.Si) * 100 / deltaTotal
	stPercent := float64(st-lastCpuTime.St) * 100 / deltaTotal
	percent := 100 - idlePercent

	lastCpuTime.User = user
	lastCpuTime.System = system
	lastCpuTime.Idle = idle
	lastCpuTime.Wa = wa
	lastCpuTime.Hi = hi
	lastCpuTime.Si = si
	lastCpuTime.St = st
	lastCpuTime.Total = total

	p.Text = fmt.Sprintf("User: %.1f%%   System: %.1f%%   Idle: %.1f%%   Wa: %.1f%%   Hi: %.1f%%   Si: %.1f%%   St: %.1f%%",
		userPercent, systemPercent, idlePercent, waPercent, hiPercent, siPercent, stPercent)

	lc.BorderLabel = fmt.Sprintf("CPU Used %.1f%%", percent)
	procStat := lc.Data[1:]
	procStat = append(procStat, percent)
	lc.Data = procStat

	if bWriteXlsx {
		xlsx.SetCellValue("CPU", fmt.Sprintf("A%d", cpuXlsxCount+2), time.Now().Format("15:04:05"))
		xlsx.SetCellValue("CPU", fmt.Sprintf("B%d", cpuXlsxCount+2), percent)
		cpuXlsxCount += 1
	}
}

func RefreshCpuData() {
	//读取/proc/stat文件内容
	bs, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}

	//读取第一行CPU状态信息
	reader := bufio.NewReader(bytes.NewBuffer(bs))
	data, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	//分割CPU状态信息行
	fields := strings.Fields(data)
	if len(fields) < 9 {
		return
	}

	user, _ := strconv.ParseUint(fields[1], 10, 64)
	system, _ := strconv.ParseUint(fields[3], 10, 64)
	idle, _ := strconv.ParseUint(fields[4], 10, 64)
	wa, _ := strconv.ParseUint(fields[5], 10, 64)
	hi, _ := strconv.ParseUint(fields[6], 10, 64)
	si, _ := strconv.ParseUint(fields[7], 10, 64)
	st, _ := strconv.ParseUint(fields[8], 10, 64)
	total := user + system + idle + wa + hi + si + st
	deltaTotal := float64(total - lastCpuTime.Total)

	idlePercent := float64(idle-lastCpuTime.Idle) * 100 / deltaTotal
	percent := 100 - idlePercent
	fmt.Println(percent)

	if bWriteXlsx {
		xlsx.SetCellValue("CPU", fmt.Sprintf("A%d", cpuXlsxCount+2), time.Now().Format("15:04:05"))
		xlsx.SetCellValue("CPU", fmt.Sprintf("B%d", cpuXlsxCount+2), percent)
		cpuXlsxCount += 1
	}
}
