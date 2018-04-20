/*
 * Copyright 2018 NEC Corporation
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package main

import (
	"github.com/distributed-monitoring/agent/pkg/annotate"
	libvirt "github.com/libvirt/libvirt-go"
	"log"
)

var testVal = 0

func writeInfo(infoPool annotate.Pool) {
	conn, err := libvirt.NewConnect("qemu:///system")
	if err != nil {
		log.Fatalln("libvirt connect error")
	}
	defer conn.Close()

	doms, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
	if err != nil {
		log.Fatalln("libvirt command error")
	}

	log.Printf("%d running domains:\n", len(doms))

	for _, dom := range doms {
		name, err := dom.GetName()
		if err != nil {
			log.Fatalf("virt GetName error: %s", err)
		}
		dom.Free()
		switch testVal {
		case 0:
			infoPool.Set("server", name, "{\"addinfo\": \"somevalue\"}")
		case 1, 3:
			log.Println(infoPool.Get("server", name))
		case 2:
			infoPool.Del("server", name)
		}
	}
	testVal = (testVal + 1) % 4
}
