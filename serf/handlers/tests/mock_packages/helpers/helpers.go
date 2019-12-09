package helpers

import (
	"fmt"
	"os"
	mg "rwogluster"
	"strconv"
)

//
var (
	SunnyDay = true
)

// GetIPAddr check if the serf iFace address and host address is assigned.
func GetIPAddr() (string, error) {
	if SunnyDay {
		return "192.168.1.1", nil
	}
	return "", fmt.Errorf("No Internet")

}

// CheckNetworkStatus function checks for the network availability for particular interval of time and exit the program if the Network is down for long time..
func CheckNetworkStatus() error {
	if SunnyDay {
		return nil
	}
	return fmt.Errorf("No Internet")
}

// ExecuteCommand helper method to execute shell commands.
func ExecuteCommand(cmdvalue string) bool {

	if len(cmdvalue) == 0 {
		fmt.Println("Pass the correct input")
		return false
	}

	if SunnyDay {
		return true
	}

	return false
}

// CreateDir helper function for creating directory
func CreateDir(path string) error {

	if SunnyDay {
		return nil
	}

	return fmt.Errorf("Directory Couldn't be created")
}

// StackDeploy deploy all containers as a part of stacks.
func StackDeploy() {

}

// Exists helper function to check existence of a directory
func Exists(path string) bool {
	if SunnyDay {
		return true
	}

	return false
}

// MountGlusterVolumes mounts the gluster volume.
func MountGlusterVolumes() error {
	if SunnyDay {
		return nil
	}
	return fmt.Errorf("Mount error")
}

func glusterfsCreds() (string, string, error) {
	if SunnyDay {
		return "admin", "password", nil
	}
	return "", "", fmt.Errorf("file not fund")

}

// GlusterLibClient returns a pointer to rwogluster client
func GlusterLibClient() *mg.Client {

	if SunnyDay {
		return &mg.Client{}
	}
	return nil
}

func CountAliveMembers() int {
	if SunnyDay {
		return 2
	}
	return 1
}

func CountLeaders() int {
	if SunnyDay {
		return 1
	}
	return 2
}

func CountTotalMembers() int {
	if SunnyDay {
		return 2
	}
	return 1
}

func CountFailedMembers() int {
	if SunnyDay {
		return 0
	}
	return 1

}

func CountAliveOrLeftMembers() int {
	if SunnyDay {
		return 1
	}
	return 2
}

func SetInitTag(value string) error {
	if SunnyDay {
		return nil
	}
	return fmt.Errorf("error setting tag")
}

func SetInProcessTag(value string) error {

	if SunnyDay {
		return nil
	}
	return fmt.Errorf("error setting tag")
}

func SetSwarmTag(value string) error {

	if SunnyDay {
		return nil
	}
	return fmt.Errorf("error setting tag")
}

func SetRoleTag(value string) error {

	if SunnyDay {
		return nil
	}
	return fmt.Errorf("error setting tag")

}

func SetWaitingForWorkerTag(value string) error {

	if SunnyDay {
		return nil
	}
	return fmt.Errorf("error setting tag")

}

func SetWaitingForLeaderTag(value string) error {

	if SunnyDay {
		return nil
	}
	return fmt.Errorf("error setting tag")
}

func SetGlusterTag(value string) error {

	if SunnyDay {
		return nil
	}
	return fmt.Errorf("error setting tag")
}

func SetGlusterRetryTag(value string) error {

	if SunnyDay {
		return nil
	}
	return fmt.Errorf("error setting tag")

}

func GetMemberIPByName(memberName string) (string, error) {

	if SunnyDay {
		return "192.168.0.1", nil
	}
	return "", fmt.Errorf("error getting IP")

}

func MemberIPByTagsAndName(tags map[string]string, Name string) (string, error) {

	if SunnyDay {
		return "192.168.1.2", nil
	}
	return "", fmt.Errorf("error setting tag")

}

func MemberNameByTagsAndName(tags map[string]string, Name string) (string, error) {

	if SunnyDay {
		return "ABC", nil
	}
	return "", fmt.Errorf("error checking tag")
}

func MemberIPByTagsAndStatus(tags map[string]string, status string) (string, error) {

	if SunnyDay {
		return "192.168.0.1", nil
	}
	return "", fmt.Errorf("error getting tag")
}

func MemberNameByTagsAndStatus(tags map[string]string, status string) (string, error) {

	if SunnyDay {
		return "ABC", nil
	}
	return "", fmt.Errorf("error checking tag")
}

func MemberForceLeave(MemberName string) error {

	if SunnyDay {
		return nil
	}
	return fmt.Errorf("error during execution")
}

func ListOfMembersByStatus(status string) ([]string, error) {

	if SunnyDay {
		return []string{"ABC", "XYZ"}, nil
	}
	return []string{""}, fmt.Errorf("error listing memebrs")
}

func ListMemberByTags(tags map[string]string) (string, error) {

	if SunnyDay {
		return "ABC", nil
	}
	return "", fmt.Errorf("error checking tag")
}

func DeleteSerfTag(tag string) error {

	if SunnyDay {
		return nil
	}
	return fmt.Errorf("error checking tag")
}

func SerfQuery(name string, payload string) (string, error) {

	if SunnyDay {
		return "ABC", nil
	}
	return "", fmt.Errorf("error in query")
}

func IsValidRole(role string) bool {

	if SunnyDay {
		return true
	}
	return false
}

func CheckForSymlink(filePath string) (bool, string, error) {
	if len(filePath) == 0 {
		return false, "", fmt.Errorf("no input provided")
	}
	if SunnyDay {
		return true, "/data/yes", nil
	}
	return false, "", nil
}

func SwarmInit(serfAdvertiseIface string, forceInit bool) error {
	if len(serfAdvertiseIface) == 0 {
		return fmt.Errorf("input not provided")
	}
	if SunnyDay {
		return nil
	}
	return fmt.Errorf("docker not up")

}

func SwarmJoin(serfAdvertiseIFace string, token string) error {

	if SunnyDay {
		return nil
	}
	return fmt.Errorf("docker not up")
}

func SwarmLeave(force bool) error {

	if SunnyDay {
		return nil
	}
	return fmt.Errorf("docker not up")
}

func CheckIfNodeExists(nodeName string) (string, error) {

	if SunnyDay {
		return "ABC", nil
	}
	return "", fmt.Errorf("docker not up")
}

func GetNodeIDByState(nodeState string) ([]string, error) {

	if SunnyDay {
		return []string{"ABC"}, nil
	}
	return []string{""}, fmt.Errorf("docker not up")
}

func GetNodeIDByStateAndHostname(nodeName string, nodeState string) (string, error) {
	if len(nodeName) == 0 || len(nodeState) == 0 {
		return "", fmt.Errorf("input paramters not provided")
	}
	if SunnyDay {
		return "ABC", nil
	}
	return "", fmt.Errorf("Docker still not up")
}

func GetSwarmNodeID() (string, error) {
	if SunnyDay {
		return "ABC", nil
	}
	return "", fmt.Errorf("Docker still not up")

}

func DockerInfoStatus() error {
	if SunnyDay {
		return nil
	}
	return fmt.Errorf("Docker still not up")
}

func GetNodeStatus(status string) (bool, error) {
	if len(status) == 0 {
		return false, fmt.Errorf("input not provided")
	}
	if SunnyDay {
		return true, nil
	}
	return false, nil

}

func CheckIfManager() (bool, error) {

	if SunnyDay {
		return true, nil
	}

	return false, nil
}

func DemoteNode(ID string) error {
	if SunnyDay {
		return nil
	}
	return fmt.Errorf("error demoting the node")
}

func PromoteNode(ID string) error {

	if SunnyDay {
		return nil
	}
	return fmt.Errorf("can't be promoted")
}

func RemoveNode(ID string, force bool) error {
	if SunnyDay {
		return nil
	}
	return fmt.Errorf("Docker still not up")

}

func UpdateNodeAvailabilityDrain(ID string) error {
	if SunnyDay {
		return nil
	}
	return fmt.Errorf("docker is not up")

}

func GetToken(input string) (string, error) {
	if SunnyDay {
		return "ABCD", nil
	}
	return "", fmt.Errorf("couldn't get token")
}

func GetSystemDockerNode(name string) (string, error) {

	if SunnyDay {
		return "A", nil
	}
	return "", fmt.Errorf("Docker still not up")
}

func GetAllSystemDockerNodes(name string) (string, error) {

	if SunnyDay {
		return "", nil
	}
	return "", fmt.Errorf("couldnt list containers")
}

func RemoveSystemDockerNode(ID string) error {

	if SunnyDay {
		return nil
	}
	return fmt.Errorf("coundn't remove container")
}

func RunArbiterContainer() error {

	if SunnyDay {
		return nil
	}
	return fmt.Errorf("arbiter couldn't be run")
}

func ExecuteProcessInNode(ID string, cmd []string) (string, error) {

	if SunnyDay {
		return "success", nil
	}
	return "", fmt.Errorf("error running command")
}

func WaitForDocker() bool {

	if SunnyDay {
		return false
	}
	return true
}

//Logger to maintain  level
type Logger struct {
	level int
}

var (
	rwolog = GetLogger()
)

//Error Logs
func (l *Logger) Error(input ...interface{}) (n int, err error) {

	if len(input) == 0 || input == nil {
		return 0, nil
	}

	// less than equal to maximum
	if l.level <= 2 {
		return fmt.Println(input...)
	}
	return 0, nil
}

//Info Logs
func (l *Logger) Info(input ...interface{}) (n int, err error) {

	if len(input) == 0 || input == nil {
		return 0, nil
	}

	// less than equal to maximum
	if l.level <= 2 {
		return fmt.Println(input...)
	}
	return 0, nil
}

//Debug Logs
func (l *Logger) Debug(input ...interface{}) (n int, err error) {

	if len(input) == 0 || input == nil {
		return 0, nil
	}

	if l.level == 2 {
		return fmt.Println(input...)
	}
	return 0, nil
}

//GetLogger will return log level
func GetLogger() *Logger {
	level, _ := os.LookupEnv("LOG_LEVEL")

	var l int
	var err error

	if len(level) == 0 {
		l = 1
	} else {
		l, err = strconv.Atoi(level)
		if err != nil {
			l = 1
		}
	}

	L := &Logger{
		level: l,
	}

	return L
}

func RestartGlusterContainer() {
	return
}

func GetNodeIDForWorkerWithMinTagValue() (string, error) {

	if SunnyDay {
		return "", nil
	}
	return "", fmt.Errorf("som error")
}
