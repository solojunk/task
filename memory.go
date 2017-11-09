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

//刷新界面数据
func RefreshMemoryView(p *termui.Par, lc *termui.LineChart, chs chan bool) {
	defer func(ch chan bool) {
		ch <- true
	}(chs)

	//读取/proc/meminfo文件内容
	bs, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		return
	}

	//读取meminfo总内存和剩余内存信息
	reader := bufio.NewReader(bytes.NewBuffer(bs))
	var memTotal, memFree, buffers, cached, swapTotal, swapFree uint64

	for {
		data, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return
		} else if memTotal > 0 && memFree > 0 && buffers > 0 && cached > 0 && swapTotal > 0 && swapFree > 0 {
			break
		}

		fields := strings.Fields(data)
		if len(fields) < 3 {
			continue
		}
		if fields[0] == "MemTotal:" {
			memTotal, err = strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				continue
			}
		} else if fields[0] == "MemFree:" {
			memFree, err = strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				continue
			}
		} else if fields[0] == "Buffers:" {
			buffers, err = strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				continue
			}
		} else if fields[0] == "Cached:" {
			cached, err = strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				continue
			}
		} else if fields[0] == "SwapTotal:" {
			swapTotal, err = strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				continue
			}
		} else if fields[0] == "SwapFree:" {
			swapFree, err = strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				continue
			}
		}
	}

	realUsed := memTotal - memFree - buffers - cached
	percent := realUsed * 100 / memTotal

	p.Text = fmt.Sprintf("Mem:  %9dk total, %9dk used, %9dk free, %9dk buffers\nSwap: %9dk total, %9dk used, %9dk free, %9dk cached",
		memTotal, memTotal-memFree, memFree, buffers, swapTotal, swapTotal-swapFree, swapFree, cached)

	lc.BorderLabel = fmt.Sprintf("Memory Used %dk %d%%", realUsed, percent)
	procStat := lc.Data[1:]
	procStat = append(procStat, float64(percent))
	lc.Data = procStat

	if bWriteXlsx {
		xlsx.SetCellValue("Memory", fmt.Sprintf("A%d", memXlsxCount+2), time.Now().Format("15:04:05"))
		xlsx.SetCellValue("Memory", fmt.Sprintf("B%d", memXlsxCount+2), percent)
		memXlsxCount += 1
	}
}

//刷新后台数据
func RefreshMemoryData() {
	//读取/proc/meminfo文件内容
	bs, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		return
	}

	//读取meminfo总内存和剩余内存信息
	reader := bufio.NewReader(bytes.NewBuffer(bs))
	var memTotal, memFree, buffers, cached, swapTotal, swapFree uint64

	for {
		data, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return
		} else if memTotal > 0 && memFree > 0 && buffers > 0 && cached > 0 && swapTotal > 0 && swapFree > 0 {
			break
		}

		fields := strings.Fields(data)
		if len(fields) < 3 {
			continue
		}
		if fields[0] == "MemTotal:" {
			memTotal, err = strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				continue
			}
		} else if fields[0] == "MemFree:" {
			memFree, err = strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				continue
			}
		} else if fields[0] == "Buffers:" {
			buffers, err = strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				continue
			}
		} else if fields[0] == "Cached:" {
			cached, err = strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				continue
			}
		} else if fields[0] == "SwapTotal:" {
			swapTotal, err = strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				continue
			}
		} else if fields[0] == "SwapFree:" {
			swapFree, err = strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				continue
			}
		}
	}

	realUsed := memTotal - memFree - buffers - cached
	percent := realUsed * 100 / memTotal

	if bWriteXlsx {
		xlsx.SetCellValue("Memory", fmt.Sprintf("A%d", memXlsxCount+2), time.Now().Format("15:04:05"))
		xlsx.SetCellValue("Memory", fmt.Sprintf("B%d", memXlsxCount+2), percent)
		memXlsxCount += 1
	}
}
