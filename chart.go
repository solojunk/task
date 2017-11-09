package main

import (
	"fmt"

	"github.com/xuri/excelize"
)

var (
	bWriteXlsx    bool
	xlsxName      string
	xlsx          *excelize.File
	cpuXlsxCount  uint64
	memXlsxCount  uint64
	diskXlsxCount uint64
	netXlsxCount  uint64
	loadXlsxCount uint64
	procXlsxCount uint64
)

//InitXlsxFile 初始化xlsx文件结构
func InitXlsxFile(pids []string, filename string) {
	bWriteXlsx = true
	xlsxName = filename

	xlsx = excelize.NewFile()

	xlsx.SetSheetName("Sheet1", "CPU")
	xlsx.SetCellValue("CPU", "B1", "CPU")

	xlsx.NewSheet("Memory")
	xlsx.SetCellValue("Memory", "B1", "Memory")

	xlsx.NewSheet("Disk")
	count := 0
	for name := range ioMap {
		xlsx.SetCellValue("Disk", fmt.Sprintf("%c1", 'B'+count), name)
		count++
	}

	xlsx.NewSheet("Network")
	for i := 0; i < len(netMap); i++ {
		xlsx.SetCellValue("Network", fmt.Sprintf("%c1", 'B'+i*2), "RX")
		xlsx.SetCellValue("Network", fmt.Sprintf("%c1", 'B'+i*2+1), "TX")
	}

	xlsx.NewSheet("Loadavg")
	xlsx.SetCellValue("Loadavg", "B1", "util")
	xlsx.SetCellValue("Loadavg", "C1", "1min")
	xlsx.SetCellValue("Loadavg", "D1", "5min")
	xlsx.SetCellValue("Loadavg", "E1", "15min")

	if len(pids) > 0 {
		xlsx.NewSheet("Process")
		for i := range pids {
			xlsx.SetCellValue("Process", fmt.Sprintf("%c1", 'B'+i*2), "pcpu")
			xlsx.SetCellValue("Process", fmt.Sprintf("%c1", 'B'+i*2+1), "pmem")
		}
	}
}

//FinishAndGenerateChart 结束时保存数据到文件
func FinishAndGenerateChart() {
	cpuChart := fmt.Sprintf(`{"type":"line","series":[{"name":"CPU!$B$1","categories":"CPU!$A$2:$A$%d","values":"CPU!$B$2:$B$%d"}],"title":{"name":"CPU使用率(%%)"}}`, cpuXlsxCount+1, cpuXlsxCount+1)
	xlsx.AddChart("CPU", "C1", cpuChart)

	memChart := fmt.Sprintf(`{"type":"line","series":[{"name":"Memory!$B$1","categories":"Memory!$A$2:$A$%d","values":"Memory!$B$2:$B$%d"}],"title":{"name":"内存使用率(%%)"}}`, memXlsxCount+1, memXlsxCount+1)
	xlsx.AddChart("Memory", "C1", memChart)

	var series string
	for i := 0; i < len(ioMap); i++ {
		ioSeries := fmt.Sprintf(`{"name":"Disk!$%c$1","categories":"Disk!$A$2:$A$%d","values":"Disk!$%c$2:$%c$%d"}`, 'B'+i, diskXlsxCount+1, 'B'+i, 'B'+i, diskXlsxCount+1)
		if len(series) > 0 {
			series = fmt.Sprintf("%s,%s", series, ioSeries)
		} else {
			series = fmt.Sprintf("%s", ioSeries)
		}
	}
	diskChart := fmt.Sprintf(`{"type":"line","series":[%s],"title":{"name":"磁盘IO使用率(%%)"}}`, series)
	xlsx.AddChart("Disk", fmt.Sprintf("%c1", 'B'+len(ioMap)), diskChart)

	i := 0
	for name := range netMap {
		series = fmt.Sprintf(`{"name":"Network!$%c$1","categories":"Network!$A$2:$A$%d","values":"Network!$%c$2:$%c$%d"}`, 'B'+i*2, netXlsxCount+1, 'B'+i*2, 'B'+i*2, netXlsxCount+1)
		ioSeries := fmt.Sprintf(`{"name":"Network!$%c$1","categories":"Network!$A$2:$A$%d","values":"Network!$%c$2:$%c$%d"}`, 'B'+i*2+1, netXlsxCount+1, 'B'+i*2+1, 'B'+i*2+1, netXlsxCount+1)
		series = fmt.Sprintf("%s,%s", series, ioSeries)
		i++

		netChart := fmt.Sprintf(`{"type":"line","series":[%s],"title":{"name":"%s网卡流量速率(KB/s)"}}`, series, name)
		xlsx.AddChart("Network", fmt.Sprintf("%c%d", 'B'+len(netMap)*2, i*15-14), netChart)
	}

	for i := 0; i < 4; i++ {
		var name string
		switch i {
		case 0:
			name = "系统负载百分比(%)"
		case 1:
			name = "1分钟系统负载"
		case 2:
			name = "5分钟系统负载"
		case 3:
			name = "15分钟系统负载"
		}
		loadChart := fmt.Sprintf(`{"type":"line","series":[{"name":"Loadavg!$%c$1","categories":"Loadavg!$A$2:$A$%d","values":"Loadavg!$%c$2:$%c$%d"}],"title":{"name":"%s"}}`, 'B'+i, loadXlsxCount+1, 'B'+i, 'B'+i, loadXlsxCount+1, name)
		xlsx.AddChart("Loadavg", fmt.Sprintf("F%d", i*15+1), loadChart)
	}

	i = 0
	for pid := range procUser {
		series = fmt.Sprintf(`{"name":"Process!$%c$1","categories":"Process!$A$2:$A$%d","values":"Process!$%c$2:$%c$%d"}`, 'B'+i*2, procXlsxCount+1, 'B'+i*2, 'B'+i*2, procXlsxCount+1)
		procSeries := fmt.Sprintf(`{"name":"Process!$%c$1","categories":"Process!$A$2:$A$%d","values":"Process!$%c$2:$%c$%d"}`, 'B'+i*2+1, procXlsxCount+1, 'B'+i*2+1, 'B'+i*2+1, procXlsxCount+1)
		series = fmt.Sprintf("%s,%s", series, procSeries)
		i++

		procChart := fmt.Sprintf(`{"type":"line","series":[%s],"title":{"name":"进程PID=%s的资源使用百分比(%%)"}}`, series, pid)
		xlsx.AddChart("Process", fmt.Sprintf("%c%d", 'B'+len(procUser)*2, i*15-14), procChart)
	}

	err := xlsx.SaveAs(xlsxName)
	if err != nil {
		logger.Printf("Save xlsx file failed! error:%s\n", err.Error())
	}
}
