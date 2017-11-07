/*! \file service.go
    \brief Main file for the Printing Press service.
    Written in GO
    Created 2017-11-02 By Nathan Thomas
    
_ = "breakpoint"
*/

package main

import (
    "log"
	"fmt"
	"os"
    "encoding/json"
	"os/signal"
	"time"
	"sync"
	"flag"
)

const APP_VER = "0.3"

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type app_t struct {
    Verbose     bool
    Services    []service_t     `json:"services"`
}


  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func outputHandle (service *service_t) *os.File {
    f, err := os.OpenFile(service.Output, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0664)
    if err != nil { panic(err) }    //this is bad
    return f
}

/*! \brief Handles parsing of the config file for this instance
*/
func parseConfig (fileName string, config *app_t) error {
    configFile, err := os.Open(fileName) //try the file
    
    if err == nil {
        defer configFile.Close()
        jsonParser := json.NewDecoder(configFile)
        err = jsonParser.Decode(config)

        if err != nil {
            return fmt.Errorf("%s file appears to have invalid json :: " + err.Error(), fileName)
        } else if len(config.Services) < 1 {
            return fmt.Errorf("Please add at least one service to your config")
        }

        //now validate all the service objects
        for idx, s := range(config.Services) {
            if len(s.Output) == 0 {
                return fmt.Errorf("Output file location for service number %d[%s] is missing", idx, s.Term)
            } else {
                f := outputHandle(&s)
                f.WriteString("**** Printing Press started monitoring servce ****\n")
                f.Close()
            }
        }

        return nil  //we're good
    } else {
        return fmt.Errorf("Unable to open '%s' file :: " + err.Error(), fileName)
    }
}


  //-------------------------------------------------------------------------------------------------------------------------//
 //----- MAIN --------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func main() {
    log.SetFlags(log.LstdFlags | log.Lshortfile) //configure the logging for this application
	
	//handle any passed in flags
	configFile := flag.String("c", "printing_press.json", "Config file")
    verboseFlag := flag.Bool("V", false, "Verbose logging for what this service is doing")
	versionFlag := flag.Bool("v", false, "Returns the version")
    
	flag.Parse()
	
	if *versionFlag {
		fmt.Printf("\nPrinting Press Version: %s\n\n", APP_VER)
		os.Exit(0)
	}

    app := app_t {} //init our application level object
    app.Verbose = *verboseFlag  //copy over the verbose flag
    
    //parse the config file
    err := parseConfig(*configFile, &app)
    if err != nil {
        log.Printf("Unable to start. Issue with the config file\n%s\n", err.Error())
        os.Exit(1)  //bail
    }

    timeChan := make(chan time.Time, len(app.Services))    //this will be our exit flag

    jService := journal_c {Verbose: app.Verbose}   //journal class
    //loop through our services and create a follower for each
    for idx, _ := range(app.Services) {
        go jService.Follow(app.Services[idx], timeChan)    //pass the channel, this is how we'll exit these loops
    }
    
	wg := new(sync.WaitGroup)
	wg.Add(1)
	
	//this handles killing the service gracefully
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func(wg *sync.WaitGroup){
		<-c
			fmt.Println("Printing Press service exiting")
			wg.Done()
	}(wg)
	
	wg.Wait()	//wait till we hear an interrupt
    //if we're here it's cause we need to exit
    for _, _ = range(app.Services) {
        timeChan <- time.Now()
    }

    time.Sleep(time.Second * 1) //give it just a second to exit
}
