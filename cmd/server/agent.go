package main

import (
	"flag"
	"fmt"
	// "os"
	"os/exec"
	"strings"
)

var (
	collectdConfig = flag.String("c", "/tmp/collectd.conf", "collectd config file")
)

func createCollectdConf() error {
	outStatus, errStatus := exec.Command("sudo", "systemctl", "status", "collectd").Output()
	if errStatus != nil {
		return fmt.Errorf("Status NG")
	}
	if !strings.Contains(string(outStatus), "running") {
		return fmt.Errorf("Status not running")
	}

	_, errStop := exec.Command("sudo", "systemctl", "stop", "collectd").Output()
	if errStop != nil {
		return fmt.Errorf("Stop NG")
	}

	/*
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
	*/

	_, errStart := exec.Command("sudo", "systemctl", "start", "collectd").Output()
	if errStart != nil {
		return fmt.Errorf("Start NG")
	}

	fmt.Println("All complete!")

	return nil
}
