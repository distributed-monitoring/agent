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
	"bytes"
	"net/http"
	"net/url"
	"os"
	"log"
	"strings"
	"fmt"
	"time"
	"context"
	"io/ioutil"
	"text/template"
	"encoding/json"
)

type UserInfo struct {
	UserDomainName string
	UserName string
	Password string
	ProjectDomainName string
	ProjectName string
}

var token_json_template string = `{
  "auth": {
    "identity": {
      "methods": [
        "password"
      ],
      "password": {
        "user": {
          "domain": {
            "name": "{{.UserDomainName}}"
          },
          "name": "{{.UserName}}",
          "password": "{{.Password}}"
        }
      }
    },
    "scope": {
      "project": {
        "domain": {
          "name": "{{.ProjectDomainName}}"
        },
        "name": "{{.ProjectName}}"
      }
    }
  }
}
`

type TokenReply struct {
	Token struct {
		IsDomain bool     `json:"is_domain"`
		Methods  []string `json:"methods"`
		Roles    []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"roles"`
		ExpiresAt time.Time `json:"expires_at"`
		Project   struct {
			Domain struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"domain"`
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"project"`
		User struct {
			PasswordExpiresAt interface{} `json:"password_expires_at"`
			Domain            struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"domain"`
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"user"`
		AuditIds []string  `json:"audit_ids"`
		IssuedAt time.Time `json:"issued_at"`
	} `json:"token"`
}

type Token struct {
	Token string
	ExpiresAt time.Time
}

func (t *Token) CheckToken() {
	now := time.Now()

	if t.ExpiresAt.Sub(now).Seconds() < 30 {
		newToken, _ := GetToken()
		t.Token = newToken.Token
		t.ExpiresAt = newToken.ExpiresAt
	}
}

func GetToken () (*Token, error) {
	var buf bytes.Buffer

	t := template.Must(template.New("json template1").Parse(token_json_template))
	p := UserInfo {
		UserDomainName: os.ExpandEnv("$OS_USER_DOMAIN_NAME"),
		UserName: os.ExpandEnv("$OS_USERNAME"),
		Password: os.ExpandEnv("$OS_PASSWORD"),
		ProjectDomainName: os.ExpandEnv("$OS_PROJECT_DOMAIN_NAME"),
		ProjectName: os.ExpandEnv("$OS_PROJECT_NAME"),
	}
	t.Execute(&buf, p)

	body := bytes.NewReader(buf.Bytes())
	req, err := http.NewRequest("POST", os.ExpandEnv("$OS_AUTH_URL/auth/tokens?nocatalog"), body)
	if err != nil {
		return &Token{"", time.Unix(0, 0)}, fmt.Errorf("Http request failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &Token{"", time.Unix(0, 0)}, fmt.Errorf("Http POST failed: %v", err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)

	tokenStr, ok := resp.Header["X-Subject-Token"]
	if ok != true && len(tokenStr) != 1 {
		return &Token{"", time.Unix(0, 0)}, fmt.Errorf("no token in openstack reply")
	}

	var repl TokenReply
	err = json.Unmarshal(b, &repl)

	return &Token{tokenStr[0], repl.Token.ExpiresAt}, nil
}


type Service struct {
	Description string `json:"description"`
	Links       struct {
		Self string `json:"self"`
	} `json:"links"`
	Enabled bool   `json:"enabled"`
	Type    string `json:"type"`
	ID      string `json:"id"`
	Name    string `json:"name"`
}

type ServiceListReply struct {
	Services []Service `json:"services"`
}

func (s *ServiceListReply) GetService (name string) (*Service, error) {
	for _, v:= range s.Services {
		if v.Name == name {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("No service id (%s) found", name)
}

type EndPoint struct {
	RegionID string `json:"region_id"`
	Links    struct {
		Self string `json:"self"`
	} `json:"links"`
	URL       string `json:"url"`
	Region    string `json:"region"`
	Enabled   bool   `json:"enabled"`
	Interface string `json:"interface"`
	ServiceID string `json:"service_id"`
	ID        string `json:"id"`
}

type EndPointReply struct {
	Endpoints []EndPoint `json:"endpoints"`
}

func (e *EndPointReply) GetEndpoint (serviceid string, ifname string) (*EndPoint, error) {
	for _, v := range e.Endpoints {
		if v.Interface == ifname && v.ServiceID == serviceid {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("no endpoint found (%s/%s)", serviceid, ifname)
}

func GetEndpoints (token *Token) (EndPointReply, error) {
	token.CheckToken()
	req, err := http.NewRequest("GET", os.ExpandEnv("$OS_AUTH_URL/endpoints"), nil)
	if err != nil {
		return EndPointReply{}, fmt.Errorf("Request failed:%v", err)
	}
	req.Header.Set("X-Auth-Token", token.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return EndPointReply{}, fmt.Errorf("http GET failed:%v", err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	//fmt.Printf("%s", string(b))

	var repl EndPointReply
	err = json.Unmarshal(b, &repl)
	if err != nil {
		return EndPointReply{}, fmt.Errorf("http reply decoding failed:%v", err)
	}
	//fmt.Printf("%v", repl)
	return repl, nil
}

func GetServiceList (token *Token) (ServiceListReply, error) {
	token.CheckToken()
	req, err := http.NewRequest("GET", os.ExpandEnv("$OS_AUTH_URL/services"), nil)
	if err != nil {
		return ServiceListReply{}, fmt.Errorf("Request failed:%v", err)
	}
	req.Header.Set("X-Auth-Token", token.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ServiceListReply{}, fmt.Errorf("http GET failed:%v", err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)

	var repl ServiceListReply
	err = json.Unmarshal(b, &repl)
	if err != nil {
		return ServiceListReply{}, fmt.Errorf("http reply decoding failed:%v", err)
	}
	return repl, nil
}

type NeutronPort struct {
	AllowedAddressPairs []interface{} `json:"allowed_address_pairs"`
	ExtraDhcpOpts       []interface{} `json:"extra_dhcp_opts"`
	UpdatedAt           time.Time     `json:"updated_at"`
	DeviceOwner         string        `json:"device_owner"`
	RevisionNumber      int           `json:"revision_number"`
	PortSecurityEnabled bool          `json:"port_security_enabled"`
	BindingProfile      struct {
	} `json:"binding:profile"`
	FixedIps []struct {
		SubnetID  string `json:"subnet_id"`
		IPAddress string `json:"ip_address"`
	} `json:"fixed_ips"`
	ID                string        `json:"id"`
	SecurityGroups    []interface{} `json:"security_groups"`
	BindingVifDetails struct {
		PortFilter    bool   `json:"port_filter"`
		DatapathType  string `json:"datapath_type"`
		OvsHybridPlug bool   `json:"ovs_hybrid_plug"`
	} `json:"binding:vif_details"`
	BindingVifType  string        `json:"binding:vif_type"`
	MacAddress      string        `json:"mac_address"`
	ProjectID       string        `json:"project_id"`
	Status          string        `json:"status"`
	BindingHostID   string        `json:"binding:host_id"`
	Description     string        `json:"description"`
	Tags            []interface{} `json:"tags"`
	QosPolicyID     interface{}   `json:"qos_policy_id"`
	Name            string        `json:"name"`
	AdminStateUp    bool          `json:"admin_state_up"`
	NetworkID       string        `json:"network_id"`
	TenantID        string        `json:"tenant_id"`
	CreatedAt       time.Time     `json:"created_at"`
	BindingVnicType string        `json:"binding:vnic_type"`
	DeviceID        string        `json:"device_id"`
}

type NeutronPortReply struct {
	Ports []NeutronPort `json:"ports"`
}

func GetNeutronPorts (token *Token, endpoint string) (NeutronPortReply, error) {
	token.CheckToken()
	req, err := http.NewRequest("GET", endpoint+"/v2.0/ports", nil)
	if err != nil {
		return NeutronPortReply{}, fmt.Errorf("Request failed:%v", err)
	}
	req.Header.Set("X-Auth-Token", token.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return NeutronPortReply{}, fmt.Errorf("http GET failed:%v", err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)

	var repl NeutronPortReply
	err = json.Unmarshal(b, &repl)
	if err != nil {
		return NeutronPortReply{}, fmt.Errorf("http reply decoding failed:%v", err)
	}
	return repl, nil
}

func (n *NeutronPortReply) GetNeutronPortfromMAC (mac string) (*NeutronPort,
	error) {
	for _, v:= range n.Ports {
		if v.MacAddress == strings.ToLower(mac) {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("No port (%s) found", mac)
}

type NeutronNetwork struct {
	ProviderPhysicalNetwork string        `json:"provider:physical_network"`
	Ipv6AddressScope        interface{}   `json:"ipv6_address_scope"`
	RevisionNumber          int           `json:"revision_number"`
	PortSecurityEnabled     bool          `json:"port_security_enabled"`
	Mtu                     int           `json:"mtu"`
	ID                      string        `json:"id"`
	RouterExternal          bool          `json:"router:external"`
	AvailabilityZoneHints   []interface{} `json:"availability_zone_hints"`
	AvailabilityZones       []string      `json:"availability_zones"`
	ProviderSegmentationID  interface{}   `json:"provider:segmentation_id"`
	Ipv4AddressScope        interface{}   `json:"ipv4_address_scope"`
	Shared                  bool          `json:"shared"`
	ProjectID               string        `json:"project_id"`
	Status                  string        `json:"status"`
	Subnets                 []string      `json:"subnets"`
	Description             string        `json:"description"`
	Tags                    []interface{} `json:"tags"`
	UpdatedAt               time.Time     `json:"updated_at"`
	IsDefault               bool          `json:"is_default"`
	QosPolicyID             interface{}   `json:"qos_policy_id"`
	Name                    string        `json:"name"`
	AdminStateUp            bool          `json:"admin_state_up"`
	TenantID                string        `json:"tenant_id"`
	CreatedAt               time.Time     `json:"created_at"`
	ProviderNetworkType     string        `json:"provider:network_type"`
}

type NeutronNetworkReply struct {
	Networks []NeutronNetwork `json:"networks"`
}

func (n *NeutronNetworkReply) GetNetworkFromID (netid string) (*NeutronNetwork, error) {
	for _, v:= range n.Networks {
		if v.ID == netid {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("No network (%s) found", netid)
}

func GetNetworkReply (token *Token, endpoint string) (NeutronNetworkReply, error) {
	token.CheckToken()
	req, err := http.NewRequest("GET", endpoint+"/v2.0/networks", nil)
	if err != nil {
		return NeutronNetworkReply{}, fmt.Errorf("Request failed:%v", err)
	}
	req.Header.Set("X-Auth-Token", token.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return NeutronNetworkReply{}, fmt.Errorf("http GET failed:%v", err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)

	var repl NeutronNetworkReply
	err = json.Unmarshal(b, &repl)
	if err != nil {
		return NeutronNetworkReply{}, fmt.Errorf("http reply decoding failed:%v", err)
	}
	return repl, nil
}

type NovaCompute struct {
	ID    string `json:"id"`
	Links []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`
	Name string `json:"name"`
}

type NovaComputeReply struct {
	Servers []NovaCompute `json:"servers"`
}

func (n *NovaComputeReply) GetComputeFromID (vmid string) (*NovaCompute, error) {
	for _, v:= range n.Servers {
		if v.ID == vmid {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("No vm (%s) found", vmid)
}

func GetComputeReply (token *Token, endpoint string) (NovaComputeReply, error) {
	token.CheckToken()
	values := url.Values{}
	values.Add("all_tenants", "1")

	req, err := http.NewRequest("GET", endpoint+"/servers", nil)
	if err != nil {
		return NovaComputeReply{}, fmt.Errorf("Request failed:%v", err)
	}
	req.Header.Set("X-Auth-Token", token.Token)
	req.URL.RawQuery = values.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return NovaComputeReply{}, fmt.Errorf("http GET failed:%v", err)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)

	var repl NovaComputeReply
	err = json.Unmarshal(b, &repl)
	if err != nil {
		return NovaComputeReply{}, fmt.Errorf("http reply decoding failed:%v", err)
	}

	return repl, nil
}

type OSNeutronInterfaceAnnotation struct {
	IfName string
	VMName string
	NetworkName string
}

func RunNeutronInfoFetch(ctx context.Context, vmIfInfo chan string) error {
	token, err := GetToken()

	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot get token: %v\n", err)
		return err
	}

	svc, _ := GetServiceList(token)
	neuId, _ := svc.GetService("neutron")
	//fmt.Printf("neutron id:%s\n", id.ID)

	novaId, _ := svc.GetService("nova")
	//fmt.Printf("nova id:%s\n", id.ID)

	endpoints, _ := GetEndpoints(token)
	neuEndpoint, _ := endpoints.GetEndpoint(neuId.ID, "admin")
	//fmt.Printf("neutron endpoint:%s\n", neuEndpoint.URL)

	novaEndpoint, _ := endpoints.GetEndpoint(novaId.ID, "admin")
	//fmt.Printf("nova endpoint:%s\n", novaEndpoint.URL)

	GetComputeReply(token, novaEndpoint.URL)
	GetNeutronPorts(token, neuEndpoint.URL)
	//vmrepl, _ := GetComputeReply(token, novaEndpoint.URL)
	//prepl, _ := GetNeutronPorts(token, neuEndpoint.URL)

	EVENTLOOP:
	for {
		select {
		case <-ctx.Done():
			break EVENTLOOP
		case key := <-vmIfInfo:
			log.Printf("incoming IF: %v", key)
			libvirtIfInfo, err := InfoPool.Get(key)
			if err != nil {
				log.Fatalf("err: %v", err)
			} else {
				var ifInfo OSVMInterfaceAnnotation
				err = json.Unmarshal([]byte(libvirtIfInfo), &ifInfo)
				if err != nil {
					log.Fatalf("err: %v", err)
				} else {
					vmrepl, _ := GetComputeReply(token, novaEndpoint.URL)
					prepl, _ := GetNeutronPorts(token, neuEndpoint.URL)
					nrepl, _ := GetNetworkReply(token, neuEndpoint.URL)
					netid, _ := prepl.GetNeutronPortfromMAC(ifInfo.MacAddr)
					net, _ := nrepl.GetNetworkFromID(netid.NetworkID)
					vm, _ := vmrepl.GetComputeFromID(netid.DeviceID)
					osIfInfo := OSNeutronInterfaceAnnotation{
						IfName: ifInfo.Target,
						VMName: vm.Name,
						NetworkName: net.Name }

					osIfInfoJson, err := json.Marshal(osIfInfo)
					if err != nil {
						log.Fatalf("err: %v", err)
					} else {
						log.Printf("vmname: %s / networkname:%s\n", vm.Name, net.Name)
						InfoPool.Set(fmt.Sprintf("if/%s/%s", ifInfo.Target, "neutron_network"), string(osIfInfoJson))
					}
				}
			}
		}
	}
	return nil
}

/*
func main() {
	token, err := GetToken()

	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot get token: %v\n", err)
		return
	}

	svc, _ := GetServiceList(token)
	neuId, _ := svc.GetService("neutron")
	//fmt.Printf("neutron id:%s\n", id.ID)

	novaId, _ := svc.GetService("nova")
	//fmt.Printf("nova id:%s\n", id.ID)

	endpoints, _ := GetEndpoints(token)
	neuEndpoint, _ := endpoints.GetEndpoint(neuId.ID, "admin")
	//fmt.Printf("neutron endpoint:%s\n", neuEndpoint.URL)

	novaEndpoint, _ := endpoints.GetEndpoint(novaId.ID, "admin")
	//fmt.Printf("nova endpoint:%s\n", novaEndpoint.URL)

	vmrepl, _ := GetComputeReply(token, novaEndpoint.URL)

	prepl, _ := GetNeutronPorts(token, neuEndpoint.URL)
	netid, _ := prepl.GetNeutronPortfromMAC("fa:16:3e:1e:04:08")
	fmt.Printf("netid:%s\ndeviceid:%s\n", netid.NetworkID, netid.DeviceID)

	nrepl, _ := GetNetworkReply(token, neuEndpoint.URL)
	net, _ := nrepl.GetNetworkFromID(netid.NetworkID)
	vm, _ := vmrepl.GetComputeFromID(netid.DeviceID)
	fmt.Printf("vmname: %s / networkname:%s\n", vm.Name, net.Name)

}
*/
