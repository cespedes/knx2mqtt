package main

import (
	"fmt"
	"strconv"
)

func GroupAddress(s string) string {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic("GroupAddress: " + err.Error())
	}
	if i < 0 || i > 65535 {
		panic(fmt.Sprintf("GroupAddress=%d out of range", i))
	}
	return fmt.Sprintf("%d/%d/%d", i>>11, (i>>8)&7, i&255)
}

func PrintProject(k *Project) {
	for _, a := range k.Topology.Area {
		fmt.Printf("Area %s (id=%q name=%q)\n", a.Address, a.Id, a.Name)
		for _, l := range a.Line {
			fmt.Printf("Line %s.%s (id=%q name=%q)\n", a.Address, l.Address, l.Id, l.Name)
			for _, d := range l.DeviceInstance {
				fmt.Printf("Device %s.%s.%s", a.Address, l.Address, d.Address)
				fmt.Printf(" (id=%q name=%q comment=%q description=%q productrefid=%q serialnumber=%q)\n",
					d.Id, d.Name, d.Comment, d.Description, d.ProductRefId, d.SerialNumber)
			}
		}
	}
	for _, s := range k.Locations.Space {
		fmt.Printf("Location: type=%q id=%q name=%q\n", s.Type, s.Id, s.Name)
	}
	for _, gr1 := range k.GroupAddresses.GroupRanges.GroupRange {
		fmt.Printf("Range: start=%s end=%s id=%q name=%q description=%q\n",
			gr1.RangeStart, gr1.RangeEnd, gr1.Id, gr1.Name, gr1.Description)
		for _, gr2 := range gr1.GroupRange {
			fmt.Printf("Subrange: start=%s end=%s id=%q name=%q description=%q\n",
				gr2.RangeStart, gr2.RangeEnd, gr2.Id, gr2.Name, gr2.Description)
			for _, a := range gr2.GroupAddress {
				fmt.Printf("Address %s (id=%q name=%q description=%q DP=%q)\n",
					GroupAddress(a.Address), a.Id, a.Name, a.Description, a.DatapointType)
			}
		}
	}
}
