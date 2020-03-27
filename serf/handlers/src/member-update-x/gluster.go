//Copyright 2019, Intel Corporation

package memberupdatex

import (
	"fmt"
	"helpers"
	"os"
	"os/exec"
	mg "rwogluster"
	"strconv"
	"strings"
	"time"
)

var (
	client                                                *mg.Client
	glusterVolumeName, glusterBrickPath, glusterMountPath string
	glusterDockerVolumePath, rwoBasePath, dockerTokenPath string
	managerTokenPath, workerTokenPath                     string
	serfTagGlusterRetry                                   int
)

// Init initialises global variables
func Init() error {
	ok := true
	glusterVolumeName, ok = os.LookupEnv("GLUSTER_VOLUME_NAME")
	if !ok {
		return fmt.Errorf("GLUSTER_VOLUME_NAME empty")
	}

	glusterBrickPath, ok = os.LookupEnv("GLUSTER_BRICK_PATH")
	if !ok {
		return fmt.Errorf("GLUSTER_BRICK_PATH empty")
	}

	glusterMountPath, ok = os.LookupEnv("GLUSTER_MOUNT_PATH")
	if !ok {
		return fmt.Errorf("GLUSTER_MOUNT_PATH empty")
	}

	glusterDockerVolumePath, ok = os.LookupEnv("GLUSTER_DOCKER_VOLUME_PATH")
	if !ok {
		return fmt.Errorf("GLUSTER_DOCKER_VOLUME_PATH empty")
	}

	rwoBasePath, ok = os.LookupEnv("RWO_BASE_PATH")
	if !ok {
		return fmt.Errorf("RWO_BASE_PATH")
	}

	tagRetry, _ := os.LookupEnv("SERF_TAG_GLUSTERRETRY")

	serfTagGlusterRetry, _ = strconv.Atoi(tagRetry)

	dockerTokenPath, ok = os.LookupEnv("TOKEN")
	if !ok {
		return fmt.Errorf("TOKEN")
	}

	managerTokenPath, ok = os.LookupEnv("MANAGER")
	if !ok {
		return fmt.Errorf("MANAGER")
	}

	workerTokenPath, ok = os.LookupEnv("WORKER")
	if !ok {
		return fmt.Errorf("WORKER")
	}

	return nil
}

// Gluster handles glusterfs specific operations.
//
// When a node comes up. If it is only node in the serf cluster, then gluster will check for any existing bricks and volumes present with it.
// If there are any, Then gluster will remove bricks of other disconnected peers and detach from them.
// Volumes will not be deleted as there might me some data which can be carry forwarded.
//
// If there are not any bricks or volumes, gluster will create a new volumes.
// All the volumes will be started.
//
// If there are more then one nodes in the serf cluster. then gluster will check for any existing bricks and volumes present with it.
// If there are any, Then gluster will remove bricks of other disconnected peers and detach from them. Also deletes the volumes.
// Then a serf query will be made to the leader to perform the peer probe the current node.
// Once the probe is successfull. gluster will add the its own bricks to the volume and mount's it.
//
// Issues:
// * Volume stop or delete fail: gluster will not allow volume operations if there are any existing ongoing transactions.
// gluster recommends to retry the volume operations after a while.
//
// * Peer Probe Fail: leader tries to do peer probe the worker. But it fails with the error "node has some volumes already configured"
// This error signifies that gluster cleanup is not successfull. gluster will start cleanup once again by calling gluster cleanup function.
//
// Global variables: These variables are environment variables which are fixed and are used across multiple functions.
func Gluster() error {

	rwolog = helpers.GetLogger()

	rwolog.Info(" ******** Running Gluster ******** ")

	// read gloabl variables
	err := Init()
	if err != nil {
		rwolog.Error("Some ENV variables are assigned. Error: ", err)
		return err
	}

	// Create  a new client
	client = helpers.GlusterLibClient()

	//check for network status
	err = helpers.CheckNetworkStatus()
	if err != nil {
		return err
	}

	var glusterClusterAddr string
	var membersCount int

	err = createDirectories()
	if err != nil {
		return err
	}

	err = waitForGlusterContainer()
	if err != nil {
		rwolog.Error("Unable to retrieve docker container status. Exiting")
		return err
	}

	glusterClusterAddr, err = helpers.GetIPAddr()
	if err != nil {
		rwolog.Error("Failed to get network IP:", err)
		return err
	}

	// Get the Serf members who are alive or left
	membersCount = helpers.CountAliveOrLeftMembers()
	if membersCount == 1 {

		//check if there are any existing disconnected peers and there volumes.
		//If Yes, then remove bricks and detach peers.
		//This will ensure that data in the brick remains as volume delete is not done.

		err = checkExistingState()
		if err != nil {
			return err
		}

		//Create a gluster volume only if volume does not exists.
		//Else gluster will throw error while creating.
		vols, err := client.ListVolumes()
		if err != nil {
			rwolog.Error("Error while listing volumes ", err)
			return err
		}
		rwolog.Debug("Volume count is ", len(vols))
		if len(vols) == 0 {
			err = createGlusterVolume(glusterClusterAddr)
			if err != nil {
				return err
			}
		}

		vols, err = client.ListVolumes()
		rwolog.Debug("Volume count is ", len(vols))
		//Start the gluster volume.
		for _, vol := range vols {

			rwolog.Debug("Starting the volume ", vol.Name)
			startGlusterVolume(vol.Name)
		}

	} else {
		//Check the existing status of gluster and manage volumes accordingly
		err = manageGlusterVolumes(glusterClusterAddr)
		if err != nil {
			return err
		}
	}

	// Set the server quorum ratio.
	err = SetServerQuorumRatio()
	if err != nil {
		return err
	}

	// Perform volume mount.
	err = helpers.MountGlusterVolumes()
	if err != nil {
		return err
	}
	// Set serf tags.
	err = setTag()
	if err != nil {
		return err
	}

	return nil
}

// createDirectories: Function to create gluster specific directories.
func createDirectories() error {

	var brickPath = glusterBrickPath + "/" + glusterVolumeName
	var mountPath = glusterMountPath + "/" + glusterVolumeName
	var volumePath = glusterDockerVolumePath

	dirPath := []string{brickPath, mountPath, volumePath}

	for i := 0; i < len(dirPath); i++ {
		if helpers.Exists(dirPath[i]) == false {
			err := helpers.CreateDir(dirPath[i])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// waitForGlusterContainer:  Method performs system call check for gluster container.
func waitForGlusterContainer() error {

	//Check if docker gluster container is up and running
	for {
		glusterServerContainerVal, err := helpers.GetSystemDockerNode("gluster-server")
		if err != nil {
			rwolog.Error("Error in checking docker gluster container for status ", err.Error())
			return err
		}

		if len(glusterServerContainerVal) > 0 {
			break
		} else {
			rwolog.Debug("Sleeping for 10 seconds to get gluster docker container up")
			time.Sleep(10 * time.Second)
		}
	}
	return nil
}

// manageGlusterVolumes: Main method responsible mainting the gluster volumes.
func manageGlusterVolumes(glusterClusterAddr string) error {

	err := checkExistingState()
	if err != nil {
		return err
	}

	err = manageVolumes(glusterClusterAddr)
	if err == nil {
		handleAppsVolumes(glusterClusterAddr)
	} else {
		return err
	}
	return nil
}

// checkExistingState checks existing peers or volumes.
func checkExistingState() error {

	// In case any previous volumes are present
	// clean up is needed

	vols, err := client.ListVolumes()
	if err != nil {
		rwolog.Error("Error while listing volumes ", err)
		return err
	}

	peers, err := client.PeerStatus()
	if err != nil {
		rwolog.Error("Error while checking gluster Pool list", err.Error())
		return err
	}

	if len(vols) >= 1 || len(peers) >= 1 {
		// check if the current member has any previous volumes/peers associated with it.
		// if yes, then check if the member is joining previous cluster,
		// else remove the volumes and join the cluster
		err = checkOldStateOfGluster()
		if err != nil {
			return err
		}

	}

	return nil
}

// checkOldStateOfGluster: Determine the gluster status of current member with gluster pool list.
// if the member is part of existing cluster, just set the GUID as serf tag for gluster.
// else perform peer detach by removing all the existing bricks.
// and delete the associated volumes so that current member can join the new cluster of volumes in case of two nodes.
func checkOldStateOfGluster() error {

	var glusterUUID string
	var partofExistingGluster string

	peers, err := client.PeerStatus()
	if err != nil {
		rwolog.Error("Error Getting Gluster Peer Status")
		return err
	}
	for _, peer := range peers {
		if strings.Contains(peer.Name, "localhost") {
			glusterUUID = peer.ID
		}
	}

	var retryCount int

	aliveMembers := helpers.CountAliveMembers()
	if aliveMembers == 1 {
		err = purgeOldStateOfGluster()
		if err != nil {
			return err
		}
		vols, _ := client.ListVolumes()
		// Check if there is any IP Change in the gluster.
		var ipChanged bool
		ipChanged = true

		for _, vol := range vols {

			// Iterate over bricks and remove bricks but one
			for _, brick := range vol.Bricks {
				// since peers are disconnected, we delete bricks individually
				ourIP, err := helpers.GetIPAddr()
				if err != nil {
					rwolog.Error("Failed to get network IP:", err)
					return err
				}

				if strings.Contains(brick, ourIP) {
					ipChanged = false
				}
			}
		}

		// incase ip has been changed then we need to remove the volumes and create the volumes.
		if ipChanged == true {
			rwolog.Debug("Ip has been changed, Deleting the volumes.")
			err = deleteGlusterVolumes()
			if err != nil {
				return err
			}
		}
	} else {

		for {
			checkForPoolList, _ := helpers.SerfQuery("gluster", "pool list")
			partofExistingGluster = strings.TrimSuffix(string(checkForPoolList), "\n")
			if len(partofExistingGluster) == 0 && retryCount < 5 {
				rwolog.Debug("Gluster pool list response is empty from the server. Retrying")
				time.Sleep(15 * time.Second)
				retryCount++
			} else {
				break
			}
		}

		if len(partofExistingGluster) > 0 {
			if strings.Contains(string(partofExistingGluster), glusterUUID) {
				err = helpers.SetGlusterTag(glusterUUID)
				if err != nil {
					rwolog.Error("Error while setting up serf tag for gluster ")
					return err
				}
			} else {
				err = purgeOldStateOfGluster()
				if err != nil {
					return err
				}
				err = deleteGlusterVolumes()
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// purgeOldStateOfGluster removes all bricks of gluster and performs peer detach.
// Check the peers status, if the status of the peers are disconnected,
// Then stop the volumes. Iterate over each of the volumes and remove the bricks.
// Once all the bricks are removed, perform peer detach.
func purgeOldStateOfGluster() error {

	peers, err := client.PeerStatus()
	if err != nil {
		rwolog.Error("Error while checking gluster Pool list", err.Error())
		return err
	}

	var otherPeers []string
	var connectedPeers []string

	// Debug Prints
	rwolog.Debug("purgeOldStateOfGluster: Peers List ", peers)
	for _, peer := range peers {
		if peer.Name != "localhost" {
			otherPeers = append(otherPeers, peer.Name)

			if peer.Status == "CONNECTED" {
				connectedPeers = append(connectedPeers, peer.Name)
			}
		}
	}

	if len(connectedPeers) == 0 {
		// All peers are disconnected.
		// Remove individual bricks and delete the volume
		// Iterate and do peer detach

		vols, err := client.ListVolumes()
		if err != nil {
			rwolog.Error("Error while List gluster volumes", err.Error())
			return err
		}

		for _, vol := range vols {

			//Set the quorum ratio before cleanup.
			err = helpers.SetQuorumRatio(client, "cleanup")
			if err != nil {
				rwolog.Error("Error while setting server quorum ratio ", err)
			}

			stopVolumeRetry(vol.Name)

			// Iterate over bricks and remove bricks but one
			replica := len(vol.Bricks) - 1
			for _, brick := range vol.Bricks {
				// since peers are disconnected, we delete bricks individually
				ourIP, err := helpers.GetIPAddr()
				if err != nil {
					rwolog.Error("Failed to get network up:", err)
					return err
				}

				if !strings.Contains(brick, ourIP) {
					if replica > 0 {
						rwolog.Debug("Remove Brick ", brick, " with replica ", replica)

						err = rmBricksFromVol(vol.Name, brick, replica)
						if err != nil {
							rwolog.Error("Error Removing Brick ", err)
							// Do not return error, As we have tried to remove the bricks five times.
							// Might be IP address has been changed and handler is trying to remove
							// its own brick with previous IP address.
						}

						fmt.Printf("Brick deleted: %s\n", brick)
					}
				}
				b, _ := client.GetBricks(vol.Name)
				replica = len(b) - 1
			}
		}

		for _, p := range otherPeers {
			err = client.PeerDetach(p)
			if err != nil {
				rwolog.Error(err)
			}
			rwolog.Debug("Detach peer ", p, " error ", err)
		}

		// Uptil here, peers can be detached or not detached
		peers, err := client.GetPeers()
		if err != nil {
			rwolog.Error(err)
		}

		// Peer detach was not successful
		if len(peers) > 1 {
			return fmt.Errorf("Not able to detach all peers")
		}

	} else {
		rwolog.Debug("purgeOldStateOfGluster: Some peers are connected. Not removing bricks.")
	}

	return nil
}

// manageVolumes will ask leader to perform peer probe by serf query.
// Once the probing is successful. Gluster will add bricks to clusterfs volumes.
func manageVolumes(glusterClusterAddr string) error {

	rwolog.Debug("Performing peer probe with serf query to the leader ", glusterClusterAddr)

	glusterConnectedToPeerVal, err := helpers.SerfQuery("gluster", "peer probe "+glusterClusterAddr)
	if strings.Contains(string(glusterConnectedToPeerVal), "having volumes") {
		rwolog.Debug("Already part of some other gluster, peer probe failed. need to purge old state")
		// For extra clean up, so not checking error
		purgeOldStateOfGluster()
		// Lets wait
		time.Sleep(2 * time.Second)

	}

	if err != nil {
		rwolog.Error("Error while doing peer probe with serf query ", err.Error())
		return err
	}

	// For slow machines: Gluster will take some time to update peer list. sleep for 15 seconds to give gluster to update its information.
	time.Sleep(15 * time.Second)
	if strings.Contains(string(glusterConnectedToPeerVal), "success") {
		err = addBricks(glusterVolumeName, glusterClusterAddr)
		if err != nil {
			return err
		}
	} else {

		rwolog.Debug("Unable to probe peer from the leader. Will set up serf tags to retry.", glusterClusterAddr)
		serfTagGlusterRetry = serfTagGlusterRetry + 1

		err = helpers.SetGlusterRetryTag(strconv.Itoa(serfTagGlusterRetry))
		if err != nil {
			rwolog.Error("Error while setting up serf tags for retry")
			return err
		}
		time.Sleep(2 * time.Second)
		return fmt.Errorf("unable to probe peer from the leader, will set up serf tags to retry", glusterClusterAddr)
	}
	return nil
}

// createGlusterVolume creates gluster volume for the current member.
func createGlusterVolume(glusterClusterAddr string) error {

	// create volume
	brick := fmt.Sprintf("%s:%s/%s", glusterClusterAddr, glusterBrickPath, glusterVolumeName)
	vol := mg.GlusterVolume{
		Name:   glusterVolumeName,
		Bricks: []string{brick},
		Force:  1,
	}

	err := client.CreateGlusterVolume(vol)
	if err != nil {
		rwolog.Error("Error in creating gluster volume")
		return err
	}

	return nil
}

// setTag: Method sets the serf tags if mount point exists, else sets the gluster retry as the status.
func setTag() error {

	checkFormountPointval, err := exec.Command("sh", "-c", "mountpoint "+glusterMountPath+"/"+glusterVolumeName).Output()
	if err != nil {
		rwolog.Error("Unable to mount GlusterFS at "+glusterMountPath+"/"+glusterVolumeName, err.Error())
		serfTagGlusterRetry = serfTagGlusterRetry + 1
		err = helpers.SetGlusterRetryTag(strconv.Itoa(serfTagGlusterRetry))
		if err != nil {
			rwolog.Error("Error in setting serf gluster retry tag for gluster")
			return err
		}
	} else {
		rwolog.Debug("Mount point exists ", string(checkFormountPointval))
		peers, err := client.PeerStatus()
		if err != nil {
			rwolog.Error("Error Getting Peer Status")
		}

		var UUID string
		for _, peer := range peers {
			if peer.Name == "localhost" {
				UUID = peer.ID
			}
		}
		err = helpers.SetGlusterTag(UUID)
		if err != nil {
			rwolog.Error("Error in setting serf tag for gluster")
			return err
		}

		helpers.DeleteSerfTag("glusterretry")
	}
	return nil
}

// handleAppsVolumes: Adds bricks to the apps volume.
func handleAppsVolumes(glusterClusterAddr string) error {
	vols, err := client.ListVolumes()
	if err != nil {
		rwolog.Error("Error in getting gluster volume list", err.Error())
		return err
	}
	for _, vol := range vols {
		if vol.Name != glusterVolumeName {
			manageAppsVolume(vol.Name, glusterClusterAddr)
		}
	}
	return nil
}

// manageAppsVolume: Function starts Apps volume by adding bricks to it..
func manageAppsVolume(Vol string, glusterClusterAddr string) error {
	err := addBricks(Vol, glusterClusterAddr)
	if err != nil {
		return err
	}
	startGlusterVolume(Vol)
	return nil
}

// SetServerQuorumRatio will enable the quorum for the servers and sets the quorum ratio.
func SetServerQuorumRatio() error {

	// Get the list of volumes.
	vols, err := client.ListVolumes()
	if err != nil {
		rwolog.Error("Error in gluster volume list", err.Error())
		return err
	}

	// Enable the server quorum for each of the volume.
	for _, vol := range vols {
		err = helpers.EnableServerQuorum(client, vol.Name)
		if err != nil {
			rwolog.Error("Error in Enabling server quorum ", err.Error())
			return err
		}
	}

	// Set the quorum ratio for all the volumes.
	err = helpers.SetQuorumRatio(client, "update")
	if err != nil {
		rwolog.Error("Error in Enabling server quorum ", err.Error())
		return err
	}

	return nil
}

// addBricks: Add bricks to gluster volume.
func addBricks(volName string, glusterClusterAddr string) error {

	var replica int
	var retryCount int

	for {
		if retryCount < 30 {
			// Get Number of Bricks for the Volume
			bricks, err := client.GetBricks(volName)
			if err != nil {
				rwolog.Error("Error while checking gluster bricks count from gluster volume info ", err.Error())
				return err
			}
			// Check that our brick is already added in the gluster before procedding to add brick.

			brickAlreadyExists := false

			for _, brick := range bricks {
				if strings.Contains(brick.Name, glusterClusterAddr) {
					brickAlreadyExists = true
					rwolog.Debug("brick Name and glusterClusterAddr", brick.Name, glusterClusterAddr)
					break
				}
			}

			if brickAlreadyExists == true {
				rwolog.Debug("Brick already exists, No need to add bricks")
				break
			}

			// Set Replica to one more than Num of Bricks
			replica = len(bricks) + 1
			brickPath := fmt.Sprintf("%s:%s/%s", glusterClusterAddr, glusterBrickPath, volName)
			err = client.AddBrick(volName, brickPath, replica, 0)
			if err != nil {
				rwolog.Error("Error while adding bricks to volume, Sleeping for two seconds before retrying.")
				time.Sleep(2 * time.Second) //sleep for two seconds before retrying adding bricks.
				retryCount++
			} else {
				break
			}
		} else {
			return fmt.Errorf("Retry exceeded... Unable to add the bricks to the volume " + volName)
		}
	}
	rwolog.Debug("Successfully added Bricks to " + volName)
	return nil
}

// stopVolume: This helper method will stop gluster volume.
func stopVolume(vol string) {
	client.StopVolume(vol)
}

// deleteVolume: This helper method will delete gluster volume.
func deleteVolume(vol string) {
	client.RemoveVolume(vol)
}

// startGlusterVolume: This method will start gluster volume.
func startGlusterVolume(name string) {
	client.StartVolume(name)
}

// rmBricksFromVol removes bricks retrying 5 times
func rmBricksFromVol(n, b string, r int) error {

	retry := 5
	var err error
	for retry > 0 {
		err = client.RemoveBrick(n, b, r)
		if err == nil {
			break
		}

		rwolog.Error("Error in RemoveBrick: Trying again", err)
		if strings.Contains(err.Error(), "lock") {
			// To ensure that glusterd has released the lock for some request
			time.Sleep(1 * time.Second)
		}
		retry--
	}

	return err
}

// stopVolumeRetry will stop the volume.
func stopVolumeRetry(name string) error {

	retry := 5
	var err error
	for retry > 0 {

		err = client.StopVolume(name)
		if err == nil {
			break
		}
		retry--
		// wait for some time
		time.Sleep(3 * time.Second)
	}

	return err
}

// delVolume will remove volume from glusterFS.
func delVolume(name string) error {
	retry := 5
	var err error
	for retry > 0 {

		err = client.RemoveVolume(name)
		if err == nil {
			break
		}
		retry--
		// wait for some time
		time.Sleep(3 * time.Second)
	}

	return err
}

func deleteGlusterVolumes() error {

	vols, err := client.ListVolumes()
	if err != nil {
		rwolog.Error("Error while List gluster volumes", err.Error())
		return err
	}

	for _, vol := range vols {
		err = delVolume(vol.Name)
		if err != nil {
			rwolog.Error("Removing vol "+vol.Name+" Error: ", err)
			return err
		}

	}
	return nil

}
