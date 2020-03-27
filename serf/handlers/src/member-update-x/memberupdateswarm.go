//Copyright 2019, Intel Corporation

package memberupdatex

import (
	"fmt"
	"helpers"
	"io/ioutil"
	"os"
	"time"
)

//If the node is leader, deploy all the apps as in /opt/stacks
// Assumption is every app, will have an docker-compose.yml
func swarmLeaderInitStackDeploy(serfAdvertiseIface string) error {

	if len(serfAdvertiseIface) == 0 {
		return fmt.Errorf("Error IP is not set")
	}

	//Leaving the node if he is part of previous swarm
	//Sleep is mandatory to address case of multiple times leader down.
	time.Sleep(5 * time.Second)
	err := helpers.SwarmLeave(true)
	time.Sleep(10 * time.Second)

	if err != nil {
		rwolog.Error("Failed to leave the node:", err)
		return err
	}

	//no Force
	err = helpers.SwarmInit(serfAdvertiseIface, false)
	if err != nil {
		rwolog.Error("Failed to create the swarm Init:", err)
		return err
	}

	swarmID, err := helpers.GetSwarmNodeID()
	if err != nil {
		rwolog.Error("Failed to docker info: ", err)
		return err
	}

	serfTagsOp := helpers.SetSwarmTag(swarmID)
	if serfTagsOp != nil {
		rwolog.Error("Error in updating tag ", serfTagsOp)
		return serfTagsOp
	}

	serfTagsOp = helpers.DeleteSerfTag("inprocess")
	if serfTagsOp != nil {
		rwolog.Error("Error Deleting, inprocessTag :", serfTagsOp)
		return serfTagsOp
	}

	rwolog.Debug("Swarm id is set and inprocess tag deleted")

	_, err = os.Stat("/opt/stacks")
	if os.IsNotExist(err) {
		rwolog.Debug("/opt/stacks doesn't exist, Nothing to deploy")
		//Dont return error, If user does not want to deploy any stacks, This folder may not exists.
	}

	helpers.StackDeploy()

	return nil
}

//SetSwarmTag to set swarmid
func SetSwarmTag() error {

	rwolog = helpers.GetLogger()

	swarmID, err := helpers.GetSwarmNodeID()
	if err != nil {
		rwolog.Error("Failed to docker info: ", err)
		return err
	}

	rwolog.Debug("Docker Op swarmNodeID: ", swarmID)

	serfTagsOp := helpers.SetSwarmTag(swarmID)
	if serfTagsOp != nil {
		rwolog.Error("Error in updating tag ", serfTagsOp)
		return serfTagsOp
	}

	serfTagsOp = helpers.DeleteSerfTag("inprocess")
	if serfTagsOp != nil {
		return fmt.Errorf("Error Deleting, inprocessTag ", serfTagsOp)
	}

	return nil
}

//JoinSwarm for joining the swarm as worker/manager
func JoinSwarm(serfAdvertiseIFACE string, serfLeader string, role string) error {

	rwolog = helpers.GetLogger()

	_, err := helpers.SerfQuery("swarm-token", role)
	if err != nil {
		rwolog.Error("Error while querying swarm token to leader :", err)
		return err
	}

	swarmToken, err := readToken(role)
	if err != nil {
		rwolog.Error("Error while reading swarm token:", err)
		return err
	}

	err = helpers.SwarmJoin(serfLeader, swarmToken)

	if err != nil {
		rwolog.Error("Failed to Join the swarm")
		return err
	}
	time.Sleep(1 * time.Second)

	_, err = helpers.SerfQuery("docker", "cleanUpStaleMember")
	if err != nil {
		rwolog.Error("Error while querying swarm token to leader :", err)
		return err
	}

	rwolog.Debug("Successfully joined the swarm")
	return nil

}

func readToken(role string) (string, error) {

	managerTokenPath, _ := os.LookupEnv("MANAGER")
	workerTokenPath, _ := os.LookupEnv("WORKER")
	glusterVolumeName, _ := os.LookupEnv("GLUSTER_VOLUME_NAME")
	glusterMountPath, _ := os.LookupEnv("GLUSTER_MOUNT_PATH")
	dockerTokenPath, _ := os.LookupEnv("TOKEN")
	if len(glusterVolumeName) == 0 || len(glusterMountPath) == 0 || len(dockerTokenPath) == 0 || len(managerTokenPath) == 0 || len(workerTokenPath) == 0 {
		fmt.Errorf("Env Variables not defined , ", "GLUSTER_VOLUME_NAME ", glusterVolumeName, " ",
			"GLUSTER_MOUNT_PATH ", glusterMountPath, " ",
			"TOKEN ", dockerTokenPath, " ",
			"MANAGER ", managerTokenPath, " ",
			"WORKER ", workerTokenPath)
	}
	var path string

	if role == "worker" {
		path = glusterMountPath + "/" + glusterVolumeName + "/" + dockerTokenPath + "/" + workerTokenPath + "/token.txt"
	} else {
		path = glusterMountPath + "/" + glusterVolumeName + "/" + dockerTokenPath + "/" + managerTokenPath + "/token.txt"
	}

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(b), err
}

//Join node as a worker and set the tag
func swarmWorker(serfAdvertiseIFACE string, serfLeader string) error {

	rwolog.Debug("Joining Swarm as Worker")

	//Leaving the node if he is part of previous swarm
	err := helpers.SwarmLeave(true)
	if err != nil {
		rwolog.Error("Failed to leave the node:", err)
		return err
	}
	err = JoinSwarm(serfAdvertiseIFACE, serfLeader, "worker")
	if err != nil {
		rwolog.Error("Node failed to join", err.Error())
		return err
	}

	err = SetSwarmTag()
	if err != nil {
		rwolog.Error("Failed to set the tag, SetSwarmTag", err.Error())
		return err
	}

	return nil

}

//If alive members are less then 3, node will join as a leader or reachable
//Before joining the swarm, previous swarm is left forcefully and join
func aliveMemberLE3(serfAdvertiseIFACE string, serfLeader string) error {

	//Get Swarm Node ID.
	swarmID, err := helpers.GetSwarmNodeID()
	if err != nil {
		return fmt.Errorf("Docker empty: ", err)
	}

	swarmManager, err := helpers.CheckIfManager()
	if err != nil {
		return fmt.Errorf("Failed to check manager status")
	}

	var leftSwarm = false

	if len(swarmID) != 0 && swarmManager == false {

		rwolog.Debug("Leave the swarm and rejoin incase if the leader is also changed")

		//Sleep is mandatory to address case of multiple times leader down.
		time.Sleep(5 * time.Second)

		err := helpers.SwarmLeave(true)
		leftSwarm = true
		time.Sleep(10 * time.Second)

		if err != nil {
			return fmt.Errorf("Failed to leave the node ", err)
		}
	}

	rwolog.Debug("Joining Swarm as manager")
	err = JoinSwarm(serfAdvertiseIFACE, serfLeader, "manager")
	if err != nil {
		return fmt.Errorf("Node failed to join")
	}

	if leftSwarm == true {

		rwolog.Debug("Joined swarm by leaving the old docker swarm. Check for docker node with status down for 25 seconds.")
		//Delete the node with status down
		var removeDownNodes []string
		retry := 1
		for retry < 5 {

			removeDownNodes, err = helpers.GetNodeIDByState("down")
			if err != nil {
				return fmt.Errorf("Node Status for down failed ")
			}

			if len(removeDownNodes) != 0 {
				break
			}
			retry++
			time.Sleep(5 * time.Second) // To wait for down node
		}

		for _, removeNode := range removeDownNodes {

			if len(removeNode) != 0 {

				//Check if error to be return
				err := helpers.DemoteNode(removeNode)
				if err != nil {
					rwolog.Error("Failed to demote the node ", err.Error())
				}

				//Check if error to be return
				err = helpers.RemoveNode(removeNode, false)
				if err != nil {
					rwolog.Error("Failed to remove the node", err.Error())
					return err
				}
			}

		}
	}
	err = SetSwarmTag()
	if err != nil {
		rwolog.Error("Failed to set the tag, SetSwarmTag", err.Error())
		return err
	}

	return nil
}

//Assign the status of tag as per the swarm rule.
//Join the swarm as per the swarm rule.
//Swarm Rule: If nodes are less then equal to 3 join as manager else worker
func manageSwarm(serfAdvertiseIface string) error {

	if len(serfAdvertiseIface) == 0 {
		return fmt.Errorf("Error IP is not set ")
	}

	//Create the map
	tags := make(map[string]string)
	tags["role"] = "leader"

	serfLeader, err := helpers.MemberIPByTagsAndStatus(tags, "alive")
	if err != nil {
		return fmt.Errorf("MemberIPByTagsAndStatus Failed, Value of serfLeader: ", string(serfLeader))
	}

	//Debug Print
	rwolog.Debug("MemberIPByTagsAndStatus Passed, value of serfLeader: ", string(serfLeader))

	if len(serfLeader) != 0 {
		rwolog.Debug("Leader Found")

		countAlive := helpers.CountAliveMembers()
		if countAlive <= 3 {

			//Join as a manager/reachable
			err = aliveMemberLE3(serfAdvertiseIface, serfLeader)
			if err == nil {
				return nil
			}
			return fmt.Errorf("aliveMemberLE3 Failed")

		}

		//Join as a worker
		err = swarmWorker(serfAdvertiseIface, serfLeader)
		if err == nil {
			return nil
		}

		return fmt.Errorf("Error in Swarm Worker")
	}

	//Debug Print
	rwolog.Debug("waiting for Leader")
	serfTagsOp := helpers.SetWaitingForLeaderTag("true")
	if serfTagsOp != nil {
		return fmt.Errorf("Error SetWaitingForLeaderTag ", serfTagsOp)

	}
	return nil
}

// Swarm decides weather current member should be a leader, Reachable or worker.
func Swarm() error {

	rwolog = helpers.GetLogger()

	//check for network status
	err := helpers.CheckNetworkStatus()
	if err != nil {
		return err
	}

	//Check if there is something to deployed to stack
	statusRole, _ := os.LookupEnv("SERF_TAG_ROLE")
	rwolog.Debug("serftagRole: ", statusRole)

	if !helpers.IsValidRole(statusRole) {
		rwolog.Debug("SERF Tag from Env is invalid. ", statusRole)
		statusRole = ""
	}

	err = helpers.CheckNetworkStatus()
	if err != nil {
		return err
	}

	serfAdvertiseIface, error := helpers.GetIPAddr()
	if error != nil {
		rwolog.Error("Failed to get network IP\n")
		return error
	}

	if len(serfAdvertiseIface) == 0 {
		return fmt.Errorf("Error IP is not set ")
	}

	if statusRole == "leader" {
		err = swarmLeaderInitStackDeploy(serfAdvertiseIface)
		if err != nil {
			rwolog.Error("swarmLeaderInitStackDeploy Failed")
			return err
		}

		rwolog.Debug("swarmLeaderInitStackDeploy completed")

	} else {
		err = manageSwarm(serfAdvertiseIface)
		if err != nil {
			rwolog.Error("manageSwarm failed")
			return err
		}

		rwolog.Debug("manageSwarm Completed")

	}
	return nil
}
