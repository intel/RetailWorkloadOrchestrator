//Copyright 2019, Intel Corporation

package main

import (
	"fmt"
	"helpers"
	"io/ioutil"
	"os"
	mg "rwogluster"
	"strings"
	"time"
)

// peerProbe will ask gluster server to perform peer probe to the querying client.
func peerProbe(IP string) {

	//Check if peer is localhost.
	hostIP, _ := helpers.GetIPAddr()
	if hostIP == IP {
		fmt.Println("No Need to probe local host. success")
		return
	}

	client := helpers.GlusterLibClient()

	//Set the quorum ratio before cleanup.
	err := helpers.SetQuorumRatio(client, "cleanup")
	if err != nil {
		fmt.Println("Error while setting server quorum ratio ", err.Error())
	}

	//Check if there are any bricks of the querying client.
	checkForExistingBricks(client, IP)

	msg, _ := client.PeerProbeWithMsg(IP)
	fmt.Println(msg)

	//Reset quorum ratio.
	err = helpers.SetQuorumRatio(client, "update")
	if err != nil {
		fmt.Println("Error in Enabling server quorum ", err.Error())
	}

}

// checkForExistingBricks will remove the brick of the queried client in the existing volumes.
// This will ensure that node is able to join the cluster even after reboot.
func checkForExistingBricks(client *mg.Client, ipAddr string) {

	vols, _ := client.ListVolumes()
	for _, vol := range vols {

		//Remove the brick.
		replica := len(vol.Bricks) - 1
		for _, brick := range vol.Bricks {

			if strings.Contains(brick, ipAddr) {
				if replica > 0 {
					err := rmBricksFromVol(vol.Name, brick, replica, client)
					if err != nil {
						fmt.Println("Error Removing Brick ", err.Error())
					}
				}
			}
		}
	}

	peers, err := client.PeerStatus()
	if err != nil {
		fmt.Println("Error Getting Peer Status")
	}

	var peerDetachRequired bool
	peerDetachRequired = false

	for _, peer := range peers {
		if peer.Name == ipAddr {
			peerDetachRequired = true
		}
	}

	if peerDetachRequired {
		retry := 5
		for retry > 0 {
			err = client.PeerDetach(ipAddr)
			if err != nil {
				retry--
			} else {
				break
			}
		}
	}
}

// rmBricksFromVol removes bricks.
func rmBricksFromVol(n, b string, r int, client *mg.Client) error {

	retry := 5
	var err error
	for retry > 0 {
		err = client.RemoveBrick(n, b, r)
		if err == nil {
			break
		}

		fmt.Println("Error in RemoveBrick: Trying again", err)
		if strings.Contains(err.Error(), "lock") {
			// To ensure that glusterd has released the lock for some request
			time.Sleep(1 * time.Second)
		}
		retry--
	}

	return err
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
