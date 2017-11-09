package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gizak/termui"
)

var (
	lastTotalTime uint64
	lastPcpuTime  map[string]uint64 = map[string]uint64{}
	procUser      map[string]string = map[string]string{}
)

//获取首批数据
func GetFirstProcessData(pids []string) {
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

	User, _ := strconv.ParseUint(fields[1], 10, 64)
	System, _ := strconv.ParseUint(fields[3], 10, 64)
	Idle, _ := strconv.ParseUint(fields[4], 10, 64)
	Wa, _ := strconv.ParseUint(fields[5], 10, 64)
	Hi, _ := strconv.ParseUint(fields[6], 10, 64)
	Si, _ := strconv.ParseUint(fields[7], 10, 64)
	St, _ := strconv.ParseUint(fields[8], 10, 64)
	lastTotalTime = User + System + Idle + Wa + Hi + Si + St

	for _, pid := range pids {
		//读取/proc/pid/stat文件内容
		file := fmt.Sprintf("/proc/%s/stat", pid)
		bs, err := ioutil.ReadFile(file)
		if err != nil {
			return
		}

		//读取第一行状态信息
		reader := bufio.NewReader(bytes.NewBuffer(bs))
		data, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		//分割状态信息行
		fields := strings.Fields(data)
		if len(fields) < 25 {
			return
		}

		utime, _ := strconv.ParseUint(fields[13], 10, 64)
		stime, _ := strconv.ParseUint(fields[14], 10, 64)
		wutime, _ := strconv.ParseUint(fields[15], 10, 64)
		wstime, _ := strconv.ParseUint(fields[16], 10, 64)
		lastPcpuTime[pid] = utime + stime + wutime + wstime

		syntax := fmt.Sprintf("ps -eo pid,user | grep -P '^\\s*%s\\s' | awk {'printf $2'}", pid)
		cmd := exec.Command("sh", "-c", syntax)

		var out bytes.Buffer
		cmd.Stdout = &out
		err = cmd.Run()
		if err != nil {
			return
		}
		procUser[pid] = strings.TrimRight(out.String(), "\n")
	}
}

//刷新界面数据
func RefreshProcessView(pids []string, p *termui.Par, lcs map[string]*termui.LineChart, chs chan bool) {
	defer func(ch chan bool) {
		ch <- true
	}(chs)

	if len(pids) == 0 {
		return
	}

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

	User, _ := strconv.ParseUint(fields[1], 10, 64)
	System, _ := strconv.ParseUint(fields[3], 10, 64)
	Idle, _ := strconv.ParseUint(fields[4], 10, 64)
	Wa, _ := strconv.ParseUint(fields[5], 10, 64)
	Hi, _ := strconv.ParseUint(fields[6], 10, 64)
	Si, _ := strconv.ParseUint(fields[7], 10, 64)
	St, _ := strconv.ParseUint(fields[8], 10, 64)
	totalTime := User + System + Idle + Wa + Hi + Si + St
	deltaTotalTime := totalTime - lastTotalTime
	lastTotalTime = totalTime

	//读取系统总使用内存
	bs, err = ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		return
	}
	reader = bufio.NewReader(bytes.NewBuffer(bs))
	data, _ = reader.ReadString('\n')
	memTotal, _ := strconv.ParseUint(strings.Fields(data)[1], 10, 64)
	data, _ = reader.ReadString('\n')
	memFree, _ := strconv.ParseUint(strings.Fields(data)[1], 10, 64)
	memUsed := float64(memTotal-memFree) / 1024

	if bWriteXlsx {
		xlsx.SetCellValue("Process", fmt.Sprintf("A%d", procXlsxCount+2), time.Now().Format("15:04:05"))
	}

	var parText string
	for i, pid := range pids {
		//读取/proc/pid/stat文件内容
		file := fmt.Sprintf("/proc/%s/stat", pid)
		bs, err = ioutil.ReadFile(file)
		if err != nil {
			return
		}

		//读取第一行状态信息
		reader = bufio.NewReader(bytes.NewBuffer(bs))
		data, err = reader.ReadString('\n')
		if err != nil {
			return
		}

		left := strings.Index(data, "(")
		right := strings.Index(data, ")")
		cmd := data[left+1 : right]

		//分割状态信息行
		fields = strings.Fields(data[right+1:])
		if len(fields) < 25 {
			return
		}

		utime, _ := strconv.ParseUint(fields[11], 10, 64)
		stime, _ := strconv.ParseUint(fields[12], 10, 64)
		wutime, _ := strconv.ParseUint(fields[13], 10, 64)
		wstime, _ := strconv.ParseUint(fields[14], 10, 64)
		virt, _ := strconv.ParseUint(fields[20], 10, 64)
		rss, _ := strconv.ParseUint(fields[21], 10, 64)

		time := utime + stime + wutime + wstime
		pcpu := float64(time-lastPcpuTime[pid]) * cpuCores * 100 / float64(deltaTotalTime)
		lastPcpuTime[pid] = time
		pmem := float64(rss) / 2.56 / memUsed

		if len(parText) > 0 {
			parText = fmt.Sprintf("%s\nPid: %6s User: %-10s Virt: %6dM Res: %6dM %%Cpu: %3.1f%% %%Mem: %3.1f%% Cmd: %-15s",
				parText, pid, procUser[pid], virt/1048576, rss/256, pcpu, pmem, cmd)
		} else {
			parText = fmt.Sprintf("Pid: %6s User: %-10s Virt: %6dM Res: %6dM %%Cpu: %3.1f%% %%Mem: %3.1f%% Cmd: %-15s",
				pid, procUser[pid], virt/1048576, rss/256, pcpu, pmem, cmd)
		}

		if lc, exist := lcs[fmt.Sprintf("%s-cpu", pid)]; exist {
			lc.BorderLabel = fmt.Sprintf("Process CPU Used %s %.1f%%", pid, pcpu)
			procStat := lc.Data[1:]
			procStat = append(procStat, pcpu)
			lc.Data = procStat
		}

		if lc, exist := lcs[fmt.Sprintf("%s-mem", pid)]; exist {
			lc.BorderLabel = fmt.Sprintf("Process Mem Used %s %.1f%%", pid, pmem)
			procStat := lc.Data[1:]
			procStat = append(procStat, pmem)
			lc.Data = procStat
		}

		if bWriteXlsx {
			xlsx.SetCellValue("Process", fmt.Sprintf("%c%d", 'B'+i*2, procXlsxCount+2), pcpu)
			xlsx.SetCellValue("Process", fmt.Sprintf("%c%d", 'B'+i*2+1, procXlsxCount+2), pmem)
		}
	}
	procXlsxCount += 1

	if len(parText) > 0 {
		p.Text = parText
	}
}

//刷新后台数据
func RefreshProcessData(pids []string) {
	if len(pids) == 0 {
		return
	}

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

	User, _ := strconv.ParseUint(fields[1], 10, 64)
	System, _ := strconv.ParseUint(fields[3], 10, 64)
	Idle, _ := strconv.ParseUint(fields[4], 10, 64)
	Wa, _ := strconv.ParseUint(fields[5], 10, 64)
	Hi, _ := strconv.ParseUint(fields[6], 10, 64)
	Si, _ := strconv.ParseUint(fields[7], 10, 64)
	St, _ := strconv.ParseUint(fields[8], 10, 64)
	totalTime := User + System + Idle + Wa + Hi + Si + St
	deltaTotalTime := totalTime - lastTotalTime
	lastTotalTime = totalTime

	//读取系统总使用内存
	bs, err = ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		return
	}
	reader = bufio.NewReader(bytes.NewBuffer(bs))
	data, _ = reader.ReadString('\n')
	memTotal, _ := strconv.ParseUint(strings.Fields(data)[1], 10, 64)
	data, _ = reader.ReadString('\n')
	memFree, _ := strconv.ParseUint(strings.Fields(data)[1], 10, 64)
	memUsed := float64(memTotal-memFree) / 1024

	if bWriteXlsx {
		xlsx.SetCellValue("Process", fmt.Sprintf("A%d", procXlsxCount+2), time.Now().Format("15:04:05"))
	}

	for i, pid := range pids {
		//读取/proc/pid/stat文件内容
		file := fmt.Sprintf("/proc/%s/stat", pid)
		bs, err = ioutil.ReadFile(file)
		if err != nil {
			return
		}

		//读取第一行状态信息
		reader = bufio.NewReader(bytes.NewBuffer(bs))
		data, err = reader.ReadString('\n')
		if err != nil {
			return
		}

		//分割状态信息行
		right := strings.Index(data, ")")
		fields = strings.Fields(data[right+1:])
		if len(fields) < 25 {
			return
		}

		utime, _ := strconv.ParseUint(fields[11], 10, 64)
		stime, _ := strconv.ParseUint(fields[12], 10, 64)
		wutime, _ := strconv.ParseUint(fields[13], 10, 64)
		wstime, _ := strconv.ParseUint(fields[14], 10, 64)
		rss, _ := strconv.ParseUint(fields[21], 10, 64)

		time := utime + stime + wutime + wstime
		pcpu := float64(time-lastPcpuTime[pid]) * cpuCores * 100 / float64(deltaTotalTime)
		lastPcpuTime[pid] = time
		pmem := float64(rss) / 2.56 / memUsed

		if bWriteXlsx {
			xlsx.SetCellValue("Process", fmt.Sprintf("%c%d", 'B'+i*2, procXlsxCount+2), pcpu)
			xlsx.SetCellValue("Process", fmt.Sprintf("%c%d", 'B'+i*2+1, procXlsxCount+2), pmem)
		}
	}
	procXlsxCount += 1
}
