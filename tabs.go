package main

import (
	"fmt"

	"github.com/gizak/termui"
	"github.com/gizak/termui/extra"
)

//InitTabs 初始化TermUI界面
func InitTabs(interval, count uint64, pids []string) {
	//CPU
	tab1 := extra.NewTab(" CPU      ")
	cpuOverview := termui.NewPar("")
	cpuOverview.Height = 4
	cpuOverview.Width = 90
	cpuOverview.BorderLabel = "CPU Overview"
	cpuOverview.Text = "User: 0%   System: 0%   Idle: 100%   Wa: 0%   Hi: 0%   Si: 0%   St: 0%"
	tab1.AddBlocks(cpuOverview)

	cpuPercent := termui.NewLineChart()
	cpuPercent.BorderLabel = "CPU Used"
	cpuPercent.Data = make([]float64, 83)
	cpuPercent.Width = 90
	cpuPercent.Height = 30
	cpuPercent.X = 0
	cpuPercent.Y = 5
	cpuPercent.Mode = "dot"
	cpuPercent.DotStyle = 'o'
	tab1.AddBlocks(cpuPercent)

	//Memory
	tab2 := extra.NewTab(" Memory   ")
	memOverview := termui.NewPar("")
	memOverview.Height = 4
	memOverview.Width = 90
	memOverview.BorderLabel = "Memory Overview"
	memOverview.Text = "Mem:  0k total, 0k used, 0k free, 0k buffers\nSwap: 0k total, 0k used, 0k free, 0k cached"
	tab2.AddBlocks(memOverview)

	memPercent := termui.NewLineChart()
	memPercent.BorderLabel = "Memory Used"
	memPercent.Data = make([]float64, 83)
	memPercent.Width = 90
	memPercent.Height = 30
	memPercent.X = 0
	memPercent.Y = 5
	memPercent.Mode = "dot"
	memPercent.DotStyle = 'o'
	tab2.AddBlocks(memPercent)

	//Disk
	tab3 := extra.NewTab(" Disk     ")
	diskCharts := make(map[string]*termui.Gauge)
	mounts := GetDiskMounts()
	for index, mount := range mounts {
		diskPecent := termui.NewGauge()
		diskPecent.Height = 3
		diskPecent.Width = 50
		diskPecent.X = 0
		diskPecent.Y = 3 * index
		diskPecent.BorderLabel = mount
		diskPecent.LabelAlign = termui.AlignRight
		diskCharts[mount] = diskPecent
		tab3.AddBlocks(diskPecent)
	}

	diskOverview := termui.NewPar("")
	diskOverview.Height = 3 + len(ioMap)
	diskOverview.Width = 70
	diskOverview.X = 52
	diskOverview.Y = 0
	diskOverview.BorderLabel = "Disk IO Overview"
	diskOverview.Text = "Disk: sda IOPS: 0 Read: 0KB/s Write: 0KB/s %Util: 0%"
	tab3.AddBlocks(diskOverview)

	position := 0
	ioCharts := make(map[string]*termui.LineChart)
	for disk := range ioMap {
		ioPercent := termui.NewLineChart()
		ioPercent.BorderLabel = "IO Used"
		ioPercent.Data = make([]float64, 63)
		ioPercent.Width = 70
		ioPercent.Height = (3*len(mounts) - len(ioMap) - 4) / len(ioMap)
		ioPercent.X = 52
		ioPercent.Y = 5 + ioPercent.Height*position
		ioPercent.Mode = "dot"
		ioPercent.DotStyle = 'o'
		ioCharts[disk] = ioPercent
		tab3.AddBlocks(ioPercent)
		position++
	}

	//Network
	tab4 := extra.NewTab(" Network  ")
	netOverview := termui.NewPar("")
	netOverview.Height = 3 + len(netMap)
	netOverview.Width = 112
	netOverview.BorderLabel = "Network Overview"
	netOverview.Text = "Name: eth1 IP: 0.0.0.0 Mac: 00:00:00:00:00:00 RXBytes: 0 TXBytes: 0"
	tab4.AddBlocks(netOverview)

	position = 0
	netCharts := make(map[string]*termui.LineChart)
	for net := range netMap {
		netRx := termui.NewLineChart()
		netRx.BorderLabel = "Network RX Speed"
		netRx.Data = make([]float64, 48)
		netRx.Width = 55
		netRx.Height = 20
		netRx.Y = 3 + len(netMap) + netRx.Height*position
		netRx.Mode = "dot"
		netRx.DotStyle = 'o'
		netCharts[fmt.Sprintf("%s-Rx", net)] = netRx
		tab4.AddBlocks(netRx)

		netTx := termui.NewLineChart()
		netTx.BorderLabel = "Network TX Speed"
		netTx.Data = make([]float64, 48)
		netTx.Width = 55
		netTx.Height = 20
		netTx.X = 57
		netTx.Y = 3 + len(netMap) + netTx.Height*position
		netTx.Mode = "dot"
		netTx.DotStyle = 'o'
		netCharts[fmt.Sprintf("%s-Tx", net)] = netTx
		tab4.AddBlocks(netTx)

		position++
	}

	//Loadavg
	tab5 := extra.NewTab(" Loadavg  ")
	loadOverview := termui.NewPar("")
	loadOverview.Height = 4
	loadOverview.Width = 112
	loadOverview.BorderLabel = "Loadavg Overview"
	loadOverview.Text = "Loadavg:   0.00   0.00   0.00   SystemUtil:   0%"
	tab5.AddBlocks(loadOverview)

	loadCharts := make(map[string]*termui.LineChart)
	for i := 0; i < 4; i++ {
		loadavg := termui.NewLineChart()
		loadavg.Data = make([]float64, 48)
		loadavg.Width = 55
		loadavg.Height = 20
		loadavg.X = 57 * (i & 1)
		loadavg.Y = 4 + 10*(i&2)
		loadavg.Mode = "dot"
		loadavg.DotStyle = 'o'
		switch i {
		case 0:
			loadavg.BorderLabel = "loadavg percent"
			loadCharts["percent"] = loadavg
		case 1:
			loadavg.BorderLabel = "1 min loadavg"
			loadCharts["1min"] = loadavg
		case 2:
			loadavg.BorderLabel = "5 min loadavg"
			loadCharts["5min"] = loadavg
		case 3:
			loadavg.BorderLabel = "15 min loadavg"
			loadCharts["15min"] = loadavg
		}
		tab5.AddBlocks(loadavg)
	}

	//Process
	tab6 := extra.NewTab(" Process  ")
	procOverview := termui.NewPar("")
	procCharts := make(map[string]*termui.LineChart)
	if len(pids) > 0 {
		procOverview.Height = 3 + len(pids)
		procOverview.Width = 112
		procOverview.BorderLabel = "Process Overview"
		procOverview.Text = "Pid: 0 User: Virt: 0M Res: 0M %Cpu: 0% %Mem: 0% Cmd:"
		tab6.AddBlocks(procOverview)

		for i, pid := range pids {
			procCPU := termui.NewLineChart()
			procCPU.BorderLabel = "Process CPU Used"
			procCPU.Data = make([]float64, 48)
			procCPU.Width = 55
			procCPU.Height = 20
			procCPU.Y = 3 + len(pids) + procCPU.Height*i
			procCPU.Mode = "dot"
			procCPU.DotStyle = 'o'
			procCharts[fmt.Sprintf("%s-cpu", pid)] = procCPU
			tab6.AddBlocks(procCPU)

			procMem := termui.NewLineChart()
			procMem.BorderLabel = "Process Mem Used"
			procMem.Data = make([]float64, 48)
			procMem.Width = 55
			procMem.Height = 20
			procMem.X = 57
			procMem.Y = 3 + len(pids) + procMem.Height*i
			procMem.Mode = "dot"
			procMem.DotStyle = 'o'
			procCharts[fmt.Sprintf("%s-mem", pid)] = procMem
			tab6.AddBlocks(procMem)
		}
	}

	tabpane := extra.NewTabpane()
	tabpane.Width = 100
	tabpane.Border = true
	tabpane.ActiveTabBg = termui.ColorBlue
	if len(pids) > 0 {
		tabpane.SetTabs(*tab1, *tab2, *tab3, *tab4, *tab5, *tab6)
	} else {
		tabpane.SetTabs(*tab1, *tab2, *tab3, *tab4, *tab5)
	}

	termui.Render(tabpane)

	termui.Handle("/sys/kbd/<left>", func(termui.Event) {
		tabpane.SetActiveLeft()
		termui.Clear()
		termui.Render(tabpane)
	})

	termui.Handle("/sys/kbd/<right>", func(termui.Event) {
		tabpane.SetActiveRight()
		termui.Clear()
		termui.Render(tabpane)
	})

	chs := make(chan bool, 6)
	termui.Handle("/timer/1s", func(e termui.Event) {
		t := e.Data.(termui.EvtTimer)
		if count > 0 && t.Count >= (interval*count) {
			delete(termui.DefaultEvtStream.Handlers, "/timer/1s")
		}

		if t.Count%interval == 0 {
			go RefreshCPUView(cpuOverview, cpuPercent, chs)
			go RefreshMemoryView(memOverview, memPercent, chs)
			go RefreshDiskView(interval, diskCharts, diskOverview, ioCharts, chs)
			go RefreshNetworkView(interval, netOverview, netCharts, chs)
			go RefreshLoadavgView(loadOverview, loadCharts, chs)
			go RefreshProcessView(pids, procOverview, procCharts, chs)

			for i := 0; i < 6; i++ {
				<-chs
			}
			termui.Render(tabpane)
		}
	})
}
