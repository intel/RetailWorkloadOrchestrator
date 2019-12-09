//Copyright 2019, Intel Corporation

package main

import (
	"fmt"

	"context"
	"github.com/docker/docker/client"
	"os"

	"helpers"
	"io/ioutil"
	"strings"
)

func main() {

	var input string

	// Get the input from standard input.
	inputCmd, _ := ioutil.ReadAll(os.Stdin)
	input = strings.TrimSpace(string(inputCmd))
	if len(input) == 0 {
		return
	}

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return
	}

	cli.NegotiateAPIVersion(context.Background())

	ID, err := helpers.GetSwarmNodeID()
	if err != nil {
		return
	}

	node, _, err := cli.NodeInspectWithRaw(context.Background(), ID)
	if err != nil {
		return
	}

	node.Spec.Role = "manager"

	err = cli.NodeUpdate(context.Background(), node.ID, node.Version, node.Spec)
	if err != nil {
		return
	}

	fmt.Println("Swarm Promote Passed")

}
