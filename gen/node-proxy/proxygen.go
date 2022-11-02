package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ipfs-force-community/chain-co/api"
	"github.com/ipfs-force-community/chain-co/gen/node-proxy/gen"
)

func main() {
	pkgName := "proxy"

	var proxy api.Proxy
	var local api.Local
	var unsupport api.UnSupport

	targets := []struct {
		def        interface{}
		structName string
		outPath    string
	}{
		{
			def:        &proxy,
			structName: "Proxy",
			outPath:    "./proxy/proxy.go",
		},
		{
			def:        &local,
			structName: "Local",
			outPath:    "./proxy/local.go",
		},
		{
			def:        &unsupport,
			structName: "UnSupport",
			outPath:    "./proxy/unsupport.go",
		},
	}

	for _, t := range targets {
		code, err := gen.Gen(pkgName, t.structName, t.def)
		if err != nil {
			fmt.Println("ERR:", err)
			os.Exit(1)
		}

		if err := ioutil.WriteFile(t.outPath, code, 0644); err != nil {
			fmt.Println("ERR:", err)
			os.Exit(1)
		}
	}
}
