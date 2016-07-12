package main

import (
	"fmt"
	"github.com/Cepave/open-falcon-backend/common/logruslog"
	"github.com/Cepave/open-falcon-backend/common/vipercfg"
	"github.com/Cepave/open-falcon-backend/modules/transfer/g"
	"github.com/Cepave/open-falcon-backend/modules/transfer/http"
	"github.com/Cepave/open-falcon-backend/modules/transfer/proc"
	"github.com/Cepave/open-falcon-backend/modules/transfer/receiver"
	"github.com/Cepave/open-falcon-backend/modules/transfer/sender"
	flag "github.com/spf13/pflag"
	"os"
)

func main() {
	cfg := flag.String("c", "cfg.json", "configuration file")
	version := flag.Bool("v", false, "show version")
	versionGit := flag.Bool("vg", false, "show version")
	flag.Parse()

	if *version {
		fmt.Println(g.VERSION)
		os.Exit(0)
	}
	if *versionGit {
		fmt.Println(g.VERSION, g.COMMIT)
		os.Exit(0)
	}

	// global config
	vipercfg.Load()
	g.ParseConfig(*cfg)
	logruslog.Init()
	// proc
	proc.Start()

	sender.Start()
	receiver.Start()

	// http
	http.Start()

	select {}
}
