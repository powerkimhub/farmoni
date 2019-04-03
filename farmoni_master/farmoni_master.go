// Proof of Concepts for the Cloud-Barista Multi-Cloud Project.
//      * Cloud-Barista: https://github.com/cloud-barista
//
// Farmoni Master to control server and fetch monitoring info.
//
// by powerkim@powerkim.co.kr, 2019.03.
package main

import (
	"flag"
	"github.com/powerkimhub/farmoni/farmoni_master/ec2handler"
	"github.com/powerkimhub/farmoni/farmoni_master/serverhandler/scp"
	"github.com/powerkimhub/farmoni/farmoni_master/serverhandler/sshrun"
	"github.com/powerkimhub/farmoni/farmoni_master/etcdhandler"
	"fmt"
	"os"
	"context"
	"time"
	"log"
	"strconv"
        "google.golang.org/grpc"
        pb "github.com/powerkimhub/farmoni/grpc_def"
)

const (
	defaultServerName = "129.254.184.79"
	port     = "2019"
)

var etcdServerPort *string
var fetchType *string

var addServer *string
var delServer *string

var addServerNumAWS *int
var delServerNumAWS *int

var addServerNumGCP *int
var delServerNumGCP *int

var serverlist *bool
var monitoring *bool
var delallServersAWS *bool


func parseRequest() {

        etcdServerPort = flag.String("etcdserver", "129.254.175.43:2379", "etcdserver=129.254.175.43:2379")
        fetchType = flag.String("fetchtype", "PULL", "fetch type: -fetchtype=PUSH")
/*
        addServer = flag.String("addserver", "none", "add a server: -addserver=192.168.0.10:5000")
        delServer = flag.String("delserver", "none", "delete a server: -delserver=192.168.0.10")
*/
        addServerNumAWS = flag.Int("addserversaws", 0, "add servers in AWS: -addserversaws=10")
        delallServersAWS = flag.Bool("delallserversaws", false, "delete all servers in AWS: -delallserversaws")

        addServerNumGCP = flag.Int("addserversgcp", 0, "add servers in GCP: -addserversgcp=10")
//       delServersNumGCP = flag.Int("delserversgcp", 0, "delete all servers in GCP: -delserversgcp=10")

        serverlist = flag.Bool("serverlist", false, "report server list: -serverlist")
        monitoring = flag.Bool("monitoring", false, "report all server' resources status: -monitoring")

        flag.Parse()
}

// 1. setup a credential info of AWS.
// 2. setup a keypair for VM ssh login.

// 1. parsing user's request.

//<add Servers in AWS/GCP>
// 1.1. create Servers(VM).
// 1.2. get servers' public IP.
// 1.3. insert Farmoni Agent into Servers.
// 1.4. execute Servers' Agent.
// 1.5. add server list into etcd.

//<get all server list>
// 2.1. get server list from etcd.
// 2.2. fetch all agent's monitoring info.

//<delete all servers inAWS/GCP>
func main() {

        fmt.Println("$ go run farmoni_master.go -addserversaws=10")
        fmt.Println("$ go run farmoni_master.go -monitoring")
        fmt.Println("$ go run farmoni_master.go -addserversgcp=5")

	// to delete all servers in aws
        fmt.Println("")
        fmt.Println("$ go run farmoni_master.go -delallserversaws")
        //fmt.Println("$ go run farmoni_master.go -delallserversgcp")
        fmt.Println("")
        fmt.Println("$ go run farmoni_master.go -serverlist")
        fmt.Println("")


 // dedicated option for PoC
	// 1. parsing user's request.
	parseRequest()

//<add servers in AWS/GCP>
	// 1.1. create Servers(VM).
	if *addServerNumAWS != 0 {
                fmt.Println("######### addServersAWS....")
                addServersAWS(*addServerNumAWS)
        }
	if *addServerNumGCP != 0 {
                fmt.Println("######### addServersGCP....")
                addServersGCP(*addServerNumGCP)
        }
//<get all server list>
	if *serverlist != false {
                //fmt.Println("######### list of all servers....")
                serverList()
        }
// 2.2. fetch all agent's monitoring info.
	if *monitoring != false {
                fmt.Println("######### monitoring all servers....")
                monitoringAll()
        }

//<delete all servers inAWS/GCP>
	if *delallServersAWS != false {
                fmt.Println("######### delete all servers in AWS....")
                delAllServersAWS()
        }

}



// 1.1. create Servers(VM).
// 1.2. get servers' public IP.
// 1.3. insert Farmoni Agent into Servers.
// 1.4. execute Servers' Agent.
// 1.5. add server list into etcd.
func addServersAWS(count int) {
// ==> AWS-EC2
    region := "ap-northeast-2" // seoul region.

    svc := ec2handler.Connect(region)

// 1.1. create Servers(VM).
    // some options are static for simple PoC.
    // These must be prepared before.
    instanceIds := ec2handler.CreateInstances(svc, "ami-047f7b46bd6dd5d84", "t2.micro", 1, count,
        "aws.powerkim.keypair", "sg-2334584f", "subnet-8c4a53e4", "powerkimInstance_")

    publicIPs := make([]*string, len(instanceIds))

// 1.2. get servers' public IP.
    // waiting for completion of new instance running.
    // after then, can get publicIP.
    for k, v := range instanceIds {
            // wait until running status
            ec2handler.WaitForRun(svc, *v)
            // get public IP
            publicIP, err := ec2handler.GetPublicIP(svc, *v)
            if err != nil {
                fmt.Println("Error", err)
                return
            }
            fmt.Println("==============> " + publicIP);
	    publicIPs[k] = &publicIP
    }

    
// 1.3. insert Farmoni Agent into Servers.
// 1.4. execute Servers' Agent.
    for _, v := range publicIPs {
	    for i:=0; ; i++ {
		err:=copyAndPlayAgent(*v)
		if(i==30) { os.Exit(3) }
		    if err == nil {
			break;	
		    }
		    // need to load SSH Service on the VM
		    time.Sleep(time.Second*3)
	    } // end of for
    } // end of for

// 1.5. add server list into etcd.
    addServersToEtcd("aws", instanceIds, publicIPs)
}

// (1) get all AWS server id list from etcd
// (2) terminate all AWS servers
// (3) remove server list from etcd
func delAllServersAWS() {

// (1) get all AWS server id list from etcd
    idList := getInstanceIdListAWS()

// (2) terminate all AWS servers
    region := "ap-northeast-2" // seoul region.

    svc := ec2handler.Connect(region)

//  destroy Servers(VMs).
    ec2handler.DestroyInstances(svc, idList)


// (3) remove all aws server list from etcd
    delProviderAllServersFromEtcd(string("aws"))
}

func addServersToEtcd(provider string, instanceIds []*string, serverIPs []*string) {

        etcdcli, err := etcdhandler.Connect(etcdServerPort)
        if err != nil {
                panic(err)
        }
        fmt.Println("connected to etcd - " + *etcdServerPort)

        defer etcdhandler.Close(etcdcli)

        ctx := context.Background()

        for i, v := range serverIPs {
                serverPort := *v + ":2019" // 2019 Port is dedicated value for PoC.
                fmt.Println("######### addServer...." + serverPort)
		// /server/aws/i-1234567890abcdef0/129.254.175:2019  PULL
                etcdhandler.AddServer(ctx, etcdcli, &provider, instanceIds[i], &serverPort, fetchType)
        }

}


func delProviderAllServersFromEtcd(provider string) {
        etcdcli, err := etcdhandler.Connect(etcdServerPort)
        if err != nil {
                panic(err)
        }
        fmt.Println("connected to etcd - " + *etcdServerPort)

        defer etcdhandler.Close(etcdcli)

        ctx := context.Background()

	fmt.Println("######### delete " + provider + " all Server....")
	etcdhandler.DelProviderAllServers(ctx, etcdcli, &provider)
}


func copyAndPlayAgent(serverIP string) error {

        // server connection info
	// some options are static for simple PoC.// some options are static for simple PoC.
        // These must be prepared before.
        userName := "ec2-user"
        keyName := "/root/.aws/awspowerkimkeypair.pem"
        port := ":22"
        serverPort := serverIP + port

        // file info to copy
        sourceFile := "/root/go/src/farmoni/farmoni_agent/farmoni_agent"
        targetFile := "/tmp/farmoni_agent"

        // command for ssh run
        cmd := "/tmp/farmoni_agent &"

        // Connect to the server for scp
        scpCli, err := scp.Connect(userName, keyName, serverPort)
        if err != nil {
                fmt.Println("Couldn't establisch a connection to the remote server ", err)
                return err
        }

        // copy agent into the server.
        if err := scp.Copy(scpCli, sourceFile, targetFile); err !=nil {
                fmt.Println("Error while copying file ", err)
                return err
        }

        // close the session
        scp.Close(scpCli)


        // Connect to the server for ssh
        sshCli, err := sshrun.Connect(userName, keyName, serverPort)
        if err != nil {
                fmt.Println("Couldn't establisch a connection to the remote server ", err)
                return err
        }

        if err := sshrun.RunCommand(sshCli, cmd); err != nil {
                fmt.Println("Error while running cmd: " + cmd, err)
                return err
        }

        sshrun.Close(sshCli)

	return err
}


func addServersGCP(count int) {
}

func serverList() {

	list := getServerList()
	fmt.Print("######### all server list....(" + strconv.Itoa(len(list)) + ")\n")

	for _, v := range list {
		fmt.Println(*v)
	}	
}

func getServerList() []*string {

        etcdcli, err := etcdhandler.Connect(etcdServerPort)
        if err != nil {
                panic(err)
        }
        fmt.Println("connected to etcd - " + *etcdServerPort)

        defer etcdhandler.Close(etcdcli)

        ctx := context.Background()
        return etcdhandler.ServerList(ctx, etcdcli)
}

func getInstanceIdListAWS() []*string {
        etcdcli, err := etcdhandler.Connect(etcdServerPort)
        if err != nil {
                panic(err)
        }
        fmt.Println("connected to etcd - " + *etcdServerPort)

        defer etcdhandler.Close(etcdcli)

        ctx := context.Background()
        return etcdhandler.InstanceIDListAWS(ctx, etcdcli)
}

func monitoringAll() {

	for {
		list := getServerList()
		for _, v := range list {
			monitoringServer(*v)
			println("-----------")
		}

                println("==============================")
		time.Sleep(time.Second)
	} // end of for
}

func monitoringServer(serverPort string) {

        // Set up a connection to the server.
        conn, err := grpc.Dial(serverPort, grpc.WithInsecure())
        if err != nil {
                log.Fatalf("did not connect: %v", err)
        }
        defer conn.Close()
        c := pb.NewResourceStatClient(conn)

        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Hour)
        defer cancel()


	r, err := c.GetResourceStat(ctx, &pb.ResourceStatRequest{})
	if err != nil {
		log.Fatalf("could not Fetch Resource Status Information: %v", err)
	}
	println("[" + r.Servername + "]")
	log.Printf("%s", r.Cpu)
	log.Printf("%s", r.Mem)
	log.Printf("%s", r.Dsk)

}
