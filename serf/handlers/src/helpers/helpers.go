//Copyright 2019, Intel Corporation

package helpers

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	mg "rwogluster"
	"strconv"
	"strings"
	"time"
)

var (
	credsFile = "creds.txt"
)

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

// GetIPAddr check if the serf iFace address and host address is assigned.
func GetIPAddr() (string, error) {

	var err error
	var IPAddr string

	for {
		conn, err := net.Dial("udp", "8.8.8.8:80")
		if err != nil {
			return "", err
		}
		defer conn.Close()

		localAddr := conn.LocalAddr().(*net.UDPAddr)

		IPAddr = localAddr.IP.String()
		if len(IPAddr) > 0 {
			break
		} else {
			rwolog.Info("IP Address not assigned yet..Looking for IP Address")
			time.Sleep(5 * time.Second)
		}
	}
	return IPAddr, err
}

// CheckNetworkStatus function checks for the network availability for particular interval of time and exit the program if the Network is down for long time..
func CheckNetworkStatus() error {

	var networkDownFlag bool
	var retryCount int
	networkIP, _ := os.LookupEnv("NETWORK_TEST_FQDN")

	if len(networkIP) == 0 {
		//setting up IP to google.com as NETWORK_TEST_FQDN is empty.
		networkIP = "www.google.com"
	}

	rwolog.Debug("Checking network Status ...")
	networkDownFlag = false
	for {

		_, err := net.LookupIP(networkIP)
		if err != nil {
			rwolog.Error("Network is not up ", err.Error())
			networkDownFlag = true
			time.Sleep(3 * time.Second) //sleep for 3 seconds to get network up.
			retryCount++
			if retryCount > 100 {
				return fmt.Errorf("Network is down for more then 3000 seconds. Exiting member failed ")
			}
		} else {
			rwolog.Debug("Network is up.")
			break
		}
	}

	if networkDownFlag == true {
		rwolog.Debug(" Network was offline, delayed start to allow other members to settle.")
		time.Sleep(30 * time.Second)
	}

	return nil
}

// ExecuteCommand helper method to execute shell commands.
func ExecuteCommand(cmdvalue string) bool {

	if len(cmdvalue) == 0 {
		rwolog.Error("Pass the correct input")
		return false
	}

	cmd, err := exec.Command("sh", "-c", cmdvalue).Output()
	if err != nil {
		rwolog.Error("Error in executing the command ", err.Error())
		return false
	}

	rwolog.Debug("OutPut Value :", cmd)

	return true
}

// CreateDir helper function for creating directory
func CreateDir(path string) error {

	_, err := exec.Command("sh", "-c", "mkdir -p "+path).Output()
	if err != nil {
		rwolog.Error("Error in creating directory ", err.Error())
		return err
	}

	return nil
}

// Exists helper function to check existence of a directory
func Exists(path string) bool {

	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true

}

// MountGlusterVolumes mounts the gluster volume.
func MountGlusterVolumes() error {

	glusterMountPath, _ := os.LookupEnv("GLUSTER_MOUNT_PATH")
	glusterVolumeName, _ := os.LookupEnv("GLUSTER_VOLUME_NAME")
	rwoBasePath, _ := os.LookupEnv("RWO_BASE_PATH")
	glusterContainerID, _ := GetSystemDockerNode("gluster-server")
	glusterClusterAddr, err := GetIPAddr()
	if err != nil || len(glusterClusterAddr) == 0 {
		return fmt.Errorf("Unable to retrieve IP  Address. Exiting")
	}

	if glusterContainerID == "" {
		return fmt.Errorf("Unable to retrieve docker container ID. Exiting")
	}
	if len(glusterMountPath) == 0 || len(glusterVolumeName) == 0 || len(rwoBasePath) == 0 {
		return fmt.Errorf("Env Variables are empty")
	}

	checkFormountPointval, err := exec.Command("sh", "-c", "mountpoint "+glusterMountPath+"/"+glusterVolumeName).Output()
	if err != nil {

		rwolog.Debug("Gluster mount point does not exists")
		glusterVolumestatusVal, err := exec.Command("sh", "-c", "system-docker exec "+glusterContainerID+" gluster volume status "+glusterVolumeName+" | grep -i Brick | grep -q "+glusterClusterAddr).Output()
		if err != nil && string(glusterVolumestatusVal) != "" {
			return fmt.Errorf("Error while checking Gluster Volume status ", err.Error())
		}

		var glusterVolMount = "system-docker exec --env GLUSTER_CLUSTER_ADDR=" + glusterClusterAddr + "  --env GLUSTER_MOUNT_PATH=" + glusterMountPath + " --env GLUSTER_VOLUME_NAME=" + glusterVolumeName + " " + glusterContainerID + " " + rwoBasePath + "/gluster/volume-entrypoint.sh"
		if ExecuteCommand(glusterVolMount) != true {
			rwolog.Debug("Error while mounting the gluster volume  ", glusterVolMount)
		}

		//wait for some time after mount is successful.
		time.Sleep(2 * time.Second)

	} else {
		rwolog.Debug("Mount point exists ", string(checkFormountPointval))
		return nil

	}

	return nil

}

func glusterfsCreds() (string, string, error) {
	// Check if the username/password file exists
	path, _ := os.LookupEnv("CREDS_DIR")
	credsFilePath := path + "/" + credsFile
	if _, err := os.Stat(credsFilePath); os.IsNotExist(err) {
		rwolog.Error("GlusterLibClient(): credential file does not exist")
		return "", "", err
	}

	data, err := ioutil.ReadFile(credsFilePath)
	if err != nil {
		rwolog.Error("GlusterLibClient(): Error Reading File ", err)
		return "", "", err
	}

	// Assuming that creds file has two lines
	// First line has username
	// Second line has a password
	usr := strings.Split(string(data), "\n")[0]
	pswd := strings.Split(string(data), "\n")[1]

	return usr, pswd, err
}

// GlusterLibClient returns a pointer to rwogluster client
func GlusterLibClient() *mg.Client {

	// To read the user/password from conf
	user, password, err := glusterfsCreds()
	if err != nil {
		panic("gluster creds not found")
	}

	key, _ := os.LookupEnv("TLS_KEY")
	cert, _ := os.LookupEnv("TLS_CERT")
	ca, _ := os.LookupEnv("TLS_CACERT")
	tls := true

	if len(key) == 0 || len(cert) == 0 || len(ca) == 0 {
		rwolog.Debug("GlusterLibClientWithTLS(): Need keys for TLS. Default to http")
		tls = false
	}

	var URL string
	if tls {
		URL = "https://localhost:5000"
	} else {
		URL = "http://localhost:5000"
	}

	cl := mg.NewClient(URL, user, password)
	if tls {
		cl.EnableTLS(key, cert, ca)
	}
	return cl
}

// CreateDirForToken will create dir for docker swarm join tokens once the mount operation is done.
func CreateDirForToken() error {

	var glusterVolumeName, _ = os.LookupEnv("GLUSTER_VOLUME_NAME")
	var glusterMountPath, _ = os.LookupEnv("GLUSTER_MOUNT_PATH")
	var dockerTokenPath, _ = os.LookupEnv("TOKEN")
	var managerTokenPath, _ = os.LookupEnv("MANAGER")
	var workerTokenPath, _ = os.LookupEnv("WORKER")

	if len(glusterVolumeName) == 0 || len(glusterMountPath) == 0 || len(dockerTokenPath) == 0 || len(managerTokenPath) == 0 || len(workerTokenPath) == 0 {
		rwolog.Error("Env Variables not defined , ", "GLUSTER_VOLUME_NAME ", glusterVolumeName, " ",
			"GLUSTER_MOUNT_PATH ", glusterMountPath, " ",
			"TOKEN ", dockerTokenPath, " ",
			"MANAGER ", managerTokenPath, " ",
			"WORKER ", workerTokenPath)
	}

	var mgrTokenPath = glusterMountPath + "/" + glusterVolumeName + "/" + dockerTokenPath + "/" + managerTokenPath
	var wrkrTokenPath = glusterMountPath + "/" + glusterVolumeName + "/" + dockerTokenPath + "/" + workerTokenPath

	dirPath := []string{mgrTokenPath, wrkrTokenPath}

	for i := 0; i < len(dirPath); i++ {
		if Exists(dirPath[i]) == false {
			fmt.Print("creating dir ", dirPath[i])
			err := CreateDir(dirPath[i])
			if err != nil {
				return err
			}
			err = os.Chmod(dirPath[i], 0600)
			if err != nil {
				return err

			}
		}
	}
	return nil
}
