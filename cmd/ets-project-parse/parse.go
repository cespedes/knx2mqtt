package main

import (
	"encoding/xml"
	"io"
	"sort"
	"strconv"
)

type Project struct {
	Topology       Topology       `xml:"Project>Installations>Installation>Topology"`
	Locations      Locations      `xml:"Project>Installations>Installation>Locations"`
	GroupAddresses GroupAddresses `xml:"Project>Installations>Installation>GroupAddresses"`
}

type Topology struct {
	Area []struct {
		Id      string `xml:",attr"`
		Address string `xml:",attr"`
		Name    string `xml:",attr"`
		Line    []struct {
			Id             string `xml:",attr"`
			Address        string `xml:",attr"`
			Name           string `xml:",attr"`
			DeviceInstance []struct {
				Id                   string `xml:",attr"`
				Address              string `xml:",attr"`
				Name                 string `xml:",attr"`
				Comment              string `xml:",attr"`
				Description          string `xml:",attr"`
				ProductRefId         string `xml:",attr"`
				SerialNumber         string `xml:",attr"`
				ComObjectInstanceRef []struct {
					Text  string `xml:",attr"`
					Links string `xml:",attr"`
				} `xml:"ComObjectInstanceRefs>ComObjectInstanceRef"`
			}
		}
	}
}

type Locations struct {
	Space []struct {
		Type string `xml:",attr"`
		Id   string `xml:",attr"`
		Name string `xml:",attr"`
	}
}

type GroupAddresses struct {
	GroupRanges struct {
		GroupRange []struct {
			Id          string `xml:",attr"`
			RangeStart  string `xml:",attr"`
			RangeEnd    string `xml:",attr"`
			Name        string `xml:",attr"`
			Description string `xml:",attr"`
			GroupRange  []struct {
				Id           string `xml:",attr"`
				RangeStart   string `xml:",attr"`
				RangeEnd     string `xml:",attr"`
				Name         string `xml:",attr"`
				Description  string `xml:",attr"`
				GroupAddress []struct {
					Id            string `xml:",attr"`
					Address       string `xml:",attr"`
					Name          string `xml:",attr"`
					Description   string `xml:",attr"`
					DatapointType string `xml:",attr"`
				}
			}
		}
	}
}

func ParseProject(r io.Reader) (*Project, error) {
	var k Project
	d := xml.NewDecoder(r)
	err := d.Decode(&k)
	if err != nil {
		return nil, err
	}

	// Sort everything:
	sort.Slice(k.Topology.Area, func(i, j int) bool {
		a1, _ := strconv.Atoi(k.Topology.Area[i].Address)
		a2, _ := strconv.Atoi(k.Topology.Area[j].Address)
		return a1 < a2
	})
	for _, a := range k.Topology.Area {
		sort.Slice(a.Line, func(i, j int) bool {
			a1, _ := strconv.Atoi(a.Line[i].Address)
			a2, _ := strconv.Atoi(a.Line[j].Address)
			return a1 < a2
		})
		for _, l := range a.Line {
			sort.Slice(l.DeviceInstance, func(i, j int) bool {
				a1, _ := strconv.Atoi(l.DeviceInstance[i].Address)
				a2, _ := strconv.Atoi(l.DeviceInstance[j].Address)
				return a1 < a2
			})
		}
	}
	sort.Slice(k.GroupAddresses.GroupRanges.GroupRange, func(i, j int) bool {
		a1, _ := strconv.Atoi(k.GroupAddresses.GroupRanges.GroupRange[i].RangeStart)
		a2, _ := strconv.Atoi(k.GroupAddresses.GroupRanges.GroupRange[j].RangeStart)
		return a1 < a2
	})
	for _, gr1 := range k.GroupAddresses.GroupRanges.GroupRange {
		sort.Slice(gr1.GroupRange, func(i, j int) bool {
			a1, _ := strconv.Atoi(gr1.GroupRange[i].RangeStart)
			a2, _ := strconv.Atoi(gr1.GroupRange[j].RangeStart)
			return a1 < a2
		})
		for _, gr2 := range gr1.GroupRange {
			sort.Slice(gr2.GroupAddress, func(i, j int) bool {
				a1, _ := strconv.Atoi(gr2.GroupAddress[i].Address)
				a2, _ := strconv.Atoi(gr2.GroupAddress[j].Address)
				return a1 < a2
			})
		}
	}
	return &k, nil
}
