/*! \file crow.go
  \brief Handles mostly initialization and setup stuff from our config file for the service
*/

package main

import (
    "fmt"
    "regexp"
    "strings"

    "github.com/coreos/go-systemd/sdjournal"
    )

  //-------------------------------------------------------------------------------------------------------------------------//
 //----- STRUCTS -----------------------------------------------------------------------------------------------------------//
//-------------------------------------------------------------------------------------------------------------------------//

type service_t struct {
    Term            string  `json:"search_term"`
    Output          string  `json:"output_file"`
    LastTimestamp   uint64  `json:"-"`
}

type journal_c struct {
    
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

/*! \brief Main entry point.  This will read our config files and make sure we can start running
 */
func (j *journal_c) Check (service *service_t) ([]string, error) {
    lines := make([]string, 0)  //if we find any new lines here

    var tm uint64
    journal, err := sdjournal.NewJournal()  //createa  new journal
    if err == nil {
        if len(service.Term) > 0 {  //we're looking for a specific service
            m := sdjournal.Match{Field: sdjournal.SD_JOURNAL_FIELD_SYSTEMD_UNIT, Value: service.Term}
            err = journal.AddMatch(m.String())    //add our term, if it exists
        }
        
        if err == nil { //still good
            //go to the end of the journal
            err = journal.SeekTail()

            if err == nil {
                tm, err = journal.Next()
                if tm != 0 {    //else we don't have an entry
                    if err == nil {
                        //get this entry
                        entry, lErr := journal.GetEntry()
                        if lErr == nil {
                            //fmt.Printf("%+v\n", entry)

                            if service.LastTimestamp == 0 {
                                service.LastTimestamp = entry.RealtimeTimestamp //copy this over for next time
                            } else if service.LastTimestamp < entry.RealtimeTimestamp {    //see if this is newer
                                mostRecentTimestamp := entry.RealtimeTimestamp  //record this for later
                                
                                for err == nil && tm != 0 && service.LastTimestamp < entry.RealtimeTimestamp {
                                    lines = append(lines, strings.TrimSpace(entry.Fields["MESSAGE"]))

                                    tm, _ = journal.Previous()  //now go back
                                    if tm != 0 {
                                        entry, err = journal.GetEntry() //get this one
                                    }
                                }

                                service.LastTimestamp = mostRecentTimestamp //save this for next time
                            }
                        } else {
                            err = lErr
                        }
                    }
                } else {
                    fmt.Printf("Unalbe to find any previous journal entries for %s\n", service.Term)
                    service.LastTimestamp = 1;  //look for any future entries for this service
                }
            }
        }

        if err == nil {
            err = journal.Close()   //we're done with it
        } else {
            journal.Close()   //we're done with it
        }
    }
    
    return lines, err  //we're done
}
