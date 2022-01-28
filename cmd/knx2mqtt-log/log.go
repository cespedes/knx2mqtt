package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"
)

func (s *Server) Log(e Event) {
	var err error
	filename := path.Join(config.Logdir, time.Now().Format("2006/0102.log"))
	if s.logFileName != filename {
		s.logFile.Close()
		os.MkdirAll(filepath.Dir(filename), 0777)
		s.logFile, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}
		s.logFileName = filename
	}
	fmt.Fprintf(s.logFile, "%s\n", e)
}
