package driver

import (
	"fmt"
	"github.com/docker/go-plugins-helpers/volume"
	"github.com/sapk/docker-volume-gluster/rest"
	"github.com/sapk/docker-volume-helpers/basic"
	"github.com/sapk/docker-volume-helpers/driver"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

type glusterAPIServer struct {
	Protocol, Address, Port string
}

func (s glusterAPIServer) URL() string {
	return s.Protocol + s.Address + ":" + s.Port
}

var (
	//MountTimeout timeout before killing a mount try in seconds
	MountTimeout = 30
	//CfgVersion current config version compat
	CfgVersion = 2
	//CfgFolder config folder
	CfgFolder = "/etc/docker-volumes/gluster/"

	gfsBase    = "/mnt/glusterfs/bricks/"
	username   = ""
	password   = ""
	TLSKey     = "/etc/ssl/glusterfs.key"
	TLSCert    = "/etc/ssl/glusterfs.pem"
	TLSCacert  = "/etc/ssl/glusterfs.ca"
	serverPORT = "5000"
	serverURL  = "localhost"
	restURL    = "https://localhost:5000"
	gServer    glusterAPIServer
)

var Peers []string

//GlusterDriver docker volume plugin driver extension of basic plugin
type GlusterDriver = basic.Driver

// LoadCredsForTLS reads the credentials and keys from env variables and loads them.
func LoadCredsForTLS(d *basic.Driver) error {
	if d == nil {
		return fmt.Errorf("driver shouldn't be nil")
	}

	// read the environment variables
	user := os.Getenv("user")
	if !isValidAuthCred(user) {
		return fmt.Errorf("no user provided")
	}

	password := os.Getenv("password")
	if !isValidAuthCred(password) {
		return fmt.Errorf("no password provided")
	}

	port := os.Getenv("port")
	if !isValidPort(port) {
		// default port to 5000
		port = "5000"
	}

	key := os.Getenv("key")
	if !isValidPEM(key) {
		return fmt.Errorf("key not a proper pem")
	}

	cert := os.Getenv("cert")
	if !isValidPEM(cert) {
		return fmt.Errorf("cert not a proper pem")
	}

	cacert := os.Getenv("cacert")
	if !isValidPEM(cacert) {
		return fmt.Errorf("cacert not a proper pem")
	}

	// Load the variables into CustomOptions
	d.Config.CustomOptions["user"] = user
	d.Config.CustomOptions["password"] = password
	d.Config.CustomOptions["key"] = key
	d.Config.CustomOptions["cert"] = cert
	d.Config.CustomOptions["ca"] = cacert
	d.Config.CustomOptions["port"] = port

	d.SaveConfig()

	return nil
}

//Init start all needed deps and serve response to API call
func Init(root string, mountUniqName bool) *GlusterDriver {
	logrus.Debugf("Init gluster driver at %s, UniqName: %v", root, mountUniqName)
	config := basic.DriverConfig{
		Version: CfgVersion,
		Root:    root,
		Folder:  CfgFolder,
		CustomOptions: map[string]interface{}{
			"mountUniqName": mountUniqName,
		},
	}
	eventHandler := basic.DriverEventHandler{
		OnMountVolume: mountVolume,
		GetMountName:  GetMountName,
	}

	return basic.Init(&config, &eventHandler)
}

func mountVolume(d *basic.Driver, v driver.Volume, m driver.Mount, r *volume.MountRequest) (*volume.MountResponse, error) {
	// load creds and keys needed to contact glusterfs API server
	loadGlusterAPIServerCreds(d)

	// Get docker volume name
	mountPath := m.GetPath()
	dockerVol := mountPath[strings.LastIndex(mountPath, "/")+1 : len(mountPath)]
	logrus.Debugf("docker volume: %s", dockerVol)

	cmd := fmt.Sprintf("glusterfs --secure-mgmt=true --volfile-id='%s' -s '%s' %s", dockerVol, strings.Join(Peers, "' -s '"), mountPath)
	//cmd := fmt.Sprintf("glusterfs %s %s", parseVolURI(v.GetOptions()["voluri"]), m.GetPath())

	logrus.Debugf("glusterfs client command: %s", cmd)
	//cmd := fmt.Sprintf("/usr/bin/mount -t glusterfs %s %s", v.VolumeURI, m.Path)
	//TODO fuseOpts   /usr/bin/mount -t glusterfs v.VolumeURI -o fuseOpts v.Mountpoint

	//volName, _ := getVolAndServerNames(v.GetOptions()["voluri"])
	if err := createGlusterVol(restURL, dockerVol, Peers); err != nil {
		return nil, err
	}

	if err := d.RunCmd(cmd); err != nil {
		return nil, err
	}
	return &volume.MountResponse{Mountpoint: m.GetPath()}, nil
}

func validatePath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}
	return nil
}

func createGlusterVol(restURL string, volName string, peers []string) error {
	logrus.Debugf("gluster REST API %s", restURL)

	client := rest.NewClient(gServer.URL(), gfsBase, username, password)

	// Enable TLS if server is https
	key := TLSKey
	if err := validatePath(key); err != nil {
		return err
	}

	cert := TLSCert
	if err := validatePath(cert); err != nil {
		return err
	}

	ca := TLSCacert
	if err := validatePath(ca); err != nil {
		return err
	}

	if len(key) > 0 && len(cert) > 0 && len(ca) > 0 {
		// enable TLS
		client.EnableTLS(key, cert, ca)
	}

	logrus.Debugf("Checking if gluster volume %s exists", volName)

	exist, err := client.VolumeExist(volName)
	if err != nil {
		logrus.Warn("Unable to check if gluster volume exists, %v", err)
		return err
	}

	if !exist {
		logrus.Debugf("Volume does not exist...")
		logrus.Debugf("Creating gluster volume %s ...", volName)
		if err := client.CreateVolume(volName, peers); err != nil {
			logrus.Errorf("Unable to create gluster volume, %s", err.Error())
			return err
		}
		logrus.Debugf("Gluster volume %s successfully created", volName)

	}

	return nil
}
