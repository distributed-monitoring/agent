package main

import (
	"flag"
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
		os.Exit(1)
	}
	if !strings.Contains(string(outStatus), "running") {
		os.Exit(1)
	}

	_, errStop := exec.Command("sudo", "systemctl", "stop", "collectd").Output()
	if errStop != nil {
		os.Exit(1)
	}

	flag.Parse()
	if _, errFile := os.Stat(*collectdConfig); errFile != nil {
		os.Exit(1)
	}
	errCp := exec.Command("sudo", "cp", *collectdConfig, "/etc/collectd/collectd.conf").Run()
	if errCp != nil {
		os.Exit(1)
	}

	_, errStart := exec.Command("sudo", "systemctl", "start", "collectd").Output()
	if errStart != nil {
		os.Exit(1)
	}
}
