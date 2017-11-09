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

//CPUClocks CPU时钟周期
type CPUClocks struct {
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
	lastCPUTime CPUClocks
)

//前置函数
func init() {
	GetFirstCPUData()
}

//GetFirstCPUData 获取首批数据
func GetFirstCPUData() {
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

	lastCPUTime.User, _ = strconv.ParseUint(fields[1], 10, 64)
	lastCPUTime.System, _ = strconv.ParseUint(fields[3], 10, 64)
	lastCPUTime.Idle, _ = strconv.ParseUint(fields[4], 10, 64)
	lastCPUTime.Wa, _ = strconv.ParseUint(fields[5], 10, 64)
	lastCPUTime.Hi, _ = strconv.ParseUint(fields[6], 10, 64)
	lastCPUTime.Si, _ = strconv.ParseUint(fields[7], 10, 64)
	lastCPUTime.St, _ = strconv.ParseUint(fields[8], 10, 64)
	lastCPUTime.Total = lastCPUTime.User + lastCPUTime.System + lastCPUTime.Idle + lastCPUTime.Wa + lastCPUTime.Hi + lastCPUTime.Si + lastCPUTime.St
}

//RefreshCPUView 刷新界面数据
func RefreshCPUView(p *termui.Par, lc *termui.LineChart, chs chan bool) {
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
	deltaTotal := float64(total - lastCPUTime.Total)

	userPercent := float64(user-lastCPUTime.User) * 100 / deltaTotal
	systemPercent := float64(system-lastCPUTime.System) * 100 / deltaTotal
	idlePercent := float64(idle-lastCPUTime.Idle) * 100 / deltaTotal
	waPercent := float64(wa-lastCPUTime.Wa) * 100 / deltaTotal
	hiPercent := float64(hi-lastCPUTime.Hi) * 100 / deltaTotal
	siPercent := float64(si-lastCPUTime.Si) * 100 / deltaTotal
	stPercent := float64(st-lastCPUTime.St) * 100 / deltaTotal
	percent := 100 - idlePercent

	lastCPUTime.User = user
	lastCPUTime.System = system
	lastCPUTime.Idle = idle
	lastCPUTime.Wa = wa
	lastCPUTime.Hi = hi
	lastCPUTime.Si = si
	lastCPUTime.St = st
	lastCPUTime.Total = total

	p.Text = fmt.Sprintf("User: %.1f%%   System: %.1f%%   Idle: %.1f%%   Wa: %.1f%%   Hi: %.1f%%   Si: %.1f%%   St: %.1f%%",
		userPercent, systemPercent, idlePercent, waPercent, hiPercent, siPercent, stPercent)

	lc.BorderLabel = fmt.Sprintf("CPU Used %.1f%%", percent)
	procStat := lc.Data[1:]
	procStat = append(procStat, percent)
	lc.Data = procStat

	if bWriteXlsx {
		xlsx.SetCellValue("CPU", fmt.Sprintf("A%d", cpuXlsxCount+2), time.Now().Format("15:04:05"))
		xlsx.SetCellValue("CPU", fmt.Sprintf("B%d", cpuXlsxCount+2), percent)
		cpuXlsxCount++
	}
}

//RefreshCPUData 刷新后台数据
func RefreshCPUData() {
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
	deltaTotal := float64(total - lastCPUTime.Total)

	idlePercent := float64(idle-lastCPUTime.Idle) * 100 / deltaTotal
	percent := 100 - idlePercent
	fmt.Println(percent)

	if bWriteXlsx {
		xlsx.SetCellValue("CPU", fmt.Sprintf("A%d", cpuXlsxCount+2), time.Now().Format("15:04:05"))
		xlsx.SetCellValue("CPU", fmt.Sprintf("B%d", cpuXlsxCount+2), percent)
		cpuXlsxCount++
	}
}
