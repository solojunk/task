package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

func InitSilent(interval, count uint64, pids []string) {
	logger.Printf("Run InitSilent!\n")

	if count > 0 {
		for i := 0; i < int(count); i++ {
			time.Sleep(time.Duration(interval) * time.Second)

			RefreshData(interval, pids)
		}

		if bWriteXlsx {
			FinishAndGenerateChart()
		}
		os.Exit(0)

	} else {
		go func() {
			c := make(chan os.Signal)
			signal.Notify(c, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)
			<-c
			if bWriteXlsx {
				FinishAndGenerateChart()
			}
			//logger.Printf("Recive signal %s, process exit!\n", s.String())
			os.Exit(0)
		}()

		for {
			time.Sleep(time.Duration(interval) * time.Second)

			RefreshData(interval, pids)
		}
	}
}

func RefreshData(interval uint64, pids []string) {
	go RefreshCpuData()
	go RefreshMemoryData()
	go RefreshDiskData(interval)
	go RefreshNetworkData(interval)
	go RefreshLoadavgData()
	go RefreshProcessData(pids)
}
