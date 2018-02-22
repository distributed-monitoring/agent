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

func create_collectd_conf() int {
	outStatus, errStatus := exec.Command("sudo", "systemctl", "status", "collectd").Output()
	if errStatus != nil {
		fmt.Println("status NG")
		return 1
	}
	if !strings.Contains(string(outStatus), "running") {
		fmt.Println("status not running")
		return 1
	}

	_, errStop := exec.Command("sudo", "systemctl", "stop", "collectd").Output()
	if errStop != nil {
		fmt.Println("stop NG")
		return 1
	}

	flag.Parse()
	if _, errFile := os.Stat(*collectdConfig); errFile != nil {
		fmt.Println("file not found")
		return 1
	}
	errCp := exec.Command("sudo", "cp", *collectdConfig, "/etc/collectd/collectd.conf.d/localagent.conf").Run()
	if errCp != nil {
		fmt.Println("cp NG")
		return 1
	}

	_, errStart := exec.Command("sudo", "systemctl", "start", "collectd").Output()
	if errStart != nil {
		fmt.Println("start NG")
		return 1
	}

	fmt.Println("All complete!")

	return 0
}
