package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var (
	collectdConfig = flag.String("c", "/tmp/collectd.conf", "collectd config file")
)

func main() {
	outStatus, errStatus := exec.Command("sudo", "systemctl", "status", "collectd").Output()
	if errStatus != nil {
		fmt.Printf("status NG\n")
		os.Exit(1)
	}
	if !strings.Contains(string(outStatus), "running") {
		fmt.Printf("status not running\n")
		os.Exit(1)
	}

	_, errStop := exec.Command("sudo", "systemctl", "stop", "collectd").Output()
	if errStop != nil {
		fmt.Printf("stop NG\n")
		os.Exit(1)
	}

	flag.Parse()
	if _, errFile := os.Stat(*collectdConfig); errFile != nil {
		fmt.Printf("file not found\n")
		os.Exit(1)
	}
	errCp := exec.Command("sudo", "cp", *collectdConfig, "/etc/collectd/collectd.conf.d/localagent.conf").Run()
	if errCp != nil {
		fmt.Printf("cp NG\n")
		os.Exit(1)
	}

	_, errStart := exec.Command("sudo", "systemctl", "start", "collectd").Output()
	if errStart != nil {
		fmt.Printf("start NG\n")
		os.Exit(1)
	}

	fmt.Printf("All complete!\n")
}
