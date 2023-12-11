package main

import (
	"flag"
	"fmt"
	"net/rpc"
)

func cli(bindAddress string, id string) {
	theme := flag.String("theme", "", "Set the current theme")
	toaster := flag.String("toaster", "", "Pop a toaster")
	download := flag.String("download", "", "Download a file")
	flag.Parse()

	rpcClient, err := rpc.DialHTTP("tcp", bindAddress)
	if err != nil {
		panic(err)
	}

	if *theme != "" {
		var reply string
		if err := rpcClient.Call("Context.SetTheme", &SetThemeArgs{RPCArgs: RPCArgs{ConnectionId: id}, ThemeId: *theme}, &reply); err != nil {
			panic(err)
		}
		fmt.Println(reply)
	}
	if *toaster != "" {
		var reply bool
		if err := rpcClient.Call("Context.PopToaster", &PopToasterArgs{RPCArgs: RPCArgs{ConnectionId: id}, Text: *toaster}, &reply); err != nil {
			panic(err)
		}
		if reply {
			fmt.Println("Popped a toaster")
		}
	}
	if *download != "" {
		var reply string
		if err := rpcClient.Call("Context.Download", &DownloadArgs{RPCArgs: RPCArgs{ConnectionId: id}, Path: *download}, &reply); err != nil {
			panic(err)
		}
		fmt.Println(reply)
	}
	rpcClient.Close()

}
