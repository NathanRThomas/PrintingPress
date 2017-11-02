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

const APP_VER = "0.1"

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type app_t struct {

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

/*! \brief main check function
    This will look for newly generated journalctl logs and parse them and if needed, write them to the log file
*/
func checkServices (config *app_t) error {
    jService := journal_c {}   //journal class

    for idx, s := range(config.Services) {
        lines, err := jService.Check(&config.Services[idx])  //pass a pointer so this can update itself
        if err == nil {
            if len(lines) > 0 { //new lines for the log, these come in in reverse order
                f := outputHandle(&s)
                defer f.Close()

                i := len(lines) -1  //start at the end
                for i >= 0 {
                    f.WriteString(lines[i] + "\n")
                    i--
                }
            }
        } else {
            return fmt.Errorf("error for service %s :: %s\n", s.Term, err.Error())
        }
    }

    return nil
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- MAIN --------------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func main() {
    log.SetFlags(log.LstdFlags | log.Lshortfile) //configure the logging for this application
	
	//handle any passed in flags
	configFile := flag.String("c", "printing_press.json", "Config file")
	versionFlag := flag.Bool("v", false, "Returns the version")
	
	flag.Parse()
	
	if *versionFlag {
		fmt.Printf("\nPrinting Press Version: %s\n\n", APP_VER)
		os.Exit(0)
	}

    app := app_t {} //init our application level object

    //parse the config file
    err := parseConfig(*configFile, &app)
    if err != nil {
        log.Printf("Unable to start. Issue with the config file\n%s\n", err.Error())
        os.Exit(1)  //bail
    }

    err = checkServices(&app)   //do the first check right away, this just gets a baseline of the last message for all our services
    if err != nil {
        log.Println(err)
        os.Exit(2)  //bail
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
	
	//this is our polling interval
	ticker := time.NewTicker(time.Minute * time.Duration(1))	//check every minute, that's our min
	go func() {
		for range ticker.C {
			err := checkServices(&app)	//does the check of all our services
            if err != nil { log.Println(err) }
		}
	} ()
	
	wg.Wait()	//wait for the slave and possible master to finish
}
