package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"secondbit.org/gifs/api"

	"github.com/coreos/go-etcd/etcd"
	"github.com/gorilla/mux"
)

var (
	etcdAddrs = StringArray{}
)

type StringArray []string

func (a *StringArray) Set(s string) error {
	*a = append(*a, s)
	return nil
}

func (a *StringArray) String() string {
	return strings.Join(*a, ",")
}

func main() {
	flag.Var(&etcdAddrs, "etcd-address", "address to the etcd server (may be specified more than once)")
	flag.Parse()
	if len(etcdAddrs) < 1 {
		fmt.Println("Must set at least once etcd address.")
		return
	}
	client := etcd.NewClient(etcdAddrs)
	resp, err := client.Get("/", false, true)
	if err != nil {
		fmt.Println(err)
		return
	}
	context, err := getEtcdContext(resp.Node)
	if err != nil {
		fmt.Println(err)
		return
	}
	listenAddr := api.DefaultListenAddr
	for _, node := range resp.Node.Nodes {
		fmt.Println(node.Key)
		if node.Key == "/listen_addr" {
			listenAddr = node.Value
			fmt.Println("Set listenAddr to " + node.Value)
			break
		}
	}

	var router *mux.Router
	if context.RootDomain == "" {
		fmt.Println("Using path muxer")
		router = api.GetPathMuxer(context)
	} else {
		fmt.Println("Using domain muxer")
		router = api.GetDomainMuxer(context)
	}
	http.Handle("/", router)
	fmt.Println("Listening on " + listenAddr)
	err = http.ListenAndServe(listenAddr, nil)
	if err != nil {
		panic(err)
	}
}
