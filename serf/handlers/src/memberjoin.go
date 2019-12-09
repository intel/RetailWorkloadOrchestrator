//Copyright 2019, Intel Corporation

package main

import (
	"fmt"
	"helpers"
	"math/rand"
	"memberupdatex"
	"os"
	mg "rwogluster"
	"strconv"
	"strings"
	"time"
)

var (
	client      *mg.Client
	serfContext *SerfContext
	rwolog      *helpers.Logger
)

// SerfContext stores tags for gluster and swarm
type SerfContext struct {
	GlusterTag string
	SwarmTag   string
}

// InitHandler reads the environment variables set by serf agent while executing the handler
func InitHandler() error {

	// read Serf Environment variables
	statusGluster, _ := os.LookupEnv("SERF_TAG_GLUSTER")
	statusSwarm, _ := os.LookupEnv("SERF_TAG_SWARM")

	serfContext = &SerfContext{
		GlusterTag: statusGluster,
		SwarmTag:   statusSwarm,
	}

	rwolog = helpers.GetLogger()
	return nil
}

//checkGlusterStatus will check if node is part of
//any swarm or gluster and update the tags accordingly
func checkGlusterStatus() error {

	peers, err := client.PeerStatus()
	if err != nil {
		rwolog.Error("Error in Gluster Peer Status")
	}

	glusterListCount := len(peers)

	rwolog.Debug("value of gluster list count", glusterListCount)

	var glusterUUID string
	if glusterListCount >= 2 {
		for _, peer := range peers {
			if peer.Name == "localhost" {
				glusterUUID = peer.ID
			}
		}

		rwolog.Debug("Gluster UUID is ", glusterUUID)

		glusterPL, err := helpers.SerfQuery("gluster", "pool list")
		if err != nil {
			rwolog.Error("Serf query for gluster pool list to the leader has failed: ", glusterPL)
			return err
		}

		//If gluster is part of previous swarm, return
		if strings.Contains(glusterPL, glusterUUID) {
			rwolog.Error("Current Member is a part of existing Gluster, Gluster Tag will not be removed.")

			// Perform volume mount.
			err = helpers.MountGlusterVolumes()
			if err != nil {
				return err
			}
			return nil
		}

		serfTagsOp := helpers.SetInitTag(strconv.Itoa(int(time.Now().UnixNano())))
		if serfTagsOp != nil {
			rwolog.Error("Error of in update Tag, Init: ", serfTagsOp)
			return serfTagsOp
		}

		//Delete gluster tag
		serfTagsOp = helpers.DeleteSerfTag("gluster")
		if serfTagsOp != nil {
			rwolog.Error("Error of in Deleting Tag, gluster: ", serfTagsOp)
			return serfTagsOp
		}

		//Delete glusterretry tag
		serfTagsOp = helpers.DeleteSerfTag("glusterretry")
		if serfTagsOp != nil {
			rwolog.Error("Error of in Deleting Tag, glusterretry: ", serfTagsOp)
			return serfTagsOp
		}

	} else {

		//Delete gluster tag
		serfTagsOp := helpers.DeleteSerfTag("gluster")
		if serfTagsOp != nil {
			rwolog.Error("Error of in Deleting Tag, gluster: ", serfTagsOp)
			return serfTagsOp
		}

		//Delete glusterretry tag
		serfTagsOp = helpers.DeleteSerfTag("glusterretry")
		if serfTagsOp != nil {
			rwolog.Error("Error of in Deleting Tag, glusterretry: ", serfTagsOp)
			return serfTagsOp
		}

	}

	return nil

}

//checkSwarmStatus will check if node is part of
//any swarm or gluster and update the tags accordingly
func checkSwarmStatus() error {

	rwolog.Debug("If the node is rebooted, Serf query will take time as leader might be executing member-failed. Please wait..")
	serfOutput, err := helpers.SerfQuery("docker", "node ls")
	if err != nil {
		rwolog.Error("Error in executing docker node ls")
		return err
	}

	//Get Swarm Node ID.
	swarmID, err := helpers.GetSwarmNodeID()
	if err != nil {
		rwolog.Error("Docker empty: ", err)
		return err
	}

	if len(swarmID) != 0 && strings.Contains(serfOutput, swarmID) {
		rwolog.Debug("Current Member is a part of existing swarm-worker, Swarm Tag will not be removed.")
		return nil
	}

	//Set tag Init
	serfTagsOp := helpers.SetInitTag(strconv.Itoa(int(time.Now().UnixNano())))
	if serfTagsOp != nil {
		rwolog.Error("Error in update Init tag: ", serfTagsOp)
		return serfTagsOp
	}

	//Delete tag Swarm
	serfTagsOp = helpers.DeleteSerfTag("swarm")
	if serfTagsOp != nil {
		rwolog.Error("Error in deleting init tag: ", serfTagsOp)
		return serfTagsOp
	}

	return nil
}

// manageLeader will check if ManagerStatus is leader; check Swarm/Gluster status
// else if there exists any rwo_arbiter remove it as there are two leaders;
// & set the node as worker; check Swarm/Gluster status
func manageLeader() error {

	inspectManagerStatus, err := helpers.GetNodeStatus("leader")
	if err != nil {
		rwolog.Error("Failed to get the leader status ", err.Error())
		return err
	}

	rwolog.Debug("grep & stop rwo_arbiter if it exists")

	containerID, err := helpers.GetSystemDockerNode("rwo_arbiter")
	if err != nil {
		rwolog.Error("Failed to get container ID for rwo_arbiter", err.Error())
		return err
	}

	if inspectManagerStatus == true {

		if len(containerID) == 0 {

			//Sleep is mandatory to address case of multiple times leader down.
			time.Sleep(5 * time.Second)
			err := helpers.SwarmLeave(true)
			if err != nil {
				rwolog.Error("Failed to leave the node:", err)
				return err
			}

			time.Sleep(10 * time.Second)

		} else {

			rwolog.Debug("I am already the leader")
			return nil
		}
	}

	if len(containerID) != 0 {

		err := helpers.RemoveSystemDockerNode(containerID)
		if err != nil {
			rwolog.Error("Failed to remove rwo_arbiter ", err.Error())
			return err
		}
	}

	serfTagsOp := helpers.SetRoleTag("worker")
	if serfTagsOp != nil {
		rwolog.Error("Error in update tag ", serfTagsOp)
		return serfTagsOp
	}

	err = checkSwarmStatus()
	if err != nil {
		rwolog.Error("Swarm Status Failed")
		return err
	}

	err = checkGlusterStatus()
	if err != nil {
		rwolog.Error("Gluster Status Failed")
		return err
	}

	return nil

}

//checkSwarmAndGlusterStatus, check the Gluster/Swarm Status
func checkSwarmAndGlusterStatus() error {

	err := checkSwarmStatus()
	if err != nil {
		rwolog.Error("Swarm Status Failed ", err.Error())
		return err
	}
	err = checkGlusterStatus()
	if err != nil {
		rwolog.Error("Gluster Status Failed ", err.Error())
		return err
	}

	return nil
}

//memberJoinAssignLeaderWorker Assign a tag to a fresh node
func memberJoinAssignLeaderWorker() error {

	CountMembers := helpers.CountAliveMembers()
	rwolog.Debug("Alive members are:", CountMembers)

	if CountMembers <= 1 {

		serfTagsOp := helpers.SetRoleTag("leader")
		if serfTagsOp != nil {
			rwolog.Error("Error in update tag, Role ", serfTagsOp)
			return serfTagsOp
		}

		//setInit Time
		serfTagsOp = helpers.SetInitTag(strconv.Itoa(int(time.Now().UnixNano())))
		if serfTagsOp != nil {
			rwolog.Error("Error in update tag, Init ", serfTagsOp)
			return serfTagsOp
		}

		serfTagsOp = helpers.SetInProcessTag("true")
		if serfTagsOp != nil {
			rwolog.Error("Error in update tag, Inprocess ", serfTagsOp)
			return serfTagsOp
		}

	} else {

		//setInit Time
		serfTagsOp := helpers.SetInitTag(strconv.Itoa(int(time.Now().UnixNano())))
		if serfTagsOp != nil {
			rwolog.Error("Error in update tag, Init ", serfTagsOp)
			return serfTagsOp
		}

		//Set Worker Tag
		serfTagsOp = helpers.SetRoleTag("worker")
		if serfTagsOp != nil {
			rwolog.Error("Error in update tag, Role ", serfTagsOp)
			return serfTagsOp
		}

		//Set InProcess Tag
		serfTagsOp = helpers.SetInProcessTag("true")
		if serfTagsOp != nil {
			rwolog.Error("Error in update tag, Inprocess ", serfTagsOp)
			return serfTagsOp
		}

	}

	return nil

}

//memberJoinManageLeaderWorker manages the leader/Worker status
func memberJoinManageLeaderWorker() error {

	statusRole, _ := os.LookupEnv("SERF_TAG_ROLE")

	if !helpers.IsValidRole(statusRole) {
		rwolog.Error("SERF Tag from Env is invalid. ", statusRole)
		statusRole = ""
	}

	inspectManagerStatus, err := helpers.GetNodeStatus("leader")
	if err != nil {
		rwolog.Error("Failed to get the leader status ", err.Error())
		return err
	}

	// If Leader reboots, The swarm status will be updated as reachable.
	// Since it is a reboot, Member-failed will be executed and will wait for 500 seconds.
	// after that a new leader will be assigned.
	// Memberjoin should ensure that Current node waits until leader is assigned in serf cluster.
	if statusRole == "leader" && inspectManagerStatus == false {
		rwolog.Debug("Since the role tag specifies this node is a serf leader but not a swarm leader due to a node reboot.")
		serfTagsOp := helpers.SetRoleTag("worker")
		if serfTagsOp != nil {
			rwolog.Error("Error in update tag, Role ", serfTagsOp)
			return serfTagsOp
		}
	}

	countLeaderMembers := helpers.CountLeaders()
	rwolog.Debug("Number of Leader members are:", countLeaderMembers)

	aliveMembers := helpers.CountAliveMembers()

	//Resolve conflicts between two leaders.
	if statusRole == "leader" && countLeaderMembers > 1 && aliveMembers == 2 {

		//Incase if arbiter is not started. Start the arbiter. if the node is true leader in docker swarm.

		if inspectManagerStatus == true {

			rwolog.Debug("There are two leaders and current node docker status specifies that this is the true leader. check arbiter for its coverage.")
			err := memberupdatex.Arbiter()
			if err != nil {
				rwolog.Debug("Error while executing arbiter.")
			}
		}

		status := manageLeader()
		if status == nil {
			rwolog.Debug("Swarm Status And Gluster Status Managed for leader")
			return nil
		}
	}

	if inspectManagerStatus == true {
		rwolog.Debug("I am already the leader")
		return nil
	}

	err = checkSwarmAndGlusterStatus()
	if err == nil {
		rwolog.Error("Swarm And gluster Status handled")
		return nil
	}

	return fmt.Errorf("error in handling swarm and gluster status")

}

func handleRoleAndSwarm() error {

	serfTagsOp := helpers.SetRoleTag("leader")
	if serfTagsOp != nil {
		rwolog.Error("Error in update tag, Role ", serfTagsOp)
		return serfTagsOp
	}

	//setInit Time
	serfTagsOp = helpers.SetInitTag(strconv.Itoa(int(time.Now().UnixNano())))
	if serfTagsOp != nil {
		rwolog.Error("Error in update tag, Init ", serfTagsOp)
		return serfTagsOp
	}

	serfTagsOp = helpers.SetInProcessTag("true")
	if serfTagsOp != nil {
		rwolog.Error("Error in update tag, Inprocess ", serfTagsOp)
		return serfTagsOp
	}

	//Delete tag Swarm
	serfTagsOp = helpers.DeleteSerfTag("swarm")
	if serfTagsOp != nil {
		rwolog.Error("Error in deleting swarm tag: ", serfTagsOp)
		return serfTagsOp
	}

	//Delete gluster tag
	serfTagsOp = helpers.DeleteSerfTag("gluster")
	if serfTagsOp != nil {
		rwolog.Error("Error of in Deleting Tag, gluster: ", serfTagsOp)
		return serfTagsOp
	}

	//Delete glusterretry tag
	serfTagsOp = helpers.DeleteSerfTag("glusterretry")
	if serfTagsOp != nil {
		rwolog.Error("Error of in Deleting Tag, glusterretry: ", serfTagsOp)
		return serfTagsOp
	}

	return nil
}

func main() {

	r := rand.Intn(100) //random sleep
	time.Sleep(time.Duration(r) * time.Microsecond)

	// Initialise handlers
	InitHandler()

	//check for network status
	if helpers.CheckNetworkStatus() != nil {
		rwolog.Error("Error: IP is not assigned.")
		return
	}

	//check for Docker status
	if helpers.WaitForDocker() != true {
		rwolog.Error("Error: DockerD is not running. Please restart the node.")
		return
	}

	// Create  a new client
	client = helpers.GlusterLibClient()

	//Fresh state Assign leader/Worker
	if len(serfContext.GlusterTag) == 0 && len(serfContext.SwarmTag) == 0 {
		err := memberJoinAssignLeaderWorker()
		if err == nil {
			rwolog.Debug("Leader or worker is assigned")
		} else {
			rwolog.Error("Failed in assigning Leader or worker ", err.Error())
		}

	} else if len(serfContext.GlusterTag) != 0 || len(serfContext.SwarmTag) != 0 {

		CountMembers := helpers.CountAliveMembers()

		rwolog.Debug("Alive members are:", CountMembers)

		if CountMembers == 1 {
			err := handleRoleAndSwarm()
			if err != nil {
				rwolog.Error("Failed to Manage Leader node ", err.Error())
			}
			return
		}

		rwolog.Debug("Manage the leader/worker")

		err := memberJoinManageLeaderWorker()
		if err != nil {
			rwolog.Error("Failed to Manage the node ", err.Error())
			return
		}

	}

}
