//Copyright 2019, Intel Corporation

package rwogluster

// GlusterVolume represents a gluster volume
type GlusterVolume struct {
	Name                          string
	Bricks                        []string
	Replica, Stripe, Force, Start int
	Transport                     string
}

// Peer represents gluster peer
type Peer struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// Brickinfo represents gluster peer
type Brickinfo struct {
	HostUUID string `json:"hostUuid"`
	Name     string `json:"name"`
}

type volume struct {
	Name       string      `json:"name"`
	UUID       string      `json:"uuid"`
	Bricks     []Brickinfo `json:"bricks"`
	Type       string      `json:"type"`
	Status     string      `json:"status"`
	NumBricks  int         `json:"num_bricks"`
	Distribute int         `json:"distribute"`
	Stripe     int         `json:"stripe"`
	Replica    int         `json:"replica"`
	Transport  string      `json:"transport"`
}

type response struct {
	Ok  bool   `json:"ok"`
	Err string `json:"error,omitempty"`
}

type peerResponse struct {
	Data []Peer `json:"data,omitempty"`
	response
}

type volumeResponse struct {
	Data []volume `json:"data,omitempty"`
	response
}

// Client is the http client that sends requests to the gluster API.
type Client struct {
	addr, user, password string
	TLS                  bool
	key, cert, cacert    string // paths for key, cert and cacert
}
