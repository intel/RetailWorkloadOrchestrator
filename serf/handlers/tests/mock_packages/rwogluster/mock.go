package rwogluster

import (
	"fmt"
)

var (
	SunnyDay = true
)

// NewClient initializes a new client.
func NewClient(addr, user, password string) *Client {
	return &Client{
		addr:     addr,
		user:     user,
		password: password,
	}
}

// EnableTLS enables TLS for client. It takes paths for key, cert and cacert as input.
func (r *Client) EnableTLS(key, cert, cacert string) error {

	// validate input params
	if len(key) == 0 || len(cert) == 0 || len(cacert) == 0 {
		return fmt.Errorf("key/cert paths not valid")
	}

	r.TLS = true
	return nil
}

// CreateGlusterVolume creates a gluster volume.
func (r Client) CreateGlusterVolume(Vol GlusterVolume) error {
	if SunnyDay {
		return nil
	}

	return fmt.Errorf("couldn't create a volume")

}

// VolumeExist returns whether a volume exist in the cluster with a given name or not.
func (r Client) VolumeExist(name string) (bool, error) {
	if SunnyDay {
		return false, nil
	}
	return true, nil
}

// StopVolume stops the volume with the given name in the cluster.
func (r Client) StopVolume(name string) error {

	if SunnyDay {
		return nil
	}

	return fmt.Errorf("couldn't stop volume")
}

// GetBricks list bricks in the gluster volume
func (r Client) GetBricks(vol string) ([]Brickinfo, error) {
	if SunnyDay {
		return []Brickinfo{{"XYZABC", "/mnt/bricks/b1"}}, nil
	}

	return nil, fmt.Errorf("No Bricks")
}

// GetPeers makes a RESTful GET call to all peers in the gluster pool
func (r Client) GetPeers() ([]string, error) {
	if SunnyDay {
		return []string{"ABC", "localhost", "Connected"}, nil
	}

	return []string{""}, nil
}

// StartVolume starts a volume
func (r Client) StartVolume(name string) error {
	if SunnyDay {
		return nil
	}

	return fmt.Errorf("Error starting volumes")

}

// ListVolumes lists volumes
func (r Client) ListVolumes() ([]GlusterVolume, error) {
	if SunnyDay {
		return []GlusterVolume{{
			Name:      "Demo",
			Bricks:    []string{"192.168.0.1:/mnt/bricks/b1", "192.168.0.2:/mnt/bricks/b1"},
			Transport: "tcp",
			Replica:   2,
			Start:     1,
		},
		}, nil
	}
	return []GlusterVolume{}, fmt.Errorf("gluster server is not up")
}

// PeerDetach detaches peer
func (r Client) PeerDetach(hostname string) error {
	if SunnyDay {
		return nil
	}
	return fmt.Errorf("some members are down")
}

// PeerStatus gives all peers with status connected or disconnected
func (r Client) PeerStatus() ([]Peer, error) {
	if SunnyDay {
		return []Peer{{"1234", "v1", "Connected"}, {"1234", "v2", "Connected"}}, nil
	}
	return []Peer{}, fmt.Errorf("some operation already in progress")
}

// PeerProbeWithMsg implements gluster peer probe and returns the message
func (r Client) PeerProbeWithMsg(hostname string) (string, error) {
	if SunnyDay {
		return "success", nil
	}
	return "error", fmt.Errorf("failed to connect")
}

// PeerProbe implements gluster peer probe
func (r Client) PeerProbe(hostname string) error {
	if SunnyDay {
		return nil
	}
	return fmt.Errorf("failed to connect")

}

// CleanupVolume does proper cleanup of the volume. It will stop the volume. remove bricks and detach from all peers
func (r Client) CleanupVolume(name string) error {
	if SunnyDay {
		return nil
	}
	return fmt.Errorf("volume in use")
}

// RemoveVolume removes gluster Volume
func (r Client) RemoveVolume(name string) error {
	if SunnyDay {
		return nil
	}
	return fmt.Errorf("volume in use")

}

// AddBrick adds bricks to an existing volume
func (r Client) AddBrick(vol string, b string, rep int, s int) error {
	if SunnyDay {
		return nil
	}
	return fmt.Errorf("gluster not up")
}

// RemoveBrick removes bricks from an existing volume
func (r Client) RemoveBrick(vol, brick string, rep int) error {
	if SunnyDay {
		return nil
	}
	return fmt.Errorf("gluster not up")
}
