package driver

import (
	"fmt"
	"github.com/docker/go-plugins-helpers/volume"
	"github.com/sapk/docker-volume-gluster/rest"
	"github.com/sapk/docker-volume-helpers/basic"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/idna"
	"io/ioutil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

const (
	validHostnameRegex = `(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9\-])\.)*([A-Za-z0-9\-]|[A-Za-z0-9\-][A-Za-z0-9\-]*[A-Za-z0-9\-])`
	validVolURIRegex   = `((` + validHostnameRegex + `)(,` + validHostnameRegex + `)*):\/?([^\/]+)(/.+)?`
)

func isValidURI(volURI string) bool {
	volURI, err := idna.ToASCII(volURI)
	if err != nil {
		return false
	}
	re := regexp.MustCompile(validVolURIRegex)
	return re.MatchString(volURI)
}

func parseVolURI(volURI string) string {
	volURI, _ = idna.ToASCII(volURI)
	re := regexp.MustCompile(validVolURIRegex)
	res := re.FindAllStringSubmatch(volURI, -1)
	volServers := strings.Split(res[0][1], ",")
	volumeID := res[0][10]
	subDir := res[0][11]

	if subDir == "" {
		return fmt.Sprintf("--volfile-id='%s' -s '%s'", volumeID, strings.Join(volServers, "' -s '"))
	}
	return fmt.Sprintf("--volfile-id='%s' --subdir-mount='%s' -s '%s'", volumeID, subDir, strings.Join(volServers, "' -s '"))
}

//GetMountName get moint point base on request and driver config (mountUniqName)
func GetMountName(d *basic.Driver, r *volume.CreateRequest) (string, error) {

	loadGlusterAPIServerCreds(d)

	// Load peers
	GetPeers()
	logrus.Debugf("GetMountName: %s", url.PathEscape(r.Name))
	r.Options["voluri"] = fmt.Sprintf("%s:%s", strings.Join(Peers, ","), url.PathEscape(r.Name))
	return url.PathEscape(r.Name), nil
}

// GetPeers is a wrapper of rest GetPeers
func GetPeers() {
	client := rest.NewClient(gServer.URL(), gfsBase, username, password)
	logrus.Debug("Getting peer list from pool...")

	// Enable TLS if server is https
	key := TLSKey
	if err := validatePath(key); err != nil {
		return
	}

	cert := TLSCert
	if err := validatePath(cert); err != nil {
		return
	}

	ca := TLSCacert
	if err := validatePath(ca); err != nil {
		return
	}

	if len(key) > 0 && len(cert) > 0 && len(ca) > 0 {
		// enable TLS
		client.EnableTLS(key, cert, ca)
	}

	var err error

	Peers, err = client.GetPeers()
	if err != nil {
		logrus.Errorf("Unable to get peer list %s", err.Error())
	}
	logrus.Debug("Peers in pool %s", Peers)
}

func getVolAndServerNames(volURI string) (string, []string) {
	volURI, _ = idna.ToASCII(volURI)
	re := regexp.MustCompile(validVolURIRegex)
	res := re.FindAllStringSubmatch(volURI, -1)
	volServers := strings.Split(res[0][1], ",")
	volumeID := res[0][10]
	return volumeID, volServers
}

func isValidAuthCred(name string) bool {
	// validate user and password
	// should not be empty
	if len(name) == 0 {
		return false
	}

	return true
}

func isValidPEM(data string) bool {
	// validate PEM
	if len(data) == 0 {
		return false
	}
/* TODO: fix it wrt pem from env variables
	pemBlock, err := pem.Decode([]byte(data))
	if err != nil {
		return false
	}
	if pemBlock == nil {
		return false
	}
*/
	return true
}

func isValidPort(port string) bool {
	_, err := strconv.Atoi(port)
	if err != nil {
		return false
	}
	return true
}

func loadGlusterAPIServerCreds(d *basic.Driver) {
	// update username and password
	username = d.Config.CustomOptions["user"].(string)
	password = d.Config.CustomOptions["password"].(string)

	storeKeys(d)

	port := d.Config.CustomOptions["port"].(string)
	// update server
	gServer = glusterAPIServer{
		Protocol: "https://",
		Address:  "localhost",
		Port:     port,
	}

}

func storeKeys(d *basic.Driver) {

	ioutil.WriteFile(TLSKey, []byte(d.Config.CustomOptions["key"].(string)), 0400)
	ioutil.WriteFile(TLSCert, []byte(d.Config.CustomOptions["cert"].(string)), 0644)
	ioutil.WriteFile(TLSCacert, []byte(d.Config.CustomOptions["ca"].(string)), 0644)

}
