//Copyright 2019, Intel Corporation

package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"helpers"
	"io/ioutil"
	"os"
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

	swarm, _ := cli.SwarmInspect(context.Background())

	if input == "worker" {
		writeToken("worker", swarm.JoinTokens.Worker)
		fmt.Println("Success")
	}

	if input == "manager" {
		writeToken("manager", swarm.JoinTokens.Manager)
		fmt.Println("Success")
	}

}

func writeToken(role string, token string) error {

	 managerTokenPath,_ := os.LookupEnv("MANAGER")
	 workerTokenPath, _ := os.LookupEnv("WORKER")
	 glusterVolumeName, _ := os.LookupEnv("GLUSTER_VOLUME_NAME")
	 glusterMountPath,_ := os.LookupEnv("GLUSTER_MOUNT_PATH")
	 dockerTokenPath,_ := os.LookupEnv("TOKEN")
	 var path string

	if len(glusterVolumeName) == 0 || len(glusterMountPath) == 0 || len(dockerTokenPath) == 0 || len(managerTokenPath) == 0 || len(workerTokenPath) == 0 {
		fmt.Errorf("Env Variables not defined , ", "GLUSTER_VOLUME_NAME ", glusterVolumeName, " ",
			"GLUSTER_MOUNT_PATH ", glusterMountPath, " ",
			"TOKEN ", dockerTokenPath, " ",
			"MANAGER ", managerTokenPath, " ",
			"WORKER ", workerTokenPath)
	}

	err := helpers.CreateDirForToken()
	if err != nil {
		fmt.Errorf(err.Error())
		return nil
	}

	if role == "worker" {
		path = glusterMountPath + "/" + glusterVolumeName + "/" + dockerTokenPath + "/" + workerTokenPath + "/token.txt"
	} else {
		path = glusterMountPath + "/" + glusterVolumeName + "/" + dockerTokenPath + "/" + managerTokenPath + "/token.txt"
	}

	f, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return err
	}
	_, err = f.WriteString(token)
	if err != nil {
		fmt.Println(err)
		f.Close()
		return err
	}
	return nil
}
