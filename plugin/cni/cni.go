package cni

import (
	"encoding/json"
	"fmt"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
	bv "github.com/containernetworking/plugins/pkg/utils/buildversion"
	"github.com/vishvananda/netlink"
	"runtime"
)

//2021年5月17日
type NetConf struct {
	types.NetConf
	Master string `json:"master"`
	Mode   string `json:"Mode"`
	MTU    int    `json:"mtu"`
}

func init() {
	runtime.LockOSThread()
}
func main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, bv.BuildString("ipvlan"))
}
func loadConf(bytes []byte, cmdCheck bool) (*NetConf, string, error) {
	n := &NetConf{}
	if err := json.Unmarshal(bytes, n); err != nil {
		return nil, "", fmt.Errorf("failed to load netconf: %v", err)
	}

	if cmdCheck {
		return n, n.CNIVersion, nil
	}

	var result *current.Result
	var err error
	// Parse previous result
	if n.NetConf.RawPrevResult != nil {
		if err = version.ParsePrevResult(&n.NetConf); err != nil {
			return nil, "", fmt.Errorf("could not parse prevResult: %v", err)
		}

		result, err = current.NewResultFromResult(n.PrevResult)
		if err != nil {
			return nil, "", fmt.Errorf("could not convert result to current version: %v", err)
		}
	}
	if n.Master == "" {
		if result == nil {
			defaultRouteInterface, err := getDefaultRouteInterfaceName()

			if err != nil {
				return nil, "", err
			}
			n.Master = defaultRouteInterface
		} else {
			if len(result.Interfaces) == 1 && result.Interfaces[0].Name != "" {
				n.Master = result.Interfaces[0].Name
			} else {
				return nil, "", fmt.Errorf("chained master failure. PrevResult lacks a single named interface")
			}
		}
	}
	return n, n.CNIVersion, nil
}

func getDefaultRouteInterfaceName() (string, error) {
	routeToDstIP, err := netlink.RouteList(nil, netlink.FAMILY_ALL)
	if err != nil {
		return "", err
	}

	for _, v := range routeToDstIP {
		if v.Dst == nil {
			l, err := netlink.LinkByIndex(v.LinkIndex)
			if err != nil {
				return "", err
			}
			return l.Attrs().Name, nil
		}
	}

	return "", fmt.Errorf("no default route interface found")
}
func cmdAdd(args *skel.CmdArgs) error {
	n, cniVersion, err := loadConf(args.StdinData, false)
	if err != nil {
		return err
	}
	return nil
}

func cmdDel(args *skel.CmdArgs) error {
	return nil
}

func cmdCheck(args *skel.CmdArgs) error {
	return nil
}
