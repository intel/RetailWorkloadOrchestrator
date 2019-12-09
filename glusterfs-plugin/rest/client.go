package rest

// https://github.com/calavera/docker-volume-glusterfs/blob/master/rest/client.gopackage rest

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	volumesPath      = "/api/1.0/volumes"
	volumeCreatePath = "/api/1.0/volume/%s"
	volumeStopPath   = "/api/1.0/volume/%s/stop"
	peersPath        = "/api/1.0/peers"
	connectStatus    = "CONNECTED"
	httpTimeout      = 10
)

type peer struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type volume struct {
	Name       string `json:"name"`
	UUID       string `json:"uuid"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	NumBricks  int    `json:"num_bricks"`
	Distribute int    `json:"distribute"`
	Stripe     int    `json:"stripe"`
	Replica    int    `json:"replica"`
	Transport  string `json:"transport"`
}

type response struct {
	Ok  bool   `json:"ok"`
	Err string `json:"error,omitempty"`
}

type peerResponse struct {
	Data []peer `json:"data,omitempty"`
	response
}

type volumeResponse struct {
	Data []volume `json:"data,omitempty"`
	response
}

// Client is the http client that sends requests to the gluster API.
type Client struct {
	addr              string
	base              string
	user, password    string
	TLS               bool
	key, cert, cacert string // paths for key, cert and cacert
}

// NewClient initializes a new client.
func NewClient(addr, base, user, password string) *Client {
	return &Client{
		addr:     addr,
		base:     base,
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
	u := fmt.Sprintf("%s%s", r.addr, volumesPath)
	logrus.Debugf("GET %s", u)

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

// CreateVolume creates a new volume with the given name in the cluster.
func (r Client) CreateVolume(name string, peers []string) error {
	u := fmt.Sprintf("%s%s", r.addr, fmt.Sprintf(volumeCreatePath, name))
	logrus.Debugf("POST %s", u)

	var bricks []string

	for _, peer := range peers {
		bricks = append(bricks, fmt.Sprintf("%s:%s", peer, filepath.Join(r.base, name)))
	}

	params := url.Values{
		"bricks":    {strings.Join(bricks, ",")},
		"transport": {"tcp"},
		"start":     {"true"},
		"force":     {"true"},
	}

	if len(bricks) > 1 {
		params["replica"] = []string{strconv.Itoa(len(peers))}
	}

	logrus.Debugf("Params %s", params)

	resp, err := r.sendRequest("POST", u, params)
	if err != nil {
		return err
	}

	return responseCheck(resp)
}

// StopVolume stops the volume with the given name in the cluster.
func (r Client) StopVolume(name string) error {
	u := fmt.Sprintf("%s%s", r.addr, fmt.Sprintf(volumeStopPath, name))

	resp, err := r.sendRequest("PUT", u, nil)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	return responseCheck(resp)
}

// GetPeers makes a RESTful GET call to all peers in the gluster pool
func (r Client) GetPeers() ([]string, error) {

	endPoint := fmt.Sprintf("%s%s", r.addr, peersPath)
	logrus.Debugf("GET %s", endPoint)

	res, err := r.sendRequest("GET", endPoint, nil)
	if err != nil {
		logrus.Errorf("Error GET peers, %s", err.Error())
		return nil, err
	}

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

func (r Client) sendRequest(reqType, URL string, params url.Values) (*http.Response, error) {

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
