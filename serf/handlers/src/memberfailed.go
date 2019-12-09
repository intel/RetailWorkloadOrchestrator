//Copyright 2019, Intel Corporation

package main

import (
	"fmt"
	"helpers"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Global variables
var serfRole string
var newCluster bool

//Global variable to be used across the file
var (
	rwolog *helpers.Logger
)

// InitHandler initilaises handler
func InitHandler() error {
	rwolog = helpers.GetLogger()
	return nil
}

func main() {
	serfRole, _ := os.LookupEnv("SERF_TAG_ROLE")

	// Initialise handlers
	InitHandler()

	// Check network status
	if helpers.CheckNetworkStatus() != nil {
		return
	}

	rwolog.Info("******** Running serf member-failed ********")
	rwolog.Info("Member failed, sleeping for 5 min in case the member is rebooting.")

	//Get the alive members count as member failed has been executed. This will be helpfull if a member is going through a reboot and only one member is alive.
	aliveMembers := helpers.CountAliveMembers()

	_, err := helpers.GetIPAddr()
	if err != nil {
		rwolog.Error("Network is not up")
		return
	}

	if !helpers.IsValidRole(serfRole) {
		rwolog.Debug("SERF Tag from Env is invalid. ", serfRole)
		serfRole = ""
	}

	if serfRole != "leader" {

		// Swarm will take time to assign new leader.
		waitForDockerSwarmRestore()

		if aliveMembers == 1 {
			err := handleSwarmAsLeader()
			if err != nil {
				rwolog.Error(err.Error())
			}
		} else {
			err := handleSwarmAsReachable(aliveMembers)
			if err != nil {
				rwolog.Error(err.Error())
			}
		}

	} else if serfRole == "leader" {

		if aliveMembers == 1 {
			waitForDockerSwarmRestore()
			err = manageSwarm()
			if err != nil {
				return
			}
		}

		err = performCleanup()
		if err != nil {
			rwolog.Error("Error while  forking member cleanup.", err.Error())
			return
		}
	}

	if newCluster == true {
		helpers.StackDeploy()
	}
}

// waitForDockerSwarmRestore will wait for specified time for docker swarm to update its status in the cluster.
func waitForDockerSwarmRestore() {

	delaySecondsFromEnv, _ := os.LookupEnv("MEMBER_REBOOT_TIME")
	var delaySeconds int
	if len(delaySecondsFromEnv) > 0 {
		delaySeconds, _ = strconv.Atoi(delaySecondsFromEnv)
		//check the delay should not be more then 3600 seconds.
		if delaySeconds > 3600 {
			delaySeconds = 30
		}
	} else {
		rwolog.Debug("MEMBER_REBOOT_TIME is not defined in the docker compose. Setting it to 30 seconds.")
		delaySeconds = 30
	}

	rwolog.Debug("Sleeping for ", strconv.Itoa(delaySeconds))
	time.Sleep(time.Duration(delaySeconds) * time.Second)
}

// handleAsLeader: When a member fails in the Docker swarm. Docker will promote a worker to a leader.
// Check if the current member is the promoted leader. If yes then perform node cleanup.
func handleSwarmAsReachable(aliveMembers int) error {

	inspectSelfForStatusOut, err := helpers.GetNodeStatus("leader")
	if err != nil {
		return err
	}

	if inspectSelfForStatusOut == true {
		helpers.SetRoleTag("leader")
		rwolog.Debug("Current node is promoted as leader")
		err = performCleanup()
		if err != nil {
			rwolog.Error("Error while  forking member cleanup.", err.Error())
			return err
		}

	} else {
		//This condition will be executed when 3 members are alive and current node is not a leader/reachable.
		aliveMembers = helpers.CountAliveMembers()
		if aliveMembers == 3 {
			err = handleWorker()
			if err != nil {
				rwolog.Error("Error in executing worker role.")
				return err
			}
		}
	}
	return nil
}

// handleSwarm: Function which initialize the swarm and perform member clean up.
func handleSwarmAsLeader() error {

	err := initializeSwarm()
	if err != nil {
		return err
	}

	err = setSwarmTag()
	if err != nil {
		rwolog.Error("Error while setting up tag as leader.")
		return err
	}

	err = performCleanup()
	if err != nil {
		rwolog.Error("Error while  forking member cleanup.", err.Error())
		return err
	}

	return nil
}

// manageSwarm: This function will check arbiter container. if not found then perform member cleanup.
func manageSwarm() error {

	arbiterContainer, err := helpers.GetSystemDockerNode("rwo_arbiter")
	if err != nil {
		rwolog.Error("Error while checking arbiter container ID.", err.Error())
		return nil
	}

	arbiterNode, err := helpers.GetNodeIDByStateAndHostname("arbiter", "ready")
	if err != nil {
		rwolog.Error("Error while checking arbiter node ID.", err.Error())
		return nil
	}

	if len((arbiterContainer)) > 0 && len(string(arbiterNode)) > 0 {
		rwolog.Debug("The arbiter exists on this node.")
	} else {
		err = initializeSwarm()
		if err != nil {
			return err
		}

		err = setTagAsNodeID()
		if err != nil {
			return fmt.Errorf("error while setting up the tags roles for swarm with swarm node id")
		}
	}
	return nil
}

// setTagAsNodeID: Set the swarm tag with swarm node id as value.
func setTagAsNodeID() error {

	swarmID, err := helpers.GetSwarmNodeID()
	if err != nil {
		rwolog.Error("Error while checking docker swarm node ID.", err.Error())
		return err
	}

	err = helpers.SetSwarmTag(swarmID)
	if err != nil {
		return err
	}
	return nil
}

// handleWorker: Handles worker role to set appropriate swarm tags
func handleWorker() error {

	dockerSwarmID := checkSwarmID()
	isReachable, err := helpers.CheckIfManager() //check if the current node is a manager/reachable.

	// On Slow Machines, DockerD may not response for swarm ID due to swarm restoration.
	// Retry if the docker has returned empty response.
	if err != nil {
		time.Sleep(15 * time.Second)
		isReachable, _ = helpers.CheckIfManager()
	}

	if len(dockerSwarmID) > 0 && isReachable == false {
		err := helpers.SetSwarmTag("")
		if err != nil {
			return err
		}
	} else {
		return nil
	}

	return nil
}

// checkSwarmID: check system swarm ID is present in leader swarm by doing swarm query.
func checkSwarmID() string {

	swarmIDFromLeader, _ := helpers.SerfQuery("docker", "node ls")
	rwolog.Debug("swarmIDFromLeader ", swarmIDFromLeader)
	if swarmIDFromLeader != "" {

		swarmIDFromDocker, _ := helpers.GetSwarmNodeID()

		if strings.Contains(swarmIDFromLeader, swarmIDFromDocker) {
			rwolog.Debug("Swarm ID is present in with the leader")
			return swarmIDFromDocker
		}
	}

	rwolog.Debug("Current Swarm ID is not present with the leader")
	return ""
}

// initializeSwarm: Function which will force a new cluster and perform clean up.
func initializeSwarm() error {

	inspectSelfForStatusOut, _ := helpers.GetNodeStatus("leader")

	if inspectSelfForStatusOut == true {
		rwolog.Debug("Current node is already a leader in swarm.")

	} else {
		err := swarmInit()
		if err != nil {
			return err
		}
	}
	err := helpers.SetRoleTag("leader")
	if err != nil {
		return err
	}

	return nil
}

// swarmInit: Function forces a swarm cluster.
func swarmInit() error {

	err := helpers.SwarmLeave(true)
	if err != nil {
		rwolog.Error("Error in leaving the swarm  ", err.Error())
	}
	time.Sleep(3 * time.Second) //Wait for docker swarm to be up.

	serfAdvertiseIFace, err := helpers.GetIPAddr()
	if err != nil {
		rwolog.Error("Network is not up")
		return err
	}

	err = helpers.SwarmInit(serfAdvertiseIFace, true)
	if err != nil {
		rwolog.Error("Error in swarm init ", err.Error())
	}

	// After Docker Swarm init. DockerD will take time to update swarm on slow machines.
	// This delay will ensure that DockerD is stable.
	time.Sleep(10 * time.Second)
	newCluster = true

	return nil
}

// performCleanup: Function which calls helper function to remove the member from Serf cluster and Docker swarm.
func performCleanup() error {

	rwolog.Debug("Forking Member cleanup process.")
	var filePathForCleanup string
	filePathForCleanup = "/opt/rwo/serf/handlers/bin/membercleanup"
	if helpers.Exists(filePathForCleanup) {
		cmd := exec.Command(filePathForCleanup)
		err := cmd.Start()
		if err != nil {
			rwolog.Error(err.Error())
		}
		rwolog.Debug("Member failed completed.")

	} else {
		return fmt.Errorf("Member Cleanup Binary was not found in path %v", filePathForCleanup)
	}

	return nil
}

// setSwarmTag: Helper function to set serf swarm node ID.
func setSwarmTag() error {

	err := setTagAsNodeID()
	if err != nil {
		return err
	}

	err = helpers.SetRoleTag("leader")
	if err != nil {
		return err
	}

	return nil
}
