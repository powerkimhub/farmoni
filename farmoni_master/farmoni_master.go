// Proof of Concepts for the Cloud-Barista Multi-Cloud Project.
//      * Cloud-Barista: https://github.com/cloud-barista
//
// Farmoni Master to control server and fetch monitoring info.
//
// by powerkim@powerkim.co.kr, 2019.03.
package main

import (
	"flag"
	"farmoni/farmoni_master/ec2handler"
	"farmoni/farmoni_master/serverhandler/scp"
	"farmoni/farmoni_master/serverhandler/sshrun"
	"farmoni/farmoni_master/etcdhandler"
	"fmt"
	"context"
	"time"
	"log"
        "google.golang.org/grpc"
        pb "farmoni/grpc_def"
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


func parseRequest() {

        etcdServerPort = flag.String("etcdserver", "129.254.175.43:2379", "etcdserver=129.254.175.43:2379")
        fetchType = flag.String("fetchtype", "PULL", "fetch type: -fetchtype=PUSH")
/*
        addServer = flag.String("addserver", "none", "add a server: -addserver=192.168.0.10:5000")
        delServer = flag.String("delserver", "none", "delete a server: -delserver=192.168.0.10")
*/
        addServerNumAWS = flag.Int("addserversaws", 0, "add servers in AWS: -addserversaws=10")
//        delServerNumAWS = flag.Int("delserversaws", 0, "delete servers in AWS: -delserversaws=10") // dedicated option for PoC

        addServerNumGCP = flag.Int("addserversgcp", 0, "add servers in GCP: -addserversgcp=10")
//       delServersNumGCP = flag.Int("delserversgcp", 0, "delete servers in GCP: -delserversgcp=10") // dedicated option for PoC

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
func main() {

	// 1. parsing user's request.
	parseRequest()

// add server in AWS/GCP	
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
                fmt.Println("######### list of all servers....")
                serverList()
        }
// 2.2. fetch all agent's monitoring info.
	if *monitoring != false {
                fmt.Println("######### monitoring all servers....")
                monitoringAll()
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

    // need to load SSH Service on the VM
    time.Sleep(time.Second*3)
    
// 1.3. insert Farmoni Agent into Servers.
// 1.4. execute Servers' Agent.
    for _, v := range publicIPs {
	    copyAndPlayAgent(*v)
    }

// 1.5. add server list into etcd.
    addServersToEtcd(publicIPs)
}

func addServersToEtcd(serverIPs []*string) {

        etcdcli, err := etcdhandler.Connect(etcdServerPort)
        if err != nil {
                panic(err)
        }
        fmt.Println("connected to etcd - " + *etcdServerPort)

        defer etcdhandler.Close(etcdcli)

        ctx := context.Background()

	for _, v := range serverIPs {
		serverPort := *v + ":2019" // 2019 Port is dedicated value for PoC.
                fmt.Println("######### addServer...." + serverPort)
                etcdhandler.AddServer(ctx, etcdcli, &serverPort, fetchType)
        }
}

func copyAndPlayAgent(serverIP string) {

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
                return
        }

        // copy agent into the server.
        if err := scp.Copy(scpCli, sourceFile, targetFile); err !=nil {
                fmt.Println("Error while copying file ", err)
        }

        // close the session
        scp.Close(scpCli)


        // Connect to the server for ssh
        sshCli, err := sshrun.Connect(userName, keyName, serverPort)
        if err != nil {
                fmt.Println("Couldn't establisch a connection to the remote server ", err)
                return
        }

        if err := sshrun.RunCommand(sshCli, cmd); err != nil {
                fmt.Println("Error while running cmd: " + cmd, err)
        }

        sshrun.Close(sshCli)
}


func addServersGCP(count int) {
}

func serverList() {

	fmt.Println("######### server list....")
	list := getServerList()
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
