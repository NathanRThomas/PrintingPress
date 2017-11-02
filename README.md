# PrintingPress
Service to pull journalctl/journald logs on linux machines and parse/record them to a log file.

##This uses
```
github.com/coreos/go-systemd
github.com/coreos/pkg

import "github.com/coreos/go-systemd/sdjournal"
```

##Overview
The goal of this is to monitor journalctl and using regex pull out important info and write it to a log file.
That log file can stay, or be used by filebeat to push to an elastic search setup, or just use logrotate to keep track manually of important info.

