//Copyright 2019, Intel Corporation

package rwogluster

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	connectStatus = "CONNECTED"
	httpTimeout   = 10
)

// NewClient initializes a new client.
func NewClient(addr, user, password string) *Client {
	return &Client{
		addr:     addr,
		user:     user,
		password: password,
	}
}

func fileNotFound(path string) bool {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return false
	}
	return true
}

// EnableTLS enables TLS for client. It takes paths for key, cert and cacert as input.
func (r *Client) EnableTLS(key, cert, cacert string) error {

	// validate input params
	if len(key) == 0 || len(cert) == 0 || len(cacert) == 0 {
		return fmt.Errorf("key/cert paths not valid")
	}

	if fileNotFound(key) || fileNotFound(cert) || fileNotFound(cacert) {
		return fmt.Errorf("key/cert file not found")
	}

	// update the fields in the object
	r.key = key
	r.cert = cert
	r.cacert = cacert

	r.TLS = true
	return nil
}

// CreateGlusterVolume creates a gluster volume.
func (r Client) CreateGlusterVolume(Vol GlusterVolume) error {

	api := fmt.Sprintf("/api/1.0/volume/%s", Vol.Name)
	u := fmt.Sprintf("%s%s", r.addr, api)
	logrus.Debugf("POST %s", u)

	var bricks []string
	bricks = Vol.Bricks

	if len(bricks) == 0 {
		return fmt.Errorf("Bricks can't be empty")
	}

	params := url.Values{
		"bricks": {strings.Join(bricks, ",")},
	}

	if len(bricks) > 1 {
		params["replica"] = []string{strconv.Itoa(Vol.Replica)}
	}

	if len(Vol.Transport) == 0 {
		params["transport"] = []string{"tcp"}
	} else {
		params["transport"] = []string{Vol.Transport}
	}

	if Vol.Start > 0 {
		params["start"] = []string{"true"}
	}

	if Vol.Force > 0 {
		params["force"] = []string{"true"}
	}

	logrus.Debugf("Params %s", params)

	resp, err := r.sendRequest("POST", u, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return responseCheck(resp)
}

// VolumeExist returns whether a volume exist in the cluster with a given name or not.
func (r Client) VolumeExist(name string) (bool, error) {
	vols, err := r.volumes()
	if err != nil {
		return false, err
	}

	for _, v := range vols {
		if v.Name == name {
			return true, nil
		}
	}

	return false, nil
}

func (r Client) volumes() ([]volume, error) {
	api := fmt.Sprintf("/api/1.0/volumes")
	u := fmt.Sprintf("%s%s", r.addr, api)

	res, err := r.sendRequest("GET", u, nil)

	if err != nil {
		logrus.Errorf("Error GET volumes, %s", err.Error())
		return nil, err
	}

	var d volumeResponse
	if err := json.NewDecoder(res.Body).Decode(&d); err != nil {
		logrus.Errorf("Error decoding response from volumes, %s", err.Error())
		return nil, err
	}

	if !d.Ok {
		return nil, fmt.Errorf(d.Err)
	}
	return d.Data, nil
}

// StopVolume stops the volume with the given name in the cluster.
func (r Client) StopVolume(name string) error {

	api := fmt.Sprintf("/api/1.0/volume/%s/stop", name)
	u := fmt.Sprintf("%s%s", r.addr, api)

	resp, err := r.sendRequest("PUT", u, nil)
	if err != nil {
		return err
	}

	return responseCheck(resp)
}

// GetBricks list bricks in the gluster volume
func (r Client) GetBricks(vol string) ([]Brickinfo, error) {

	vols, err := r.volumes()
	if err != nil {
		return nil, err
	}
	for _, v := range vols {

		if v.Name == vol {
			return v.Bricks, err
		}
	}

	return nil, fmt.Errorf("No Bricks")
}

// GetPeers makes a RESTful GET call to all peers in the gluster pool
func (r Client) GetPeers() ([]string, error) {

	api := fmt.Sprintf("/api/1.0/peers")
	u := fmt.Sprintf("%s%s", r.addr, api)

	logrus.Debugf("GET %s", u)

	res, err := r.sendRequest("GET", u, nil)
	if err != nil {
		logrus.Errorf("Error GET peers, %s", err.Error())
		return nil, err
	}
	defer res.Body.Close()

	var response peerResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		logrus.Errorf("Error decoding response from peers, %s", err.Error())
		return nil, err
	}

	if !response.Ok {
		return nil, fmt.Errorf(response.Err)
	}

	// Loop through peers and ignore disconnected peers
	// Replace localhost peer to its hostname

	var peers []string

	for _, peer := range response.Data {

		if peer.Status == connectStatus {

			if peer.Name == "localhost" {
				ip, err := resolveLocalIP()
				if err != nil {
					return nil, err
				}
				peer.Name = ip
			}

			peers = append(peers, peer.Name)
		}

	}

	return peers, nil

}

func responseCheck(resp *http.Response) error {
	var p response
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return err
	}

	if !p.Ok {
		return fmt.Errorf(p.Err)
	}

	return nil
}

// StartVolume starts a volume
func (r Client) StartVolume(name string) error {

	api := fmt.Sprintf("/api/1.0/volume/%s/start", name)
	u := fmt.Sprintf("%s%s", r.addr, api)

	resp, err := r.sendRequest("PUT", u, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return responseCheck(resp)
}

// ListVolumes lists volumes
func (r Client) ListVolumes() ([]GlusterVolume, error) {
	vols, err := r.volumes()
	if err != nil {
		return nil, err
	}

	var Gvols []GlusterVolume
	for _, v := range vols {
		gv := GlusterVolume{
			Name:    v.Name,
			Replica: v.Replica,
		}
		for _, b := range v.Bricks {
			gv.Bricks = append(gv.Bricks, b.Name)
		}
		Gvols = append(Gvols, gv)
	}

	return Gvols, err
}

// PeerDetach detaches peer
func (r Client) PeerDetach(hostname string) error {
	// url: peer/:hostname
	// DELETE
	var err error
	api := "/api/1.0/peer/" + hostname
	u := fmt.Sprintf("%s%s", r.addr, api)

	params := url.Values{}
	params["force"] = []string{"true"}

	resp, err := r.sendRequest("DELETE", u, params)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return responseCheck(resp)
}

// PeerStatus gives all peers with status connected or disconnected
func (r Client) PeerStatus() ([]Peer, error) {
	// GET
	// api: /api/1.0/peers

	api := "/api/1.0/peers"
	u := fmt.Sprintf("%s%s", r.addr, api)

	res, err := r.sendRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var response peerResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		logrus.Errorf("Error decoding response from peers, %s", err.Error())
		return nil, err
	}

	if !response.Ok {
		return nil, fmt.Errorf(response.Err)
	}

	var peers []Peer
	for _, peer := range response.Data {
		peers = append(peers, peer)
	}

	return peers, nil
}

// PeerProbeWithMsg implements gluster peer probe and returns the message
func (r Client) PeerProbeWithMsg(hostname string) (string, error) {
	// POST
	// api: /api/1.0/peer/glusternode1
	api := "/api/1.0/peer/" + hostname
	u := fmt.Sprintf("%s%s", r.addr, api)

	resp, err := r.sendRequest("POST", u, nil)
	if err != nil {
		return err.Error(), err
	}

	defer resp.Body.Close()

	var rp response
	if err := json.NewDecoder(resp.Body).Decode(&rp); err != nil {
		logrus.Errorf("Error decoding response from peers, %s", err.Error())
		return err.Error(), err
	}

	if !rp.Ok {
		return rp.Err, fmt.Errorf(rp.Err)
	}

	return "success", err
}

// PeerProbe implements gluster peer probe
func (r Client) PeerProbe(hostname string) error {
	// POST
	// api: /api/1.0/peer/glusternode1
	api := "/api/1.0/peer/" + hostname
	u := fmt.Sprintf("%s%s", r.addr, api)

	resp, err := r.sendRequest("POST", u, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return responseCheck(resp)
}

// CleanupVolume does proper cleanup of the volume. It will stop the volume. remove bricks and detach from all peers
func (r Client) CleanupVolume(name string) error {

	// Stop Volume
	err := r.StopVolume(name)
	if err != nil {
		return err
	}

	// Check replica
	vols, err := r.volumes()
	if err != nil {
		return err
	}

	var volobj volume

	// Get the volume object corresponding to name
	for _, v := range vols {
		if v.Name == name {
			volobj = v
		}
	}

	// remove bricks
	// No error check needed.
	// if there is a single brick left remobe brick fails
	for _, b := range volobj.Bricks {
		r.RemoveBrick(name, b.Name, volobj.Replica)
	}

	// Delete Volume

	return r.RemoveVolume(name)
}

// RemoveVolume removes gluster Volume
func (r Client) RemoveVolume(name string) error {
	api := fmt.Sprintf("/api/1.0/volume/%s", name)
	u := fmt.Sprintf("%s%s", r.addr, api)

	resp, err := r.sendRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return responseCheck(resp)
}

// AddBrick adds bricks to an existing volume
func (r Client) AddBrick(vol string, b string, rep int, s int) error {
	/*
	   endpoints are not exposed in glusterrestd
	   Custom Endpoint:
	   	POST
	   	api: /api/1.0/volume/:name/addbrick
	   	params: { brick,replica,stripe,force}
	*/

	api := fmt.Sprintf("/api/1.0/volume/%s/addbrick", vol)
	u := fmt.Sprintf("%s%s", r.addr, api)
	logrus.Debugf("POST %s", u)

	params := url.Values{
		"brick": {b},
		"force": {"true"},
	}

	if rep > 0 {
		params["replica"] = []string{strconv.Itoa(rep)}
	}

	if s > 0 {
		params["stripe"] = []string{strconv.Itoa(s)}
	}

	logrus.Debugf("Params %s", params)

	resp, err := r.sendRequest("POST", u, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return responseCheck(resp)
}

// RemoveBrick removes bricks from an existing volume
func (r Client) RemoveBrick(vol, brick string, rep int) error {
	/*
	   endpoints are not exposed in glusterrestd
	   Custom Endpoint:
	   	POST
	   	api: /api/1.0/volume/:name/removebrick
	   	params: { brick,replica}
	*/

	api := fmt.Sprintf("/api/1.0/volume/%s/removebrick", vol)
	u := fmt.Sprintf("%s%s", r.addr, api)
	logrus.Debugf("POST %s", u)

	params := url.Values{
		"brick": {brick},
	}

	if rep > 0 {
		params["replica"] = []string{strconv.Itoa(rep)}
	}

	logrus.Debugf("Params %s", params)

	resp, err := r.sendRequest("POST", u, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return responseCheck(resp)
}

// EnableServerQuorum specifies server quorum type for a particular volume
func (r Client) EnableServerQuorum(vol string) error {

	api := fmt.Sprintf("/api/1.0/volume/quorum/type")
	u := fmt.Sprintf("%s%s", r.addr, api)
	logrus.Debugf("POST %s", u)
	quorumType := fmt.Sprintf("server")

	params := url.Values{
		"vol":  {vol},
		"type": {quorumType},
	}

	logrus.Debugf("Params %s", params)

	resp, err := r.sendRequest("POST", u, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return responseCheck(resp)
}

// SetServerQuorumratio specifies server quorum ratio for all the volumes.
func (r Client) SetServerQuorumratio(ratio string) error {

	api := fmt.Sprintf("/api/1.0/volume/quorum/ratio")
	u := fmt.Sprintf("%s%s", r.addr, api)
	logrus.Debugf("POST %s", u)

	quorumType := fmt.Sprintf("server")

	params := url.Values{
		"type":  {quorumType},
		"ratio": {ratio},
	}

	logrus.Debugf("Params %s", params)

	resp, err := r.sendRequest("POST", u, params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return responseCheck(resp)
}

func (r Client) sendRequest(reqType, URL string, params url.Values) (*http.Response, error) {
	retry := 5 // Retry 5 times after EOF error

	var res *http.Response
	var err error

	for retry > 0 {

		res, err = r.sendRequestDefault(reqType, URL, params)
		if err != nil {
			// Look for EOF error
			if strings.Contains(err.Error(), "EOF") {
				retry--
				if retry == 0 {
					return nil, err
				}
				time.Sleep(2 * time.Second)
				continue
			}
			return nil, err
		}
		break
	}
	return res, err
}

func (r Client) sendRequestDefault(reqType, URL string, params url.Values) (*http.Response, error) {

	var body *strings.Reader

	if len(params) > 0 {
		body = strings.NewReader(params.Encode())
	} else {
		body = strings.NewReader("")
	}

	req, err := http.NewRequest(reqType, URL, body)

	if len(params) > 0 {
		// Set content header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	}

	// Set authorisation header
	req.SetBasicAuth(r.user, r.password)

	if err != nil {
		return nil, fmt.Errorf("Error in Request: %v", err)
	}

	// Get a HTTP or HTTPS cleint based on TLS
	c := r.mgHTTP()
	if c == nil {
		return nil, fmt.Errorf("Error Setting HTTPS Client ")
	}

	resp, err := c.Do(req)
	// resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error in Response: %v", err)
	}

	return resp, err
}

func (r Client) mgHTTP() *http.Client {

	if !r.TLS {
		// default http client
		return http.DefaultClient
	}

	// TLS
	// Prepare a TLS config
	caCert, err := ioutil.ReadFile(r.cacert)
	if err != nil {
		return nil
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cert, err := tls.LoadX509KeyPair(r.cert, r.key)
	if err != nil {
		logrus.Errorf("Error loading keys %v", err)
		return nil
	}

	tlsc := &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: false, //self signed certificate
		Certificates:       []tls.Certificate{cert},
	}

	http.DefaultClient.Timeout = httpTimeout * time.Second

	transport := &http.Transport{
		TLSClientConfig: tlsc,
	}

	//adding the Transport object to the http Client
	return &http.Client{
		Transport: transport,
	}
}
