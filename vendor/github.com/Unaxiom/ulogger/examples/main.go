package main

import (
	"fmt"

	"github.com/Unaxiom/ulogger"
)

func main() {
	log := ulogger.New()
	log.LineNumber = false
	localMap := make(map[string]int)
	localMap["one"] = 1
	localMap["two"] = 2
	localMap["three"] = 3
	log.SetLogLevel(ulogger.DebugLevel)
	fmt.Println("*********************DEBUG*********************************")
	log.Debugf("%s\n", "Hey there!")
	log.Debug("Hey there!\n")
	log.Debugln("Hey there!")
	log.DebugDump(localMap)
	fmt.Println("*********************INFO*********************************")
	log.Infof("%s\n", "Hey There")
	log.Info("Hey there!\n")
	log.Infoln("Hey there!")
	log.InfoDump(localMap)
	fmt.Println("*********************WARNING*********************************")
	log.Warningf("%s\n", "Hey There")
	log.Warning("Hey there!\n")
	log.Warningln("Hey there!")
	log.WarningDump(localMap)
	fmt.Println("*********************ERROR*********************************")
	log.Errorf("%s\n", "Hey There")
	log.Error("Hey there!\n")
	log.Errorln("Hey there!")
	log.ErrorDump(localMap)
	fmt.Println("*********************FATAL*********************************")
	// log.Fatalf("%s\n", "Hey There")
	// log.Fatal("Hey there!\n")
	// log.Fatalln("Hey there!")
	// log.FatalDump(localMap)
}
