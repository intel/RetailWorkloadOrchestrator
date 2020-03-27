//Copyright 2019, Intel Corporation

package helpers

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// StackDeploy deploy all containers as a part of stacks.
func StackDeploy() {

	// /opt/rwo/Application_name/docker-compose.yml
	// /opt/rwo/ := Base path where RWO searches for Application compose files.
	// Application_name := Name of the application for which service needs to deployed.

	// Check the base dir for any application Directories.
	path := "/opt/stacks"
	apps, err := ioutil.ReadDir(path)
	if err != nil {
		rwolog.Error(err)
		return
	}

	for _, app := range apps {

		if app.IsDir() {

			appPath := filepath.Join(path, app.Name())
			files, _ := ioutil.ReadDir(appPath)

			for _, file := range files {
				filePath := filepath.Join(appPath, file.Name())

				// Check for symlink.
				isSymlink, symPath, err := CheckForSymlink(filePath)
				if err != nil {
					rwolog.Error("Error while checking Symlink, ", err.Error())
					continue
				}

				if isSymlink == true {
					rwolog.Debug("Encountered a Symlink, Assigning the Absolute path before deploying....")
					filePath = symPath
				}

				if strings.HasSuffix(filePath, ".yml") {
					// Create a buffer.
					var buf bytes.Buffer
					buf.WriteString("docker stack deploy -c ")
					buf.WriteString(filePath)
					buf.WriteString(" ")
					buf.WriteString(app.Name())
					command := buf.String()
					rwolog.Debug("command:", command)

					commandop, _ := exec.Command("sh", "-c", command).CombinedOutput()
					rwolog.Debug("Output for stack deploy: ", string(commandop))
				}

			}
		}
	}
}

// CheckForSymlink checks the file path for symlinks.
// if the file path is symlink, It will verify the
// symlink and provide absolute path.
// If the Symlink is broken, An appropriate error
// will be returned to calling function.
func CheckForSymlink(filePath string) (bool, string, error) {
	fileInfo, err := os.Lstat(filePath)
	if os.IsNotExist(err) || err != nil {
		return false, "", err
	}

	// Check path is symlink or not
	if fileInfo.Mode()&os.ModeSymlink != 0 {
		actualPath, err := filepath.EvalSymlinks(filePath)
		if err != nil {
			// Extract actual path from error msg
			apath := strings.Split(err.Error(), ":")[0]
			filePath = strings.Split(apath, " ")[1]
			return false, "", err
		}
		filePath = actualPath
		return true, filePath, nil
	}

	return false, "", nil
}

// SwarmInit initialise the swarm
// Pass false for force init, serfAdvertiseIface is IP Adress
func SwarmInit(serfAdvertiseIface string, forceInit bool) error {

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		rwolog.Error("Error in Initiating New Client:", err)
		return err
	}
	cli.NegotiateAPIVersion(context.Background())
	defer cli.Close()

	//docker init
	initRequest := swarm.InitRequest{
		ListenAddr:      "0.0.0.0:2377",
		AdvertiseAddr:   serfAdvertiseIface,
		ForceNewCluster: forceInit,
	}

	_, err = cli.SwarmInit(context.Background(), initRequest)
	if err != nil {
		rwolog.Error("Failed in creating Swarm:", err)
		return err
	}

	return nil

}

// SwarmJoin makes the node join the swarm
// serfAdvertiseIface is leader IP Adress
func SwarmJoin(serfAdvertiseIFace string, token string) error {

	var buf bytes.Buffer

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		rwolog.Error("Error in Initiating New Client:", err)
		return err
	}
	cli.NegotiateAPIVersion(context.Background())
	defer cli.Close()

	buf.WriteString(serfAdvertiseIFace)
	buf.WriteString(":2377") //port
	IP := buf.String()

	//docker join
	joinRequest := swarm.JoinRequest{
		ListenAddr: "0.0.0.0:2377",
		JoinToken:  token, RemoteAddrs: []string{IP},
	}

	err = cli.SwarmJoin(context.Background(), joinRequest)
	if err != nil {
		if strings.Contains(err.Error(), "Timeout was reached") {
			rwolog.Debug("Timeout before join..waiting for 30 seconds to complete in background ", err)
			return nil
		}

		rwolog.Error("Failed to Join Address", err)
		return err
	}
	return nil
}

// SwarmLeave is used to leave the swarm
// Pass true as parameter for force leave
func SwarmLeave(force bool) error {

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		rwolog.Error("Error in Initiating New Client:", err)
		return err
	}
	cli.NegotiateAPIVersion(context.Background())
	defer cli.Close()

	err = cli.SwarmLeave(context.Background(), force)

	if err != nil {
		if strings.Contains(err.Error(), "not part of a swarm") {
			rwolog.Debug("Not a part of swarm")
			return nil
		}

		rwolog.Error("Failed to Leave the node ", err)
		return err
	}

	return nil
}

// CheckIfNodeExists returns the list of IDs, if the node exists
func CheckIfNodeExists(nodeName string) (string, error) {

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		rwolog.Error(err)
		return "", err
	}
	cli.NegotiateAPIVersion(context.Background())
	defer cli.Close()

	swarmNodes, err := cli.NodeList(context.Background(), types.NodeListOptions{})
	if err != nil {
		rwolog.Error(err)
		return "nil", err
	}

	for _, swarmNode := range swarmNodes {

		if swarmNode.Description.Hostname == nodeName {
			return swarmNode.ID, nil
		}
	}

	return "", nil
}

// GetNumberofMembersinSwarm returns the number of nodes present in the docker cluster
func GetNumberofMembersinSwarm() (int, error) {

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		rwolog.Error(err)
		return 0, err
	}
	cli.NegotiateAPIVersion(context.Background())
	defer cli.Close()

	swarmNodes, err := cli.NodeList(context.Background(), types.NodeListOptions{})
	if err != nil {
		rwolog.Error(err)
		return 0, err
	}

	return len(swarmNodes), nil
}

// GetNodeIDByState checks node Status and return list of IDs else error
func GetNodeIDByState(nodeState string) ([]string, error) {

	var ID []string

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		rwolog.Error(err)
		return nil, err
	}
	cli.NegotiateAPIVersion(context.Background())
	defer cli.Close()

	swarmNodes, err := cli.NodeList(context.Background(), types.NodeListOptions{})
	if err != nil {
		rwolog.Error(err)
		return nil, err
	}

	for _, swarmNode := range swarmNodes {

		if swarmNode.Status.State == swarm.NodeState(nodeState) {

			ID = append(ID, swarmNode.ID)
		}
	}

	return ID, nil
}

// GetNodeIDByStateAndHostname checks node Status & hostname if found return list of IDs else error
func GetNodeIDByStateAndHostname(nodeName string, nodeState string) (string, error) {

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		rwolog.Error(err)
		return "", err
	}

	cli.NegotiateAPIVersion(context.Background())
	defer cli.Close()

	swarmNodes, err := cli.NodeList(context.Background(), types.NodeListOptions{})
	if err != nil {
		rwolog.Error(err)
		return "", err
	}

	for _, swarmNode := range swarmNodes {

		if swarmNode.Status.State == swarm.NodeState(nodeState) && swarmNode.Description.Hostname == nodeName {

			return swarmNode.ID, nil
		}
	}

	return "", nil
}

// GetNodeIDByDownState checks hostname with down & It returns node ID only if same hostname exists with ready state.
func GetNodeIDByDownState() (string, error) {

	var hostName string
	var nodeID string

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		rwolog.Error(err)
		return "", err
	}

	cli.NegotiateAPIVersion(context.Background())
	defer cli.Close()

	swarmNodes, err := cli.NodeList(context.Background(), types.NodeListOptions{})
	if err != nil {
		rwolog.Error(err)
		return "", err
	}

	for _, swarmNode := range swarmNodes {

		if swarmNode.Status.State == "down" {

			hostName = swarmNode.Description.Hostname
			nodeID = swarmNode.ID
			break
		}
	}

	if len(hostName) != 0 && len(nodeID) != 0 {
		rwolog.Debug(hostName)
		rwolog.Debug(nodeID)
		for _, swarmNode := range swarmNodes {

			if hostName == swarmNode.Description.Hostname && swarmNode.Status.State == "ready" {
				return nodeID, nil
			}
		}
	}

	return "", nil
}

// GetSwarmNodeID gets the swarm node id
func GetSwarmNodeID() (string, error) {

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		rwolog.Error(err)
		return "", err
	}

	cli.NegotiateAPIVersion(context.Background())
	defer cli.Close()

	info, err := cli.Info(context.Background())
	if err != nil {
		rwolog.Error("Failed to get Swarm Node ID")
		return "", err
	}

	rwolog.Debug("node ID ", info.Swarm.NodeID)

	return info.Swarm.NodeID, nil

}

// DockerInfoStatus checks if docker is up
func DockerInfoStatus() error {

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		rwolog.Error(err)
		return err
	}
	cli.NegotiateAPIVersion(context.Background())
	defer cli.Close()

	_, err = cli.Info(context.Background())
	if err != nil {
		rwolog.Error("Failed to get Swarm Node ID")
		return err
	}

	return nil

}

// GetNodeStatus checks if Node is leader,reachable or worker
func GetNodeStatus(status string) (bool, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		rwolog.Error("Failed to create client")
		return false, err
	}

	cli.NegotiateAPIVersion(context.Background())
	defer cli.Close()

	ID, err := GetSwarmNodeID()
	if err != nil {
		rwolog.Error("Failed to get Node ID")
		return false, err
	}

	if len(ID) == 0 {
		rwolog.Debug("Not a part of swarm, so Node ID is empty")
		return false, nil
	}

	node, _, err := cli.NodeInspectWithRaw(context.Background(), ID)
	if err != nil {
		rwolog.Error("NodeInspect with Raw:", err)
		return false, nil
	}

	if node.ManagerStatus != nil {

		if status == "reachable" {
			if node.ManagerStatus.Reachability == "reachable" {
				return true, nil
			}

		}

		if status == "leader" {
			if node.ManagerStatus.Leader == true {
				return true, nil
			}

		}

	}

	return false, nil

}

// CheckIfManager checks if node is manager
func CheckIfManager() (bool, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		rwolog.Error(err)
		return false, err
	}

	cli.NegotiateAPIVersion(context.Background())
	defer cli.Close()

	ID, err := GetSwarmNodeID()
	if err != nil {
		rwolog.Error("Failed to get Node ID")
		return false, err
	}

	if len(ID) == 0 {
		rwolog.Debug("Not a part of swarm, so Node ID is empty")
		return false, nil
	}

	node, _, err := cli.NodeInspectWithRaw(context.Background(), ID)
	if node.ManagerStatus != nil {
		if node.ManagerStatus.Reachability == "reachable" || node.ManagerStatus.Leader == true {
			return true, nil
		}
	}
	return false, nil
}

// DemoteNode demotes the node.
func DemoteNode(ID string) error {

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		rwolog.Error(err)
		return err
	}
	cli.NegotiateAPIVersion(context.Background())
	defer cli.Close()

	node, _, err := cli.NodeInspectWithRaw(context.Background(), ID)

	node.Spec.Role = "worker"

	err = cli.NodeUpdate(context.Background(), node.ID, node.Version, node.Spec)
	if err != nil {
		rwolog.Error("Failed to update the node")
		return err

	}
	return nil

}

// PromoteNode promotes the node
func PromoteNode(ID string) error {

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		rwolog.Error(err)
		return err
	}

	cli.NegotiateAPIVersion(context.Background())
	defer cli.Close()

	node, _, err := cli.NodeInspectWithRaw(context.Background(), ID)

	node.Spec.Role = "manager"

	err = cli.NodeUpdate(context.Background(), node.ID, node.Version, node.Spec)
	if err != nil {
		rwolog.Error("Failed to update the node")
		return err

	}
	return nil
}

// RemoveNode removes the node.
func RemoveNode(ID string, force bool) error {

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		rwolog.Error(err)
		return err
	}
	cli.NegotiateAPIVersion(context.Background())
	defer cli.Close()

	err = cli.NodeRemove(context.Background(), ID, types.NodeRemoveOptions{Force: force})

	if err != nil {
		rwolog.Error("Failed to remove the node")
		return err

	}
	return nil
}

// UpdateNodeAvailabilityDrain Update status to drain
func UpdateNodeAvailabilityDrain(ID string) error {

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		rwolog.Error(err)
		return err
	}
	cli.NegotiateAPIVersion(context.Background())
	defer cli.Close()

	node, _, err := cli.NodeInspectWithRaw(context.Background(), ID)

	node.Spec.Availability = "drain"

	err = cli.NodeUpdate(context.Background(), node.ID, node.Version, node.Spec)
	if err != nil {
		rwolog.Error("Failed to update the node")
		return err

	}
	return nil
}

// GetToken return the token
// parameter passed:worker/manager
func GetToken(input string) (string, error) {

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", err
	}
	cli.NegotiateAPIVersion(context.Background())
	defer cli.Close()

	swarm, _ := cli.SwarmInspect(context.Background())

	if input == "worker" {
		return (swarm.JoinTokens.Worker), nil
	}

	if input == "manager" {
		return (swarm.JoinTokens.Manager), nil

	}

	rwolog.Debug("Pass the right input")

	return "", err
}

// GetSystemDockerNode return the node ID for system-docker host.
// Hostname is passed as parameter
func GetSystemDockerNode(name string) (string, error) {

	sock := getSystemDockerSock()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithHost(sock))
	if err != nil {
		rwolog.Error("failed to get client")
		return "", err
	}

	cli.NegotiateAPIVersion(context.Background())
	defer cli.Close()

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		rwolog.Error("failed to get client list")
		return "", err
	}

	for _, container := range containers {
		if strings.Contains(container.Names[0], name) {
			return container.ID, nil
		}
	}

	rwolog.Debug("There is no container name ", name)
	return "", nil

}

// RestartSystemDockerContainer will restart docker container.
func RestartSystemDockerContainer(name string) error {
	sock := getSystemDockerSock()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithHost(sock))
	if err != nil {
		rwolog.Error("failed to get client", err.Error())
		return err
	}

	cli.NegotiateAPIVersion(context.Background())
	defer cli.Close()

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		rwolog.Error("failed to get client list", err.Error())
		return err
	}

	var time time.Duration
	time = 5000000000 //nano seconds

	for _, container := range containers {
		if strings.Contains(container.Names[0], name) {
			err = cli.ContainerRestart(context.Background(), container.ID, &time)
			if err != nil {
				return err
			}
			return nil
		}
	}

	rwolog.Debug("There is no container name ", name)
	return nil

}

// GetAllSystemDockerNodes return the node ID for system-docker host by scanning all the non running containers.
func GetAllSystemDockerNodes(name string) (string, error) {

	sock := getSystemDockerSock()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithHost(sock))
	if err != nil {
		rwolog.Error("failed to get client")
		return "", err
	}

	cli.NegotiateAPIVersion(context.Background())
	defer cli.Close()

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		rwolog.Error("failed to get client list")
		return "", err
	}

	for _, container := range containers {
		if strings.Contains(container.Names[0], name) {
			return container.ID, nil
		}
	}

	rwolog.Debug("There is no container name ", name)
	return "", nil

}

// RemoveSystemDockerNode removes the node for system-docker host.
func RemoveSystemDockerNode(ID string) error {
	sock := getSystemDockerSock()
	cli, err := client.NewClientWithOpts(client.FromEnv,
		client.WithHost(sock))

	if err != nil {
		rwolog.Error("Failed to create new client")
		return err
	}

	cli.NegotiateAPIVersion(context.Background())
	defer cli.Close()

	err = cli.ContainerStop(context.Background(), ID, nil)
	if err != nil {
		rwolog.Error("Container Stop Error:", err)
		return err
	}

	err = cli.ContainerRemove(context.Background(), ID, types.ContainerRemoveOptions{Force: true})
	if err != nil {
		rwolog.Error("Container Remove Error:", err)
		return err
	}

	return nil
}

// RunArbiterContainer runs docker daemon as a container
func RunArbiterContainer() error {

	sock := getSystemDockerSock()
	var cmd []string
	cmd = append(cmd, "sh")
	cmd = append(cmd, "-c")
	cmd = append(cmd, "/usr/local/bin/dockerd --iptables=false")

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv,
		client.WithAPIVersionNegotiation(),
		client.WithHost(sock))

	if err != nil {
		rwolog.Error(err)
		return err
	}
	defer cli.Close()

	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		rwolog.Error(err)
	}

	var containerCreated bool
	containerCreated = false

	for _, image := range images {

		if containerCreated {
			break
		}

		rwolog.Debug(image.RepoTags)
		tags := image.RepoTags
		var image string

		for i := 0; i < len(tags); i++ {
			if strings.Contains(tags[i], "dind") {
				image = tags[i]
				resp, err := cli.ContainerCreate(ctx, &container.Config{User: "root", Cmd: cmd,
					NetworkDisabled: false, Hostname: "arbiter", Image: image,
				}, &container.HostConfig{Privileged: true}, nil, "rwo_arbiter")
				if err != nil {
					rwolog.Error(err)
					return err
				}

				if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
					rwolog.Error(err)
					return err
				}
				containerCreated = true
				break
			}
		}
	}
	return nil
}

func getSystemDockerSock() string {
	sock, _ := os.LookupEnv("SYSTEM_DOCKER_SOCK")

	if len(sock) == 0 {
		rwolog.Error("SYSTEM_DOCKER_SOCK is not defined in ENV. Setting up with default path.")
		sock = "/opt/system-docker.sock"
	}
	return "unix://" + sock

}

// ExecuteProcessInNode executes the passed string inside a container for system-docker
func ExecuteProcessInNode(ID string, cmd []string) (string, error) {

	sock := getSystemDockerSock()
	cli, err := client.NewClientWithOpts(client.FromEnv,
		client.WithAPIVersionNegotiation(),
		client.WithHost(sock))

	if err != nil {
		rwolog.Error("Failed to create the client")
		return "", err
	}
	defer cli.Close()

	// prepare exec
	execConfig := types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
	}

	containerExec, err := cli.ContainerExecCreate(context.Background(), ID, execConfig)
	if err != nil {
		rwolog.Error("Failed to Exec create", err)
		return "", err
	}

	executionID := containerExec.ID

	attachOp, err := cli.ContainerExecAttach(context.Background(), executionID, types.ExecStartCheck{})
	if err != nil {
		rwolog.Error("Failed to Exec Attach")
		return "", err
	}
	defer attachOp.Close()

	// read the output
	var opBuf, errorBuf bytes.Buffer
	ret := make(chan error)

	go func() {
		// StdCopy demultiplexes the stream into two buffers
		_, err = stdcopy.StdCopy(&opBuf, &errorBuf, attachOp.Reader)
		ret <- err
	}()

	select {
	case err := <-ret:
		if err != nil {
			return "", err
		}
		break

	case <-context.Background().Done():
		return "", err
	}

	// get the exit code
	_, err = cli.ContainerExecInspect(context.Background(), executionID)
	if err != nil {
		return "", err
	}

	return string(opBuf.Bytes()), nil

}

// WaitForDocker Checks the docker system status. This will wait until dockerD is fully up, in case dockerD is not running,
func WaitForDocker() bool {

	var retry int
	retry = 1
	for {
		if retry < 200 {
			err := DockerInfoStatus()
			if err != nil {
				rwolog.Error("Error while connecting to Docker  ", err.Error())
				retry++
				time.Sleep(1 * time.Second)
			} else {
				return true
			}
		} else {
			return false
		}
	}
}

// RestartGlusterContainer restart gluster container.
func RestartGlusterContainer() error {
	rwolog.Debug("Restart the gluster container")
	err := RestartSystemDockerContainer("gluster-server")
	if err != nil {
		rwolog.Error("Error while restarting gluster container ID")
	}
	return nil
}

// RemoveDockerNodes Get the docker nodes which are down and remove then from swarm.
func RemoveDockerNodes() error {

	//remove all nodes which are down
	node, err := GetNodeIDByState("down")
	if err != nil {
		rwolog.Debug("Error while retrieving the docker nodes which are down ", err.Error())
		return err
	}

	for _, element := range node {
		if len(element) > 0 {
			rwolog.Debug("Removing member from docker swarm cluster " + element)

			rwolog.Debug("DemoteNode " + element)
			err = DemoteNode(element)
			if err != nil {
				rwolog.Debug("Error in demoting a docker member. ", err.Error())
				//Do not return control back, If a leader goes down from the cluster, Docker will throw the error " Can't find manager in raft member list"
			}

			rwolog.Debug("RemoveNode " + element)
			err := RemoveNode(element, true)
			if err != nil {
				rwolog.Debug("Error while removing docker node ", err.Error())
				return err
			}
		}
	}
	return nil
}
