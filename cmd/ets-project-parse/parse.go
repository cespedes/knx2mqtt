package main

import (
	"encoding/xml"
	"io"
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
				Id           string `xml:",attr"`
				Address      string `xml:",attr"`
				Name         string `xml:",attr"`
				Comment      string `xml:",attr"`
				Description  string `xml:",attr"`
				ProductRefId string `xml:",attr"`
				SerialNumber string `xml:",attr"`
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
	return &k, nil
}
