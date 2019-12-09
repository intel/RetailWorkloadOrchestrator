//Copyright 2019, Intel Corporation

package memberupdatex

import (
	"helpers"
	"os"
	"strings"
	"time"
)

//Global variable to be used across the file
var (
	rwolog *helpers.Logger
)

// Arbiter will help in resolving conflicts when two members with serf tags and docker status shows as "Leader".
//
// This problem can occur in following way.
// Start a node say (member A). This member will create its own swarm and will be a leader.
// Now shutdown the member A.
// Start another node  say (member B). This member will also create its own swarm and will be a leader.
// Now start back the member A.
//
// Since both the members are leaders and have active Docker swarm.
// It is not possible to determine who is the true leader as memberjoin is executed in both the nodes.
//
// Arbiter helps to resolve this issue by starting an arbiter container for the leader who comes up first.
// If there are two leaders. Memberjoin will check whether the arbiter container is running on the node.
// If arbiter is running, Then set the status as leader.
// Else set the tag status as worker and join the existing swarm as manager.
//
// This problem can also be replicated in following way.
// Start a node say (member A) and let it form the cluster.
// Shutdown the node.
// Start another node Say (member B) and let it form Another cluster.
// Start another two nodes and let them join member B cluster.
// Start the earlier member A.
// Now the member A will not join the cluster of member B. Instead it will maintain its own cluster.
func Arbiter() error {
	rwolog = helpers.GetLogger() //For Logging

	rwolog.Info(" ******** Running Arbiter ******** ")

	helpers.WaitForDocker()

	serfRole, _ := os.LookupEnv("SERF_TAG_ROLE")

	if !helpers.IsValidRole(serfRole) {
		rwolog.Error("SERF Tag from Env is invalid. ", serfRole)
		serfRole = ""
	}

	if serfRole == "leader" {
		err := handleAsaLeader()
		if err != nil {
			return err
		}
	} else {
		removeArbiter()
	}

	rwolog.Debug(" ******** Arbiter completed ******** ")
	return nil
}

// handleAsaLeader: manage arbiter with respective to number of alive members
func handleAsaLeader() error {

	rwolog.Debug("SERF_TAG_ROLE set to leader")

	// When second leader joins the cluster, if there is a conflict between two leaders,
	// serf will take time to update the status of second node.
	// Arbiter should start at first leader. So wait until the serf status is updated.
	// In VM's it is observed that serf status is updated after some time.
	membersAliveCount := helpers.CountAliveMembers()
	if membersAliveCount == 1 {
		time.Sleep(5 * time.Second)
		membersAliveCount = helpers.CountAliveMembers()
	}

	// Debug prints to check the number of members for arbiter.
	rwolog.Debug("Members count in the serf cluster ", membersAliveCount)

	// Remove Rwo container if member count is one or greater then 3
	if membersAliveCount > 2 || membersAliveCount == 1 {
		err := removeArbiter()
		if err != nil {
			return err
		}

	} else if membersAliveCount == 2 { //check if Rwo container is running else start one.

		err := updateArbiter()
		if err != nil {
			return err
		}
	}
	return nil
}

// removeArbiter: Helper function to remove arbiter node from swarm.
// If there are two nodes and one of the node went down leading to only one node. We need to make sure that we are demoting and removing the arbiter node from docker swarm.
// Also remove the arbiter container.
func removeArbiter() error {

	// During two node scenario, when leader reboots. There will be two arbiter nodes in the swarm cluster.
	// One with status ready and other with status down.
	// Leader should remove the node with status ready and worker should remove the node with status down.
	serfRole, _ := os.LookupEnv("SERF_TAG_ROLE")
	var state string
	if serfRole == "leader" {
		state = "ready"
	} else {
		state = "down"
	}

	arbiter, err := helpers.GetNodeIDByStateAndHostname("arbiter", state)
	if err != nil {
		rwolog.Error("Error verifying node arbiter node presence.", arbiter, err.Error())
		return err
	}

	if len(arbiter) > 0 {

		// Debug prints
		rwolog.Debug("Arbiter node to be demoted and removed " + arbiter + " with  state " + state)

		err = helpers.DemoteNode(arbiter)
		if err != nil {
			rwolog.Error("Error in Docker demote for arbiter node ", arbiter, err.Error())
			return err
		}

		err = helpers.RemoveNode(arbiter, true)
		if err != nil {
			rwolog.Error("Error in Docker remove for arbiter node ", arbiter, err.Error())
			return err

		}
	}

	// Remove arbiter container if it exists.
	rwoArbiterContainerVal, err := helpers.GetAllSystemDockerNodes("rwo_arbiter")
	if err != nil {
		rwolog.Error("Error while checking arbiter container ID. ", err.Error())
		return err
	}

	if len(rwoArbiterContainerVal) > 0 {

		err := helpers.RemoveSystemDockerNode("rwo_arbiter")
		if err != nil {
			rwolog.Error("Error while removing RWO arbiter container. ", err.Error())
			return err
		}
	}

	return nil

}

// updateArbiter: Check if arbiter is running else join the arbiter container to docker swarm
func updateArbiter() error {

	//Get the Serf members who are alive
	checkRwoArbiterVal, err := helpers.GetSystemDockerNode("rwo_arbiter")
	if err != nil {
		rwolog.Error("Error while checking arbiter container ID. ", err.Error())
		return err
	}

	checkForDockerNode, err := helpers.CheckIfNodeExists("arbiter")
	if err != nil {
		rwolog.Error("Error verifying node arbiter node presence.", checkForDockerNode, err.Error())
		return err
	}

	if len(checkRwoArbiterVal) > 0 && len(checkForDockerNode) > 0 {
		rwolog.Debug("Arbiter exists on this leader.")
	} else {
		tag := make(map[string]string)
		tag["role"] = "leader"

		leaderIPCmd, err := helpers.MemberIPByTagsAndStatus(tag, "alive")
		if err != nil {
			return err
		}

		var leaderAddr = strings.Split(string(leaderIPCmd), ":")[0]
		dockerSwarmJoinTokenVal, err := helpers.GetToken("manager")
		if err != nil {
			rwolog.Error("Error while retriving Docker join token. ", err.Error())
		}

		if leaderAddr != "" && dockerSwarmJoinTokenVal != "" {

			// If The node is shutdown without proper cleanup.
			// Rwo Arbiter container will be running and will not start.
			// Clean up container before running it.
			rwoArbiterContainerVal, err := helpers.GetAllSystemDockerNodes("rwo_arbiter")
			if err != nil {
				rwolog.Error("Error while checking arbiter container ID. ", err.Error())
				return err
			}

			if len(rwoArbiterContainerVal) > 0 {

				err := helpers.RemoveSystemDockerNode("rwo_arbiter")
				if err != nil {
					rwolog.Error("Error while removing RWO arbiter container. ", err.Error())
					return err
				}
			}

			err = helpers.RunArbiterContainer()
			if err != nil {
				rwolog.Error("Error while running rwo arbiter. ", err.Error())
			}

			time.Sleep(2) //wait for rwo arbiter container to come up.

			checkRwoArbiterVal, err = helpers.GetSystemDockerNode("rwo_arbiter")
			if err != nil {
				rwolog.Error("Error while checking arbiter container ID. ", err.Error())
				return err
			}

			// Once the Rwo Arbiter runs successfully, Join the arbiter node to swarm cluster.
			// As ExecuteProcessInNode() expects docker CLI commands for executing docker inside docker.
			// Form a cmd(string array[]) which will be passed as argument.
			var cmd []string
			cmd = append(cmd, "sh")
			cmd = append(cmd, "-c")
			cmd = append(cmd, "docker swarm join --token "+dockerSwarmJoinTokenVal+" "+leaderAddr+":2377")

			_, err = helpers.ExecuteProcessInNode(checkRwoArbiterVal, cmd)
			if err != nil {
				rwolog.Error("Error joining rwo node to the swarm cluster. ", err.Error())
				return err
			}

			cmd[2] = "docker system info --format \"{{.Swarm.NodeID}}\""
			dockerSwarmIDVal, err := helpers.ExecuteProcessInNode(checkRwoArbiterVal, cmd)
			if err != nil {
				rwolog.Error("Error in getting docker swarm id for the rwo arbiter container. ", err.Error())
				return err
			}

			var dockerSwarmID = strings.TrimSpace(string(dockerSwarmIDVal))
			err = helpers.UpdateNodeAvailabilityDrain(dockerSwarmID)
			if err != nil {
				rwolog.Error("Error while setting Arbiter node status to Drain. ", err.Error())
				return err
			}

		} else {
			rwolog.Debug("Leader address/Docker join token is empty.")
			return err
		}

	}
	return nil
}
