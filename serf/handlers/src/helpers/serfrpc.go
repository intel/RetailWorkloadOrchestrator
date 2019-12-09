//Copyright 2019, Intel Corporation

package helpers

import (
	"time"

	"github.com/hashicorp/serf/client"
	"os"
	"sort"
	"strconv"
	"strings"
)

var (
	rpcURL = "127.0.0.1:7373" //RPC localhost URL.
)

// CountAliveMembers count total number of alive members.
func CountAliveMembers() int {

	aliveCount, err := memberCount("alive")

	if err != nil {
		aliveCount = 0
	}

	return aliveCount
}

// CountLeaders count number of leaders in the cluster.
func CountLeaders() int {

	tag := make(map[string]string)
	tag["role"] = "leader"

	memberCount, err := memberCountWithTag("alive", tag)

	if err != nil {
		memberCount = 0
	}

	return memberCount
}

// CountTotalMembers count Alive and Failed members in the cluster.
func CountTotalMembers() int {

	aliveCount, err := memberCount("alive")
	failedCount, err := memberCount("failed")

	if err != nil {
		return 0
	}

	return aliveCount + failedCount
}

// CountFailedMembers count  total number of failed members.
func CountFailedMembers() int {

	failedCount, err := memberCount("failed")

	if err != nil {
		failedCount = 0
	}

	return failedCount
}

// CountAliveOrLeftMembers count total number of alive/left members.
func CountAliveOrLeftMembers() int {

	leftCount, err := memberCount("left")
	if err != nil {
		leftCount = 0
	}

	aliveCount, err := memberCount("alive")
	if err != nil {
		aliveCount = 0
	}

	return leftCount + aliveCount
}

// SetInitTag set init Tag for the serf member.
func SetInitTag(value string) error {

	tag := make(map[string]string)
	tag["init"] = value
	return setSerfTag(tag)
}

// SetInProcessTag set inprocess Tag for the serf member.
func SetInProcessTag(value string) error {

	tag := make(map[string]string)
	tag["inprocess"] = value
	return setSerfTag(tag)
}

// SetSwarmTag set swarm Tag for the serf member.
func SetSwarmTag(value string) error {

	tag := make(map[string]string)
	tag["swarm"] = value
	return setSerfTag(tag)
}

// SetRoleTag set role Tag for the serf member.
func SetRoleTag(value string) error {

	tag := make(map[string]string)
	tag["role"] = value
	return setSerfTag(tag)
}

// SetWaitingForWorkerTag set waitingforworker Tag for the serf member.
func SetWaitingForWorkerTag(value string) error {

	tag := make(map[string]string)
	tag["waitingforworker"] = value
	return setSerfTag(tag)
}

// SetWaitingForLeaderTag set waitingforleader Tag for the serf member.
func SetWaitingForLeaderTag(value string) error {

	tag := make(map[string]string)
	tag["waitingforworker"] = value
	return setSerfTag(tag)

}

// SetGlusterTag set gluster Tag for the serf member.
func SetGlusterTag(value string) error {

	tag := make(map[string]string)
	tag["gluster"] = value
	return setSerfTag(tag)
}

// SetGlusterRetryTag set gluster retry Tag for the serf member.
func SetGlusterRetryTag(value string) error {

	tag := make(map[string]string)
	tag["glusterretry"] = value
	return setSerfTag(tag)
}

// GetMemberIPByName get the serf member IP address.
func GetMemberIPByName(memberName string) (string, error) {

	RPCClientCur, _ := client.NewRPCClient(rpcURL)
	defer RPCClientCur.Close()
	//Get the serf members.
	members, err := RPCClientCur.Members()
	if err != nil {
		rwolog.Error(err.Error())
		return "", err
	}

	for i := 0; i < len(members); i++ {
		if members[i].Name == memberName {
			return members[i].Addr.String(), nil
		}
	}

	return "", nil
}

// GetTagValue get the value of a particular Tag.
func GetTagValue(tags string) (string, error) {

	RPCClientCur, _ := client.NewRPCClient(rpcURL)
	defer RPCClientCur.Close()

	//Get the serf members.
	members, err := RPCClientCur.Members()
	if err != nil {
		rwolog.Error(err.Error())
		return "", err
	}

	for i := 0; i < len(members); i++ {
		hostname, _ := os.Hostname()
		if members[i].Name == hostname {
			tagValue := members[i].Tags[tags]
			return tagValue, nil

		}
	}

	return "", nil
}

// GetNodeIDForWorkerWithMinTagValue Checks and return worker Nodeid with minimum value of init.
// Init Value of leader/Reachable node are avoided.
// Sort all the worker's node init in ascending order.
// Node with the minimum value is compared with the initarray[0].
func GetNodeIDForWorkerWithMinTagValue() (string, error) {

	RPCClientCur, _ := client.NewRPCClient(rpcURL)
	defer RPCClientCur.Close()

	//Get the serf members
	members, err := RPCClientCur.Members()
	if err != nil {
		rwolog.Error(err.Error())
		return "", err
	}

	//Array for storing all inits of worker
	initArray := make([]int, len(members))
	var initCount int

	//For storing all reachable node Swarm ID, to distinguish between manager/worker.
	reachableNodeIDs, err := SerfQuery("docker", "reachableNodeIDs")
	if err != nil {
		rwolog.Error("Node Failed to check the ReachableNodeCount ", err)
		return "", err
	}

	rwolog.Error("reachablenodeinfo", reachableNodeIDs) //Debug Info

	for i := 0; i < len(members); i++ {

		//Avoid the value for leader node
		if members[i].Tags["role"] == "leader" {
			rwolog.Debug("Leader return\n")
			continue
		}

		//If reachable node avoid the init value for it.
		if strings.Contains(reachableNodeIDs, members[i].Tags["swarm"]) {
			rwolog.Debug("Avoiding init for reachable node")
			continue
		}

		if members[i].Status != "alive" {
			continue
		}

		initArray[initCount], _ = strconv.Atoi(members[i].Tags["init"])
		rwolog.Debug("init array :", initArray[initCount])
		initCount++

	}

	// Find and remove 0 values.
	for i := 0; i < len(initArray); i++ {
		if initArray[i] == 0 {
			initArray = append(initArray[:i], initArray[i+1:]...)
			i--
		}
	}

	sort.Ints(initArray) //Sort the array

	rwolog.Debug("sorted init:", initArray) //Debug Info

	//Return the swarmid to be promoted
	for i := 0; i < len(members); i++ {

		if len(initArray) == 0 {
			rwolog.Debug("No Value in init, returning back")
			continue
		}

		tmp, _ := strconv.Atoi(members[i].Tags["init"])
		rwolog.Debug("Member init  :", tmp)
		rwolog.Debug("Member swarm id:", members[i].Tags["swarm"])

		if tmp == initArray[0] {

			swarmID, err := GetSwarmNodeID()
			if err != nil {
				rwolog.Error("Failed to docker info: ", err)
				return "", err
			}

			rwolog.Debug("SwarmNodeID from docker API :", swarmID)

			if swarmID == members[i].Tags["swarm"] {
				rwolog.Debug("Matched") //Debug
				return members[i].Tags["swarm"], nil
			}
		}
	}

	return "", nil

}

// MemberNameByTagsAndName get the member name for the specified tag and member name.
func MemberNameByTagsAndName(tags map[string]string, Name string) (string, error) {

	RPCClientCur, _ := client.NewRPCClient(rpcURL)
	defer RPCClientCur.Close()

	//Get the serf members.
	members, err := RPCClientCur.Members()
	if err != nil {
		rwolog.Error(err.Error())
		return "", err
	}

	var matchFound bool
	matchFound = false
	for i := 0; i < len(members); i++ {

		if members[i].Name == Name {
			for k, v := range tags {
				tagValue := members[i].Tags[k]
				if tagValue == v {
					matchFound = true
					rwolog.Debug(tagValue)
				} else {
					matchFound = false
					break
				}
			}

			if matchFound == true {
				return members[i].Name, nil
			}
		}
	}
	return "", nil
}

// MemberIPByTagsAndStatus get the member IP for the specified tag and status.
func MemberIPByTagsAndStatus(tags map[string]string, status string) (string, error) {

	RPCClientCur, _ := client.NewRPCClient(rpcURL)
	defer RPCClientCur.Close()

	//Get the serf members.
	members, err := RPCClientCur.Members()
	if err != nil {
		rwolog.Error(err.Error())
		return "", err
	}

	var matchFound bool
	matchFound = false
	for i := 0; i < len(members); i++ {

		if members[i].Status == status {
			for k, v := range tags {
				tagValue := members[i].Tags[k]
				if tagValue == v {
					matchFound = true
					rwolog.Debug(tagValue)
				} else {
					matchFound = false
					break

				}
			}

			if matchFound == true {
				return members[i].Addr.String(), nil
			}
		}
	}

	return "", nil
}

// MemberNameByTagsAndStatus get the member name for the specified tag and status.
func MemberNameByTagsAndStatus(tags map[string]string, status string) (string, error) {

	RPCClientCur, _ := client.NewRPCClient(rpcURL)
	defer RPCClientCur.Close()
	//Get the serf members.
	members, err := RPCClientCur.Members()
	if err != nil {
		rwolog.Error(err.Error())
		return "", err
	}

	var matchFound bool
	matchFound = false
	for i := 0; i < len(members); i++ {

		if members[i].Status == status {
			for k, v := range tags {
				tagValue := members[i].Tags[k]
				if tagValue == v {
					matchFound = true
					rwolog.Debug(tagValue)
				} else {
					matchFound = false
					break

				}
			}

			if matchFound == true {
				return members[i].Name, nil
			}
		}
	}

	return "", nil
}

// GetMemberCount get the count of serf members with specified status.
func memberCount(status string) (int, error) {

	RPCClientCur, _ := client.NewRPCClient(rpcURL)
	defer RPCClientCur.Close()
	//Get the serf members.
	members, err := RPCClientCur.Members()
	if err != nil {
		rwolog.Error(err.Error())
		return 0, err
	}

	//get the Serf members list and print the information
	var memberCount int
	for i := 0; i < len(members); i++ {
		if members[i].Status == status {
			memberCount++
		}
	}
	return memberCount, nil
}

// memberCountWithTag get the count of serf members with specified status and filtered tags.
func memberCountWithTag(status string, tags map[string]string) (int, error) {

	RPCClientCur, _ := client.NewRPCClient(rpcURL)
	defer RPCClientCur.Close()
	//Get the serf members.
	members, err := RPCClientCur.Members()
	if err != nil {
		rwolog.Error(err.Error())
		return 0, err
	}

	//get the Serf members list and print the information
	var memberCount int
	for i := 0; i < len(members); i++ {
		if members[i].Status == status {
			var matched bool
			matched = false
			for k, v := range tags {
				tagValue := members[i].Tags[k]
				if tagValue == v {
					matched = true

				} else {
					matched = false
					break
				}
			}

			if matched == true {
				memberCount++
			}
		}
	}

	return memberCount, nil
}

// MemberForceLeave remove the specified member from the serf cluster.
func MemberForceLeave(MemberName string) error {

	RPCClientCur, _ := client.NewRPCClient(rpcURL)
	defer RPCClientCur.Close()

	err := RPCClientCur.ForceLeave(MemberName)
	if err != nil {
		rwolog.Error("Error while removing failed node", err.Error())
		return err
	}

	return nil
}

// ListOfMembersByStatus get the list of members with specified status.
func ListOfMembersByStatus(status string) ([]string, error) {

	RPCClientCur, _ := client.NewRPCClient(rpcURL)
	defer RPCClientCur.Close()
	//Get the serf members.
	members, err := RPCClientCur.Members()
	if err != nil {
		rwolog.Error(err.Error())
		return nil, err
	}

	var membersName []string

	for i := 0; i < len(members); i++ {
		if members[i].Status == status {
			membersName = append(membersName, members[i].Name)
		}
	}

	return membersName, nil
}

// ListMemberByTags get the list of members with specified Tag.
func ListMemberByTags(tags map[string]string) (string, error) {

	RPCClientCur, _ := client.NewRPCClient(rpcURL)
	defer RPCClientCur.Close()
	//Get the serf members.
	members, err := RPCClientCur.Members()
	if err != nil {
		rwolog.Error(err.Error())
		return "", err
	}

	for i := 0; i < len(members); i++ {
		var matchFound bool
		matchFound = false
		for k, v := range tags {
			tagValue := members[i].Tags[k]

			if tagValue == v {
				matchFound = true
			} else {
				matchFound = false
				break
			}
		}

		if matchFound == true {
			return members[i].Name, nil
		}

	}
	return "", nil
}

// setSerfTag helper function to set Serf tag.
func setSerfTag(tag map[string]string) error {

	RPCClientCur, err := client.NewRPCClient(rpcURL)
	if err != nil {
		rwolog.Error(err.Error())
		return err
	}
	defer RPCClientCur.Close()

	var deleteTag []string

	err = RPCClientCur.UpdateTags(tag, deleteTag)
	if err != nil {
		rwolog.Error("Error while setting up serf tags", err.Error())
		return err
	}
	return nil
}

// DeleteSerfTag helper function to delete Serf tag.
func DeleteSerfTag(tag string) error {

	RPCClientCur, err := client.NewRPCClient(rpcURL)
	if err != nil {
		rwolog.Error(err.Error())
	}
	defer RPCClientCur.Close()

	emptyTags := map[string]string{}

	tags := []string{tag}
	err = RPCClientCur.UpdateTags(emptyTags, tags)
	if err != nil {
		rwolog.Error("Error while setting up serf tags", err.Error())
	}
	return nil
}

// SerfQuery function to handle serf queries which retries if the response is empty.
func SerfQuery(name string, payload string) (string, error) {

	var retry int //Number of times serf query to be retried.
	var data string
	var err error
	var timeout int
	timeout = 5000000000 //nano Seconds
	for {
		//Retry for 10 times if leader does not return back output.
		if retry <= 10 {
			data, err = query(name, payload, timeout)
			if err != nil {
				return "", err
			}

			if len(data) > 0 {
				break
			} else {
				timeout = timeout + timeout
				retry++
			}
		} else {
			return "", nil
		}
	}
	return data, nil
}

// query helper function to make serf query via RPC.
func query(name string, payload string, timeout int) (string, error) {

	RPCClientCur, err := client.NewRPCClient(rpcURL)
	if err != nil {
		rwolog.Error(err.Error())
		return "", err
	}
	defer RPCClientCur.Close()

	var out = timeout
	tOut := int64(out)

	//Make a serf query .
	respCh := make(chan client.NodeResponse, 128)
	TagRole := make(map[string]string)
	TagRole["role"] = "leader"
	var noAck bool
	var nodes []string
	//var timeout time.Duration

	var relayFactor int
	ackCh := make(chan string, 128)

	params := client.QueryParam{
		FilterNodes: nodes,
		FilterTags:  TagRole,
		RequestAck:  !noAck,
		RelayFactor: uint8(relayFactor),
		Timeout:     time.Duration(tOut),
		Name:        name,
		Payload:     []byte(payload),
		AckCh:       ackCh,
		RespCh:      respCh,
	}

	RPCClientCur.Query(&params)

	select {

	case r := <-respCh:
		if r.From == "" {
			rwolog.Debug("Serf query response from leader is empty. This may signify that leader might be busy executing member-failed. Retrying ...... ")
			return "", nil
		}
		payload := r.Payload
		if n := len(payload); n > 0 && payload[n-1] == '\n' {
			payload = payload[:n-1]
		}

		return string(payload), nil

	}

}

// IsValidRole Check for valid role from the serf agent.
func IsValidRole(role string) bool {

	if role == "leader" || role == "worker" {
		return true
	}

	return false
}
