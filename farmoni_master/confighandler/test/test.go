// Proof of Concepts for the Cloud-Barista Multi-Cloud Project.
//      * Cloud-Barista: https://github.com/cloud-barista
//
// test for configuration handling with YAML.
// ref) https://github.com/go-yaml/yaml/tree/v3
//
// by powerkim@powerkim.co.kr, 2019.04.

package main

import (
        "fmt"

	"github.com/powerkimhub/farmoni/farmoni_master/confighandler"
)

func main() {
	masterConfigInfos := confighandler.GetMasterConfigInfos()

        fmt.Printf("\n<unmarshalled config values>\n")
        fmt.Printf("%v\n\n", masterConfigInfos)

 
	fmt.Printf("  %s\n", masterConfigInfos.ETCDSERVERPORT)

	fmt.Printf("\t%s\n", masterConfigInfos.AWS.KEYFILEPATH)
	fmt.Printf("\t%s\n", masterConfigInfos.AWS.REGION)
	fmt.Printf("\t%s\n", masterConfigInfos.AWS.IMAGEID)
	fmt.Printf("\t%s\n", masterConfigInfos.AWS.INSTANCENAMEPREFIX)
	fmt.Printf("\t%s\n", masterConfigInfos.AWS.INSTANCETYPE)
	fmt.Printf("\t%s\n", masterConfigInfos.AWS.KEYNAME)
	fmt.Printf("\t%s\n", masterConfigInfos.AWS.SECURITYGROUPID)
	fmt.Printf("\t%s\n\n", masterConfigInfos.AWS.SUBNETID)
/*

	d, err := yaml.Marshal(&t)
        if err != nil {
                log.Fatalf("error: %v", err)
        }
        fmt.Printf("<marshaled data>\n")
        fmt.Printf("%s\n\n", string(d))
*/
}
