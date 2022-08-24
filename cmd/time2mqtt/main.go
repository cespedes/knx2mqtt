package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/at-wat/mqtt-go"
	"github.com/sj14/astral/pkg/astral"
)

const (
	MQTTPort = 1883
)

type Config struct {
	Debug      bool
	MQTTServer string
	MQTTPrefix string
	Lat        float64
	Lon        float64
	Elev       float64
}

type Server struct {
	debug      bool
	mqtt       mqtt.Client
	mqttPrefix string
	observer   astral.Observer
	last       struct {
		year          string
		month         string
		dom           string
		dow           int
		hhmm          string
		prevSunrise   time.Time
		nextSunrise   time.Time
		deltaSunrise  string
		prevSunset    time.Time
		nextSunset    time.Time
		deltaSunset   string
		prevNoon      time.Time
		nextNoon      time.Time
		deltaNoon     string
		prevMidnight  time.Time
		nextMidnight  time.Time
		deltaMidnight string
	}
}

func calcPrevNext(server *Server, t time.Time, f func(astral.Observer, time.Time) time.Time) (prev time.Time, next time.Time) {
	calc := f(server.observer, t)
	if calc.After(t) {
		return f(server.observer, t.Add(-24*time.Hour)), calc
	}
	return calc, f(server.observer, t.Add(24*time.Hour))
}

func calcDelta(prev, next, now time.Time) string {
	var sign string
	var delta time.Duration
	delta = next.Sub(now) + time.Minute - 1 // negative minutes are rounded up
	sign = "-"
	if now.Sub(prev) < delta {
		delta = now.Sub(prev)
		sign = "+"
	}
	return fmt.Sprintf("%s%02d:%02d", sign, int(delta.Hours()), int((delta % time.Hour).Minutes()))
}

func formatDuration(d time.Duration) string {
	sign := "+"
	if d < 0 {
		d = -d
		sign = "-"
	}
	return sign
}

func announce(s *Server, key, value string) {
	if s.debug {
		log.Printf("MQTT: %s = %s\n", key, value)
	}
	mqttPublish(s.mqtt, fmt.Sprintf("%s/%s", s.mqttPrefix, key), value)
}

func loop(s *Server, t time.Time) {
	if s.debug {
		log.Printf("tick: %s\n", t.Format("2006-01-02/15:04:05.000000"))
	}

	if s.last.nextSunrise == (time.Time{}) || t.After(s.last.nextSunrise) {
		fSunrise := func(observer astral.Observer, t time.Time) time.Time {
			calc, _ := astral.Sunrise(observer, t)
			return calc
		}
		s.last.prevSunrise, s.last.nextSunrise = calcPrevNext(s, t, fSunrise)
		announce(s, "prev-sunrise", s.last.prevSunrise.Format("2006-01-02/15:04:05"))
		announce(s, "next-sunrise", s.last.nextSunrise.Format("2006-01-02/15:04:05"))
	}
	delta := calcDelta(s.last.prevSunrise, s.last.nextSunrise, t)
	if delta != s.last.deltaSunrise {
		s.last.deltaSunrise = delta
		announce(s, "delta-sunrise", delta)
	}

	if s.last.nextSunset == (time.Time{}) || t.After(s.last.nextSunset) {
		fSunset := func(observer astral.Observer, t time.Time) time.Time {
			calc, _ := astral.Sunset(observer, t)
			return calc
		}
		s.last.prevSunset, s.last.nextSunset = calcPrevNext(s, t, fSunset)
		announce(s, "prev-sunset", s.last.prevSunset.Format("2006-01-02/15:04:05"))
		announce(s, "next-sunset", s.last.nextSunset.Format("2006-01-02/15:04:05"))
	}
	delta = calcDelta(s.last.prevSunset, s.last.nextSunset, t)
	if delta != s.last.deltaSunset {
		s.last.deltaSunset = delta
		announce(s, "delta-sunset", delta)
	}

	if s.last.nextNoon == (time.Time{}) || t.After(s.last.nextNoon) {
		s.last.prevNoon, s.last.nextNoon = calcPrevNext(s, t, astral.Noon)
		announce(s, "prev-noon", s.last.prevNoon.Format("2006-01-02/15:04:05"))
		announce(s, "next-noon", s.last.nextNoon.Format("2006-01-02/15:04:05"))
	}
	delta = calcDelta(s.last.prevNoon, s.last.nextNoon, t)
	if delta != s.last.deltaNoon {
		s.last.deltaNoon = delta
		announce(s, "delta-noon", delta)
	}

	if s.last.nextMidnight == (time.Time{}) || t.After(s.last.nextMidnight) {
		s.last.prevMidnight, s.last.nextMidnight = calcPrevNext(s, t, astral.Midnight)
		announce(s, "prev-midnight", s.last.prevMidnight.Format("2006-01-02/15:04:05"))
		announce(s, "next-midnight", s.last.nextMidnight.Format("2006-01-02/15:04:05"))
	}
	delta = calcDelta(s.last.prevMidnight, s.last.nextMidnight, t)
	if delta != s.last.deltaMidnight {
		s.last.deltaMidnight = delta
		announce(s, "delta-midnight", delta)
	}

	year := t.Format("2006")
	if year != s.last.year {
		announce(s, "year", t.Format("2006"))
		s.last.year = year
	}
	month := t.Format("1")
	if month != s.last.month {
		announce(s, "month", t.Format("1"))
		s.last.month = month
	}
	dom := t.Format("2")
	if dom != s.last.dom {
		announce(s, "day-of-month", t.Format("2"))
		announce(s, "day-of-week", fmt.Sprint(int(t.Weekday())))
		announce(s, "yyyy-mm-dd", t.Format("2006-01-02"))
		s.last.dom = dom
	}
	hhmm := t.Format("15:04")
	if hhmm != s.last.hhmm {
		announce(s, "hh:mm", t.Format("15:04"))
		s.last.hhmm = hhmm
	}
	announce(s, "hh:mm:ss", t.Format("15:04:05"))
}

func main() {
	var err error
	var config Config
	flag.BoolVar(&config.Debug, "debug", false, "Debugging")
	flag.StringVar(&config.MQTTServer, "mqtt", "", "MQTT server")
	flag.StringVar(&config.MQTTPrefix, "mqtt-prefix", "timer", "MQTT prefix to use")
	flag.Float64Var(&config.Lat, "lat", 40.417, "Latitude (degrees)")
	flag.Float64Var(&config.Lon, "lon", -3.703, "Longitude (degrees)")
	flag.Float64Var(&config.Elev, "elev", 650, "Elevation (meters)")
	flag.Parse()

	if config.Debug {
		log.Printf("config = %+v\n", config)
	}

	if len(config.MQTTServer) == 0 {
		log.Fatalf("No MQTT server specified")
	}

	var server Server
	server.debug = config.Debug
	server.mqttPrefix = config.MQTTPrefix

	// get channel to write MQTT messages
	if config.Debug {
		log.Printf("connecting to MQTT server %s\n", config.MQTTServer)
	}
	server.mqtt, err = mqttConnect(config.MQTTServer, MQTTPort)
	if err != nil {
		log.Fatalf("MQTT: Could not connect to %q: %v", config.MQTTServer, err)
	}
	if config.Debug {
		log.Printf("MQTT: Connected.")
	}

	server.observer = astral.Observer{Latitude: config.Lat, Longitude: config.Lon, Elevation: config.Elev}

	t := time.Now()

	// fmt.Printf("t = %v\n", t)

	ttrunc := t.Truncate(time.Second)
	// fmt.Printf("ttrunc = %v\n", ttrunc)
	tafter := ttrunc.Add(time.Second)
	// fmt.Printf("tafter = %v\n", tafter)
	sleep := tafter.Sub(t)
	// fmt.Printf("sleep = %v\n", sleep)
	time.Sleep(sleep)

	loop(&server, t)
	// t = time.Now()
	// fmt.Printf("t = %v\n", t)
	tick := time.Tick(time.Second)
	for c := range tick {
		loop(&server, c)
	}
}
