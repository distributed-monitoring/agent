package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func createCollectdConf() error {
	outStatus, errStatus := exec.Command("sudo", "systemctl", "status", "collectd").Output()
	if errStatus != nil {
		return fmt.Errorf("status NG")
	}
	if !strings.Contains(string(outStatus), "running") {
		return fmt.Errorf("status not running")
	}

	_, errStop := exec.Command("sudo", "systemctl", "stop", "collectd").Output()
	if errStop != nil {
		return fmt.Errorf("stop NG")
	}

	_, errStart := exec.Command("sudo", "systemctl", "start", "collectd").Output()
	if errStart != nil {
		return fmt.Errorf("start NG")
	}

	fmt.Println("All complete!")

	return nil
}
