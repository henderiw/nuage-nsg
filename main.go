package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/henderiw/nuagewim"
	"github.com/nuagenetworks/go-bambou/bambou"
	"github.com/nuagenetworks/vspk-go/vspk"
	"github.com/sirupsen/logrus"
)

// VSD crednetials
var vsdURL = "https://138.203.39.87:8443"
var vsdUser = "csproot"
var vsdPass = "csproot"
var vsdEnterprise = "csp"

// Usr is a user
var Usr *vspk.Me

// NSG imported Configuration
var nsgCfg nuagewim.NuageNSGCfg

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	var s *bambou.Session
	s, Usr = vspk.NewSession(vsdUser, vsdPass, vsdEnterprise, vsdURL)
	if err := s.Start(); err != nil {
		fmt.Println("Unable to connect to Nuage VSD: " + err.Description)
		os.Exit(1)
	}

	//find enterprise
	enterpriseCfg := map[string]interface{}{
		"Name": "vspkPublicNonDpdk",
	}

	enterprise := nuagewim.NuageEnterprise(enterpriseCfg, Usr)
	fmt.Println(enterprise)

	nsgFiles, err := ioutil.ReadDir("nsgs")
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, nsgFile := range nsgFiles {
		//read NSG Template information
		rcvdNSGData, readErr := ioutil.ReadFile("nsgs/" + nsgFile.Name())
		if readErr != nil {
			log.Fatal(readErr)
		}
		fmt.Printf("File contents: %s", rcvdNSGData)

		// unmarshal (deserialize) the json and save the result in the struct &cfg
		err := json.Unmarshal([]byte(rcvdNSGData), &nsgCfg)
		if err != nil {
			log.Fatal(err)
		}

		nsGatewayTemplateCfg := map[string]interface{}{
			"Name": nsgCfg.NSGTemplateName,
		}
		//fmt.Printf("NSG Template ID: %s \n", nsgCfg.NSGTemplateID)

		nsGatewayTemplate := nuagewim.NuageNSGatewayTemplate(nsGatewayTemplateCfg, Usr)
		nsgCfg.NSGTemplateID = nsGatewayTemplate.ID
		//fmt.Printf("NSG Template ID: %s \n", nsgCfg.NSGTemplateID)

		for i, networkPort := range nsgCfg.NetworkPorts {
			fmt.Printf("NSG Network Port %d Name: %s \n", i, networkPort.Name)
			vscProfCfg := map[string]interface{}{
				"Name": networkPort.VscName,
			}
			vscProf := nuagewim.NuageInfraVSCProf(vscProfCfg, Usr)
			nsgCfg.NetworkPorts[i].VscID = vscProf.ID

			if networkPort.UnderlayName != "" {
				underlayCfg := map[string]interface{}{
					"Name": networkPort.UnderlayName,
				}
				underlay := nuagewim.NuageUnderlay(underlayCfg, Usr)
				nsgCfg.NetworkPorts[i].UnderlayID = underlay.ID
			}
		}

		fmt.Printf("NSG Config: %s \n", nsgCfg)
		nsg1 := nuagewim.NuageCreateEntireNSG(nsgCfg, enterprise)
		fmt.Println(nsg1)
	}

}
