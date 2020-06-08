package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
)

type CONFIG_MAP struct {
	CurrentConfigFile string
	HostsFromFile     string `json:"hosts-from-file"`
	Hosts             map[string]string
	Bind              string
	ServeFiles        bool     `json:"serve-files"`
	ScriptExts        []string `json:"script-extensions"`
	AbuseHeader       string   `json:"X-Abuse-Info"`
	API_PASSWORD      PASSWORD `json:"api-password"`
	NetRawBlacklist   []string `json:"net-blacklist"`
	NetBlacklist      []net.IPNet
	NetAllowSelf      bool  `json:"net-allow-self"`
	MaxWaitMs         int64 `json:"max-wait-ms"`
	VMTimeoutMs       int64 `json:"vm-timeout-ms"`
}

var CONFIG = CONFIG_MAP{
	Hosts:      make(map[string]string),
	ScriptExts: make([]string, 0),
}

func Load(file string) {
	CONFIG.CurrentConfigFile = file
	Reload()
}

func Reload() {
	jsonFile, err := os.Open(CONFIG.CurrentConfigFile)
	checkErr(err)

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	checkErr(err)
	json.Unmarshal(byteValue, &CONFIG)

	// Convert NetRawBlacklist entries to NetBlacklist
	for _, cidr := range CONFIG.NetRawBlacklist {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			fmt.Println("Failed to parse net-blacklist cider: ", cidr)
		} else {
			CONFIG.NetBlacklist = append(CONFIG.NetBlacklist, *ipnet)
		}
	}

	// Make it possible to load hosts from a file
	if len(CONFIG.HostsFromFile) > 0 {
		hostsFile, err := os.Open(CONFIG.HostsFromFile)
		checkErr(err)
		defer hostsFile.Close()

		byteValue, err := ioutil.ReadAll(hostsFile)
		checkErr(err)

		CONFIG.Hosts = make(map[string]string)
		json.Unmarshal(byteValue, &CONFIG.Hosts)
	}

	// Default timeout of 15s for script execution
	if CONFIG.VMTimeoutMs == 0 {
		CONFIG.VMTimeoutMs = 15000
	}

	// Default max wait ms if not defined
	if CONFIG.MaxWaitMs == 0 {
		CONFIG.MaxWaitMs = 5000
	}
}
