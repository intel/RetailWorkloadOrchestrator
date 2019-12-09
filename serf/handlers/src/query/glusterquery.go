//Copyright 2019, Intel Corporation

package main

import (
	"fmt"
	"helpers"
	"io/ioutil"
	"os"
	"strings"
)

func peerProbe(hostname string) {
	client := helpers.GlusterLibClient()
	msg, _ := client.PeerProbeWithMsg(hostname)
	fmt.Println(msg)
}

func poolList() {
	client := helpers.GlusterLibClient()
	peers, _ := client.PeerStatus()

	for _, entry := range peers {
		fmt.Printf("%s %s %s\n", entry.ID, entry.Name, entry.Status)
	}
}

func main() {

	var input string

	// Get the input from standard input.
	inputCmd, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Print("Error in reading input.", err.Error())
	}
	if len(inputCmd) == 0 {
		fmt.Println("Serf query for gluster received empty inputs")
		return
	}

	input = string(inputCmd)
	if len(input) == 0 {
		fmt.Println("Serf query for gluster received empty inputs")
		return
	}

	// Supported query inputs are:
	// - peer probe
	// - pool list
	if strings.Contains(input, "peer probe") {
		// parse the string for the peer id
		strs := strings.Split(input, " ")

		if len(strs) > 2 {
			_peerID := strs[2]
			// Filter any newlines
			hostname := strings.Split(_peerID, "\n")[0]
			peerProbe(hostname)
		}
	}

	if strings.Contains(input, "pool list") {
		poolList()
	}
}
