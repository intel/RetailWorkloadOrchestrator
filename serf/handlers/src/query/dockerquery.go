//Copyright 2019, Intel Corporation

package main

import (
	"fmt"
	"helpers"

	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func main() {

	var input string

	var dockerInfo []string
	var reachableNodeIDs []string

	var count int
	// Get the input from standard input.
	inputCmd, _ := ioutil.ReadAll(os.Stdin)
	input = string(inputCmd)
	if len(input) == 0 {
		return
	}

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return
	}

	cli.NegotiateAPIVersion(context.Background())

	swarmNodes, err := cli.NodeList(context.Background(), types.NodeListOptions{})
	if err != nil {
		return
	}

	count = 0

	//List all nodes - works only in Swarm Mode
	for _, swarmNode := range swarmNodes {
		dockerInfo = append(dockerInfo, swarmNode.Description.Hostname)
		dockerInfo = append(dockerInfo, swarmNode.ID)

		if strings.Contains(input, "reachable") &&
			swarmNode.ManagerStatus != nil &&
			swarmNode.ManagerStatus.Reachability == "reachable" &&
			swarmNode.ManagerStatus.Leader == false &&
			swarmNode.Spec.Availability == "active" {
			reachableNodeIDs = append(reachableNodeIDs, swarmNode.ID)
			count++
		}
	}

	if strings.Contains(input, "reachableNodeIDs") {
		fmt.Println(reachableNodeIDs)
		return
	}

	if strings.Contains(input, "reachable") {
		fmt.Println(strconv.Itoa(count))
		return
	}

	if strings.Contains(input, "cleanUpStaleMember") {
		helpers.RemoveDockerNodes()
		fmt.Println("Success")
		return
	}

	fmt.Println(dockerInfo)

}
