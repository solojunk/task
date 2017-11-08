package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gizak/termui"
)

type IOStat struct {
	ReadIO       uint64
	ReadSectors  uint64
	WriteIO      uint64
	WriteSectors uint64
	Util         uint64
	Percent      float64
}

var (
	fsMap map[string]string  = map[string]string{}
	ioMap map[string]*IOStat = map[string]*IOStat{}
)

func init() {
	GetFirstIOData()
}

func GetFirstIOData() {
	//读取/proc/diskstats文件内容
	bs, err := ioutil.ReadFile("/proc/diskstats")
	if err != nil {
		return
	}

	reader := bufio.NewReader(bytes.NewBuffer(bs))

	for {
		data, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return
		}

		fields := strings.Fields(data)
		if len(fields) < 14 {
			continue
		}

		major, _ := strconv.ParseUint(fields[0], 10, 64)
		minor, _ := strconv.ParseUint(fields[1], 10, 64)
		disk := fields[2]
		if major%8 != 0 || minor%16 != 0 {
			continue
		}

		var lastIOStat IOStat
		lastIOStat.ReadIO, _ = strconv.ParseUint(fields[3], 10, 64)
		lastIOStat.ReadSectors, _ = strconv.ParseUint(fields[5], 10, 64)
		lastIOStat.WriteIO, _ = strconv.ParseUint(fields[7], 10, 64)
		lastIOStat.WriteSectors, _ = strconv.ParseUint(fields[9], 10, 64)

		//读取/sys/block/sda/stat文件内容
		statFile := fmt.Sprintf("/sys/block/%s/stat", disk)
		statData, err := ioutil.ReadFile(statFile)
		if err != nil {
			continue
		}

		statReader := bufio.NewReader(bytes.NewBuffer(statData))
		statLine, err := statReader.ReadString('\n')
		if err != nil {
			continue
		}
		fields = strings.Fields(statLine)
		if len(fields) < 11 {
			continue
		}
		lastIOStat.Util, _ = strconv.ParseUint(fields[10], 10, 64)

		ioMap[disk] = &lastIOStat
	}
}

func GetDiskMounts() []string {
	mounts := make([]string, 0)

	//读取/proc/mounts文件内容
	bs, err := ioutil.ReadFile("/proc/mounts")
	if err != nil {
		return mounts
	}

	reader := bufio.NewReader(bytes.NewBuffer(bs))

	for {
		data, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return mounts
		}

		if !strings.HasPrefix(data, "/dev/") {
			continue
		}

		fields := strings.Fields(data)
		if len(fields) < 3 {
			continue
		}

		if _, exist := fsMap[fields[0]]; !exist {
			fsMap[fields[0]] = fields[1]
			mounts = append(mounts, fields[1])
		}
	}

	return mounts
}

func RefreshDiskView(interval uint64, gauges map[string]*termui.Gauge, p *termui.Par, lcs map[string]*termui.LineChart, chs chan bool) {
	defer func(ch chan bool) {
		ch <- true
	}(chs)

	var diskFree, percent uint64

	//读取关联的挂载点信息
	fs := syscall.Statfs_t{}
	for _, v := range fsMap {
		err := syscall.Statfs(v, &fs)
		if err != nil {
			return
		}

		diskFree = uint64(fs.Frsize) * fs.Bavail
		percent = (fs.Blocks - fs.Bfree) * 100 / (fs.Blocks - fs.Bfree + fs.Bavail)

		if gauge, exist := gauges[v]; exist {
			gauge.Percent = int(percent)
			gauge.Label = fmt.Sprintf("{{percent}}%% (%dMB free)", diskFree/1048576)
			switch {
			case percent < 60:
				gauge.BarColor = termui.ColorGreen
			case percent < 80:
				gauge.BarColor = termui.ColorYellow
			default:
				gauge.BarColor = termui.ColorRed
			}
		}
	}

	//读取/proc/diskstats文件内容
	bs, err := ioutil.ReadFile("/proc/diskstats")
	if err != nil {
		return
	}

	var parText string
	reader := bufio.NewReader(bytes.NewBuffer(bs))

	for {
		data, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return
		}

		fields := strings.Fields(data)
		if len(fields) < 14 {
			continue
		}

		major, _ := strconv.ParseUint(fields[0], 10, 64)
		minor, _ := strconv.ParseUint(fields[1], 10, 64)
		disk := fields[2]
		if major%8 != 0 || minor%16 != 0 {
			continue
		}

		lastIOStat, exist := ioMap[disk]
		if !exist {
			continue
		}

		ReadIO, _ := strconv.ParseUint(fields[3], 10, 64)
		ReadSectors, _ := strconv.ParseUint(fields[5], 10, 64)
		WriteIO, _ := strconv.ParseUint(fields[7], 10, 64)
		WriteSectors, _ := strconv.ParseUint(fields[9], 10, 64)

		iops := (ReadIO + WriteIO - lastIOStat.ReadIO - lastIOStat.WriteIO) / interval
		rps := (ReadSectors - lastIOStat.ReadSectors) / interval / 2
		wps := (WriteSectors - lastIOStat.WriteSectors) / interval / 2

		//读取/sys/block/sda/stat文件内容
		statFile := fmt.Sprintf("/sys/block/%s/stat", disk)
		statData, err := ioutil.ReadFile(statFile)
		if err != nil {
			continue
		}

		statReader := bufio.NewReader(bytes.NewBuffer(statData))
		statLine, err := statReader.ReadString('\n')
		if err != nil {
			continue
		}
		fields = strings.Fields(statLine)
		if len(fields) < 11 {
			continue
		}
		Util, _ := strconv.ParseUint(fields[10], 10, 64)
		percent := float64(Util-lastIOStat.Util) / float64(interval) / 1000
		if percent > 100 {
			percent = 100
		}

		lastIOStat.ReadIO = ReadIO
		lastIOStat.ReadSectors = ReadSectors
		lastIOStat.WriteIO = WriteIO
		lastIOStat.WriteSectors = WriteSectors
		lastIOStat.Util = Util
		lastIOStat.Percent = percent

		if len(parText) > 0 {
			parText = fmt.Sprintf("%s\nDisk: %s IOPS: %d Read: %dKB/s Write: %dKB/s %%Util: %.2f%%",
				parText, disk, iops, rps, wps, percent)
		} else {
			parText = fmt.Sprintf("Disk: %s IOPS: %d Read: %dKB/s Write: %dKB/s %%Util: %.2f%%",
				disk, iops, rps, wps, percent)
		}
	}

	if len(parText) > 0 {
		p.Text = parText
	}

	if bWriteXlsx {
		xlsx.SetCellValue("Disk", fmt.Sprintf("A%d", diskXlsxCount+2), time.Now().Format("15:04:05"))
	}

	var diskCount int = 0
	for name, stat := range ioMap {
		if lc, exist := lcs[name]; exist {
			lc.BorderLabel = fmt.Sprintf("IO Used %s %.2f%%", name, stat.Percent)
			procStat := lc.Data[1:]
			procStat = append(procStat, stat.Percent)
			lc.Data = procStat
		}

		if bWriteXlsx {
			xlsx.SetCellValue("Disk", fmt.Sprintf("%c%d", 'B'+diskCount, diskXlsxCount+2), stat.Percent)
			diskCount += 1
		}
	}
	diskXlsxCount += 1
}

func RefreshDiskData(interval uint64) {
	//读取/proc/diskstats文件内容
	bs, err := ioutil.ReadFile("/proc/diskstats")
	if err != nil {
		return
	}

	reader := bufio.NewReader(bytes.NewBuffer(bs))

	for {
		data, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return
		}

		fields := strings.Fields(data)
		if len(fields) < 14 {
			continue
		}

		major, _ := strconv.ParseUint(fields[0], 10, 64)
		minor, _ := strconv.ParseUint(fields[1], 10, 64)
		disk := fields[2]
		if major%8 != 0 || minor%16 != 0 {
			continue
		}

		lastIOStat, exist := ioMap[disk]
		if !exist {
			continue
		}

		//读取/sys/block/sda/stat文件内容
		statFile := fmt.Sprintf("/sys/block/%s/stat", disk)
		statData, err := ioutil.ReadFile(statFile)
		if err != nil {
			continue
		}

		statReader := bufio.NewReader(bytes.NewBuffer(statData))
		statLine, err := statReader.ReadString('\n')
		if err != nil {
			continue
		}
		fields = strings.Fields(statLine)
		if len(fields) < 11 {
			continue
		}
		Util, _ := strconv.ParseUint(fields[10], 10, 64)
		percent := float64(Util-lastIOStat.Util) / float64(interval) / 1000
		if percent > 100 {
			percent = 100
		}

		lastIOStat.Util = Util
		lastIOStat.Percent = percent
	}

	if bWriteXlsx {
		var diskCount int = 0
		xlsx.SetCellValue("Disk", fmt.Sprintf("A%d", diskXlsxCount+2), time.Now().Format("15:04:05"))
		for _, stat := range ioMap {
			xlsx.SetCellValue("Disk", fmt.Sprintf("%c%d", 'B'+diskCount, diskXlsxCount+2), stat.Percent)
			diskCount += 1
		}
		diskXlsxCount += 1
	}
}
