// Proof of Concepts for the Cloud-Barista Multi-Cloud Project.
//      * Cloud-Barista: https://github.com/cloud-barista
//
// ETCD Hander (ETCD Version 3 API, Thanks ETCD.)
//
// by powerkim@powerkim.co.kr, 2019.03.
package etcdhandler

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
	"log"
	"strings"
)


func Connect(moniServerPort *string) (*clientv3.Client, error) {

	etcdcli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"http://" + *moniServerPort},
                DialTimeout: 5 * time.Second,
        })

	return etcdcli, err
}

func Close(cli *clientv3.Client) {
	cli.Close()
}

func AddServer(ctx context.Context, cli *clientv3.Client, addserver *string, fetchtype *string) {
	cli.Put(ctx, "/server/"+ *addserver, *fetchtype)
	fmt.Println("added a " + *addserver + " into the Server List...\n")
}

func DelServer(ctx context.Context, cli *clientv3.Client, delserver *string) {
	cli.Delete(ctx, "/server/"+ *delserver)
	/*if err != nil {
		log.Fatal(err)
	}*/
	fmt.Println("deleted a " + *delserver + " from the Server List...\n")
}

func ServerList(ctx context.Context, cli *clientv3.Client) []*string {
	// get with prefix, all list of /server's key
	resp, err := cli.Get(ctx, "/server", clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	if err != nil {
		log.Fatal(err)
	}

	serverList := make([]*string, len(resp.Kvs))
	for k, ev := range resp.Kvs {
		//fmt.Printf("%s : %s\n", strings.Trim(string(ev.Key), "/server/"), ev.Value)
		tmpStr := strings.Trim(string(ev.Key), "/server/")
		serverList[k] = &tmpStr
	}

	return serverList
}

func ServerListWithFetchType(ctx context.Context, cli *clientv3.Client) []*string {
        // get with prefix, all list of /server's key
        resp, err := cli.Get(ctx, "/server", clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
        if err != nil {
                log.Fatal(err)
        }

        serverList := make([]*string, len(resp.Kvs))
        for k, ev := range resp.Kvs {
                //fmt.Printf("%s : %s\n", strings.Trim(string(ev.Key), "/server/"), ev.Value)
		tmpStr := strings.Trim(string(ev.Key), "/server/") + " : " + string(ev.Value)
		serverList[k] = &tmpStr
        }

        return serverList
}
