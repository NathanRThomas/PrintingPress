/*! \file crow.go
  \brief Handles mostly initialization and setup stuff from our config file for the service
*/

package main

import (
//    "fmt"
    "log"
    "regexp"
    "time"
    "os"

    "github.com/coreos/go-systemd/sdjournal"
    )

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type service_t struct {
    Term            string  `json:"search_term"`
    Output          string  `json:"output_file"`
}

type journal_c struct {
    Verbose         bool
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PRIVATE FUNCTIONS -------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

/*! \brief Validates the syntax of the regex for capturing the page output
 */
func (j *journal_c) validateRegex (in string) {
    if len(in) > 0 {
        regexp.MustCompile(in)
    }
}

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- PUBLIC FUNCTIONS --------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

func (j *journal_c) Follow (service service_t, c <- chan time.Time) {
    m := make([]sdjournal.Match, 0)
    m = append(m, sdjournal.Match{Field: sdjournal.SD_JOURNAL_FIELD_SYSTEMD_UNIT, Value: service.Term})

    journal, err := sdjournal.NewJournalReader(sdjournal.JournalReaderConfig {Since: time.Duration(time.Millisecond), Matches: m})

    if err == nil {
        f, err := os.OpenFile(service.Output, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0664)
        if err == nil {
            defer f.Close()

            //we're good, now setup a follow
            log.Printf("Starting logging of %s\n", service.Term)
            journal.Follow (c, f)
            log.Printf("Service %s exiting\n", service.Term)
        } else {
            log.Printf("Error opening file %s for writing : %s\n", service.Output, err.Error())
        }
    } else {
        log.Printf("Error creating new Journal Reader : %s\n", err.Error())    //this is bad
    }
}