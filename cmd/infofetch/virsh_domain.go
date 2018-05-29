/*
 * Copyright 2017 Red Hat
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
	"context"
	"encoding/xml"
	"fmt"
	"encoding/json"
	libvirt "github.com/libvirt/libvirt-go"
	"log"
)

type Instance struct {
	Name  string `xml:"name"`
	Owner struct {
		User    string `xml:"user"`
		Project string `xml:"project"`
	} `xml:"owner"`
	Flavor struct {
		Name string `xml:"name,attr"`
	} `xml:"flavor"`
}

type Domain struct {
	Name    string `xml:"name"`
	Devices struct {
		Interfaces []struct {
			Type string `xml:"type,attr"`
			Mac  struct {
				Address string `xml:"address,attr"`
			} `xml:"mac"`
			Target struct {
				Dev string `xml:"dev,attr"`
			} `xml:"target"`
		} `xml:"interface"`
	} `xml:"devices"`
}

type OSVMAnnotation struct {
	Name string
	Owner string
	Project string
	Flavor string
}

type OSVMInterfaceAnnotation struct {
	Type string
	MacAddr string
	Target string
	VMName string
}

func parseNovaMetadata(metadata string) (*OSVMAnnotation, error) {
	data := new(Instance)

	if err := xml.Unmarshal([]byte(metadata), data); err != nil {
		log.Println("XML Unmarshal error:", err)
		return nil, err
	}
	log.Printf("name: %s user: %s, project: %s, flavor: %s", data.Name, data.Owner.User, data.Owner.Project, data.Flavor.Name)
	return &OSVMAnnotation{
		Name:		data.Name,
		Owner:		data.Owner.User,
		Project:	data.Owner.Project,
		Flavor:		data.Flavor.Name }, nil
}

func parseXMLForMAC(dumpxml string) (*[]OSVMInterfaceAnnotation, error) {
	data := new(Domain)

	if err := xml.Unmarshal([]byte(dumpxml), data); err != nil {
		log.Println("XML Unmarshal error:", err)
		return nil, err
	}

	ifAnnotation := make([]OSVMInterfaceAnnotation, len(data.Devices.Interfaces))
	for i, v := range data.Devices.Interfaces {
		log.Printf("interface type: %s, mac_addr: %s, target_dev: %s", v.Type, v.Mac.Address, v.Target.Dev)
		ifAnnotation[i] = OSVMInterfaceAnnotation{
			Type: v.Type,
			MacAddr: v.Mac.Address,
			Target: v.Target.Dev,
			VMName: data.Name}
	}
	return &ifAnnotation, nil
}

func setInterfaceAnnotation (ifInfo *[]OSVMInterfaceAnnotation) {
	for _, v := range *ifInfo {
		ifInfoJson, err := json.Marshal(v)
		if err != nil {
			log.Fatalf("err: %v", err)
		}
		log.Printf("byte: %s\n", string(ifInfoJson))
		InfoPool.Set(fmt.Sprintf("if/%s/%s", v.Target, "network"), string(ifInfoJson))
	}
	return
}

func domainEventLifecycleCallback(c *libvirt.Connect,
	d *libvirt.Domain, event *libvirt.DomainEventLifecycle) {
	domName, _ := d.GetName()

	switch event.Event {
	case libvirt.DOMAIN_EVENT_DEFINED:
		// VM defined: vmname (libvirt, nova), user, project, flavor
		// Redis: <vnname>/vminfo
		log.Printf("defined!: domName: %s, event: %v\n", domName, event)
		metadata, err := d.GetMetadata(libvirt.DOMAIN_METADATA_ELEMENT, "http://openstack.org/xmlns/libvirt/nova/1.0", libvirt.DOMAIN_AFFECT_CONFIG)
		if err != nil {
			log.Fatalf("err: %v", err)
		}
		vmInfo, err := parseNovaMetadata(metadata)
		if err != nil {
			log.Fatalf("err: %v", err)
		}
		vmInfoJson, err := json.Marshal(vmInfo)
		if err != nil {
			log.Fatalf("err: %v", err)
		}
		log.Printf("byte: %s\n", string(vmInfoJson))
		InfoPool.Set(fmt.Sprintf("vm/%s/%s", domName, "vminfo"), string(vmInfoJson))
	case libvirt.DOMAIN_EVENT_STARTED:
		// VM started: interface type, interface mac addr, intarface type
		// Redis: <vnname>/interfaces
		log.Printf("started!: domName: %s, event: %v\n", domName, event)

		xml, err := d.GetXMLDesc(0)
		if err != nil {
			log.Fatalf("err: %v", err)
		}
		ifInfo, err := parseXMLForMAC(xml)
		if err != nil {
			log.Fatalf("err: %v", err)
		}
		setInterfaceAnnotation(ifInfo)

		ifInfoJson, err := json.Marshal(ifInfo)
		if err != nil {
			log.Fatalf("err: %v", err)
		}
		log.Printf("byte: %s\n", string(ifInfoJson))
		InfoPool.Set(fmt.Sprintf("vm/%s/%s", domName, "interfaces"), string(ifInfoJson))
	case libvirt.DOMAIN_EVENT_UNDEFINED:
		log.Printf("undefined!: domName: %s, event: %v\n", domName, event)
		vmIFInfo, err := InfoPool.Get(fmt.Sprintf("vm/%s/%s", domName, "interfaces"))
		if err != nil {
			log.Fatalf("err: %v", err)
		} else {
			var interfaces []OSVMInterfaceAnnotation
			err = json.Unmarshal([]byte(vmIFInfo), &interfaces)
			if err != nil {
				log.Fatalf("err: %v", err)
			} else {
				for _, v := range interfaces {
					InfoPool.Del(fmt.Sprintf("if/%s/%s", v.Target, "network"))
				}
			}
		}
		InfoPool.Del(fmt.Sprintf("vm/%s/%s", domName, "vminfo"))
		InfoPool.Del(fmt.Sprintf("vm/%s/%s", domName, "interfaces"))
	default:
		log.Printf("domName: %s, event: %v\n", domName, event)
	}
}

func GetActiveDomain(conn *libvirt.Connect) error {
	doms, err := conn.ListAllDomains(libvirt.CONNECT_LIST_DOMAINS_ACTIVE)
	if err != nil {
		log.Fatalf("libvirt dom list error: %s", err)
		return err
	}

	for _, d := range doms {
		name, err := d.GetName()

		// Get VM Info
		metadata, err := d.GetMetadata(libvirt.DOMAIN_METADATA_ELEMENT, "http://openstack.org/xmlns/libvirt/nova/1.0", libvirt.DOMAIN_AFFECT_CONFIG)
		if err != nil {
			log.Fatalf("err: %v", err)
			return err
		}
		vmInfo, err := parseNovaMetadata(metadata)
		if err != nil {
			log.Fatalf("err: %v", err)
			return err
		}
		vmInfoJson, err := json.Marshal(vmInfo)
		if err != nil {
			log.Fatalf("err: %v", err)
			return err
		}
		log.Printf("byte: %s\n", string(vmInfoJson))
		InfoPool.Set(fmt.Sprintf("vm/%s/%s", name, "vminfo"), string(vmInfoJson))

		// Get Network info
		xml, err := d.GetXMLDesc(0)
		if err != nil {
			log.Fatalf("err: %v", err)
			return err
		}
		ifInfo, err := parseXMLForMAC(xml)
		if err != nil {
			log.Fatalf("err: %v", err)
			return err
		}
		setInterfaceAnnotation(ifInfo)

		ifInfoJson, err := json.Marshal(ifInfo)
		if err != nil {
			log.Fatalf("err: %v", err)
			return err
		}
		log.Printf("byte: %s\n", string(ifInfoJson))
		InfoPool.Set(fmt.Sprintf("vm/%s/%s", name, "interfaces"), string(ifInfoJson))
	}
	return nil
}

func RunVirshEventLoop(ctx context.Context, conn *libvirt.Connect) error {
	callbackId, err := conn.DomainEventLifecycleRegister(nil, domainEventLifecycleCallback)
	if err != nil {
		log.Fatalf("err: callbackid: %d %v", callbackId, err)
	}

	log.Printf("Entering libvirt event loop()")
	for {
		select {
		case <-ctx.Done():
			break
		default:
			if err := libvirt.EventRunDefaultImpl(); err != nil {
				log.Fatalf("%v", err)
			}
		}
	}
	log.Printf("Quitting libvirt event loop()")

	if err := conn.DomainEventDeregister(callbackId); err != nil {
		log.Fatalf("%v", err)
	}
	return nil
}
