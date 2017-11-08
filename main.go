package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gizak/termui"
)

var logger *log.Logger

func main() {
	logFile, _ := os.OpenFile(fmt.Sprintf("%d.log", os.Getpid()), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	logger = log.New(logFile, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)

	//解析命令行参数
	daemon, silent, interval, count, pids, filename, err := InitFlag()
	if err != nil {
		return
	}

	//结果写入文件
	if len(filename) > 0 {
		InitXlsxFile(pids, filename)
	}

	logger.Printf("daemon=%v, silent=%v, interval=%d, count=%d, pids=%v, filename=%s\n",
		daemon, silent, interval, count, pids, filename)

	if daemon {
		flags := make([]string, 0)
		for _, arg := range os.Args[1:] {
			if arg != "-d" {
				flags = append(flags, arg)
			}
		}
		flags = append(flags, "-s")
		logger.Printf("exec.Command! Args:%v\n", flags)
		cmd := exec.Command(os.Args[0], flags...)
		cmd.Start()
		os.Exit(0)

	} else {
		if silent {
			InitSilent(interval, count, pids)

		} else {
			err = InitTermUi()
			if err != nil {
				logger.Printf("Init TermUi failed! error:%s\n", err.Error())
				return
			}

			InitTabs(interval, count, pids)
		}
	}

	logger.Printf("termui.Loop()\n")
	termui.Loop()
	termui.Close()
}

func InitFlag() (bool, bool, uint64, uint64, []string, string, error) {
	var err error
	var daemon, silent, help bool
	var interval, count, pids, filename string
	flag.BoolVar(&daemon, "d", false, "run as daemon")
	flag.BoolVar(&silent, "s", false, "silent and not show data(do not use it directly)")
	flag.BoolVar(&help, "h", false, "show help")
	flag.StringVar(&interval, "i", "1", "delay time `interval`")
	flag.StringVar(&count, "n", "", "number of `count` limit")
	flag.StringVar(&pids, "p", "", "monitor only processes with specified `pids`")
	flag.StringVar(&filename, "w", "", "write the raw data to `filename`(.xlsx)")
	flag.Usage = usage
	flag.Parse()

	var nDelay uint64
	var nCount uint64
	var arrPids []string

	if help {
		err = flag.ErrHelp
		flag.Usage()
		return daemon, silent, nDelay, nCount, arrPids, filename, err
	}

	//刷新频率
	nDelay, err = strconv.ParseUint(interval, 10, 64)
	if err != nil {
		flag.Usage()
		return daemon, silent, nDelay, nCount, arrPids, filename, err
	}
	if nDelay == 0 {
		flag.Usage()
		return daemon, silent, nDelay, nCount, arrPids, filename, os.ErrInvalid
	}

	//显示次数
	if len(count) > 0 {
		nCount, err = strconv.ParseUint(count, 10, 64)
		if err != nil {
			flag.Usage()
			return daemon, silent, nDelay, nCount, arrPids, filename, err
		}
	}

	//是否指定进程
	if len(pids) > 0 {
		arrPids = strings.Split(pids, ",")
		for _, pid := range arrPids {
			nPid, err := strconv.ParseUint(pid, 10, 64)
			if err != nil {
				flag.Usage()
				return daemon, silent, nDelay, nCount, arrPids, filename, err
			}
			_, err = os.Stat(fmt.Sprintf("/proc/%d", nPid))
			if err != nil {
				fmt.Printf("Pid[%d] does not exist, please check!\n", nPid)
				return daemon, silent, nDelay, nCount, arrPids, filename, err
			}
		}
		go GetFirstProcessData(arrPids)
	}

	//结果写入到文件
	if len(filename) > 0 && !strings.HasSuffix(filename, ".xlsx") {
		filename = fmt.Sprintf("%s.xlsx", filename)
	}

	return daemon, silent, nDelay, nCount, arrPids, filename, err
}

func usage() {
	fmt.Println(`task version: task/1.0.0
Usage: task [-d] [-h] [-i interval] [-n count] [-p pid[,pid ...]] [-w filename]
Options:`)
	flag.PrintDefaults()
	fmt.Printf(`Example:
  task -d -i 1 -n 100
  task -d -i 3 -p 10111,10112
  task -d -i 5 -p 10111,10112 -w result.xlsx

`)
}

func InitTermUi() error {
	if err := termui.Init(); err != nil {
		fmt.Println("Init failed! reason:", err)
		return err
	}

	termui.ColorMap = map[string]termui.Attribute{
		"fg":        termui.ColorWhite,
		"bg":        termui.ColorDefault,
		"border.fg": termui.ColorYellow,
		"label.fg":  termui.ColorRed,
	}

	termui.Handle("/sys/kbd/q", func(termui.Event) {
		termui.StopLoop()

		if bWriteXlsx {
			FinishAndGenerateChart()
		}
	})

	return nil
}
