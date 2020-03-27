//Copyright 2019, Intel Corporation

package main

import (
	"helpers"
	"log"
	"os"
	mg "rwogluster"
	"strconv"
	"strings"
	"time"
)

// handlemembercleanup will be a forked process from member-failed.
// This will ensure that a rebooted member can join the serf cluster immediately without waiting for five minutes.
// This will wait for 300 seconds for member to show up.
// If member does not show up.
// Remove the member from docker swarm.
// Remove the bricks from gluster.
// Remove from serf cluster.

// Global variables
var serfRole string

var (
	client     *mg.Client
	newCluster bool
)

func main() {
	serfRole, _ := os.LookupEnv("SERF_TAG_ROLE")
	//check for network status
	err := helpers.CheckNetworkStatus()
	if err != nil {
		return
	}

	// Create  a new client
	client = helpers.GlusterLibClient()

	f, err := os.OpenFile("/var/log/membercleanup.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	if !helpers.IsValidRole(serfRole) {
		log.Println("SERF Tag from Env is invalid. ", serfRole)
		serfRole = ""
	}

	log.Println("******** Running serf member-Cleanup ********")
	waitForaMemberReboot()

	err = memberCleanup()
	if err != nil {
		log.Println(err.Error())
		return
	}

	err = removeDownNode()
	if err != nil {
		log.Println(err.Error())
		return
	}

	log.Println("********  Member-Cleanup completed ******** ")

	return
}

func waitForaMemberReboot() {
	delaySeconds := helpers.GetSleepTimeFromEnv("MEMBER_REBOOT_TIME_IN_SECONDS")
	log.Println("Handle member cleanup, sleeping for ", strconv.Itoa(delaySeconds), " seconds in case the member is rebooting. ")
	time.Sleep(time.Duration(delaySeconds) * time.Second)
}

// memberCleanup: Function to remove the failed member from docker & serf cluster.
func memberCleanup() error {

	//Do the Cleanup for failed nodes
	membersFailedCount := helpers.CountFailedMembers()
	log.Println("Get the list of failed members ", membersFailedCount)
	if membersFailedCount > 0 {
		serfFailedMembers, err := helpers.ListOfMembersByStatus("failed")

		if err != nil {
			log.Println("Error while retrieving failed members from serf RPC.", err.Error())
			return err
		}

		for _, element := range serfFailedMembers {
			if len(element) > 0 {
				log.Println("failed member: ", element)
				removeMember(element)
			}
		}
	}
	return nil
}

func setQuorum(mode string) error {

	err := helpers.SetQuorumRatio(client, mode)
	return err
}

// removeMember: Remove member from serf cluster along with gluster bricks.
func removeMember(nodeName string) error {

	//Set the quorum to zero before performing gluster cleanup.
	err := setQuorum("cleanup")
	if err != nil {
		log.Println("Error while setting server quorum ratio ", err)
		// Do not return back, Continue removing bricks.
	}

	time.Sleep(1 * time.Second)

	//Remove all the gluster bricks associated with the failed node and remove from docker swarm.
	err = glusterCleanup(nodeName)
	if err != nil {
		log.Println("Error while performing gluster cleanup.")
		log.Println(err.Error())
		//do not return back, As gluster bricks have been removed already.
	}

	log.Println("Gluster and Docker node cleanup successful.")

	// reset the quorum once the member is removed from the cluster.
	err = setQuorum("update")
	if err != nil {
		log.Println("Error while setting server quorum ratio ", err)
		// Do not return back, Continue removing member.
	}

	time.Sleep(1 * time.Second)

	err = helpers.MemberForceLeave(nodeName)
	if err != nil {
		log.Println("Error while forcing a member to leave from serf cluster." + nodeName)
		return err
	}

	log.Println("Member " + nodeName + " has left the serf cluster successfully")
	return nil
}

// glusterCleanup: Helper function which removes the gluster peers and associated volumes.
func glusterCleanup(node string) error {

	var glusterHostToDetach string
	glusterHostToDetach, err := helpers.GetMemberIPByName(node)
	if err != nil {
		log.Println("Error while retrieving IP address by host from serf RPC ")
		return err
	}

	if len(glusterHostToDetach) > 0 {

		//remove the gluster bricks.
		err = removeGlusterBrick(glusterHostToDetach)
		if err != nil {
			return err
		}
		log.Println("Gluster brick removal successful for member " + glusterHostToDetach)

		//detach peer from gluster pools.
		err = detachPeer(glusterHostToDetach)
		if err != nil {
			return err
		}
		log.Println("Gluster peer detach successful for member " + glusterHostToDetach)

		//remove the node from docker swarm.
		err = removeNode(node)
		if err != nil {
			return err
		}
		log.Println("Docker Swarm Removal successful" + node)
	}
	return nil
}

// detachPeer: Helper function to detach the peer from gluster.
func detachPeer(host string) error {

	err := client.PeerDetach(host)
	if err != nil {
		log.Println("Error while removing gluster peer. ")
		log.Println(err.Error())
		//dont return back as peer might be already removed during previous iteration of cleaning
	}
	return nil
}

// removeNode: Helper function to demote and remove node from docker swarm.
func removeNode(node string) error {

	nodeID, err := helpers.GetNodeIDByStateAndHostname(node, "down")
	log.Println("nodeID" + nodeID)
	if err != nil {
		log.Println(" Error while checking node id for the failed member in docker swarm", err.Error())
		return err
	}

	if nodeID != "" {
		log.Println("DemoteNode " + nodeID)
		err = helpers.DemoteNode(nodeID)
		if err != nil {
			log.Println("Error in demoting a docker member. ", err.Error())
			//Do not return control back, If a leader goes down from the cluster, Docker will throw the error " Can't find manager in raft member list"
		}
		log.Println("RemoveNode " + nodeID)
		err = helpers.RemoveNode(nodeID, true)
		if err != nil {
			log.Println("Error in demoting a docker member. ", err.Error())
			return err
		}
	}
	return nil
}

// removeDownNode: Helper function to demote and remove node from docker swarm  with duplicate entry
func removeDownNode() error {

	nodeID, err := helpers.GetNodeIDByDownState()
	log.Println("nodeID" + nodeID)
	if err != nil {
		log.Println(" Error while checking node id for the failed member in docker swarm", err.Error())
		return err
	}

	if nodeID != "" {
		log.Println("DemoteNode " + nodeID)
		err = helpers.DemoteNode(nodeID)
		if err != nil {
			log.Println("Error in demoting a docker member. ", err.Error())
			//Do not return control back, If a leader goes down from the cluster, Docker will throw the error " Can't find manager in raft member list"
		}
		log.Println("RemoveNode " + nodeID)
		err = helpers.RemoveNode(nodeID, true)
		if err != nil {
			log.Println("Error in demoting a docker member. ", err.Error())
			return err
		}
	}
	return nil
}

// removeGlusterBrick: Function to remove bricks from gluster volumes.
func removeGlusterBrick(glusterHostToDetach string) error {

	var glusterPeerCount int

	peers, err := client.PeerStatus()
	if err != nil {
		log.Println("Error while checking gluster Peers status", err.Error())
		return err
	}

	glusterPeerCount = len(peers)

	vols, err := client.ListVolumes()
	if err != nil {
		log.Println("Error while checking gluster volumes")
		return err
	}

	if len(vols) > 0 {
		for _, vol := range vols {

			if glusterPeerCount > 0 {
				/*Remove bricks which correspond to glusterHostToDetach in the volumes*/
				for _, brick := range vol.Bricks {
					if strings.Contains(brick, glusterHostToDetach) {
						// In case of Replicated volumes, replica has to be decremented manually while using RemoveBrick
						replica := glusterPeerCount - 1
						log.Println("Removing gluster volume brick")
						err = client.RemoveBrick(vol.Name, brick, replica)
						if err != nil {
							log.Println("Error Removing Brick: ")
							log.Println(err.Error())
						}

					}
				}
			}
		}
	}
	return nil
}
