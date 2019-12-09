package rwogluster

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	mg "rwogluster"
	"testing"
)

func GlusterLibClient(server *httptest.Server) *mg.Client {
	//URL := "http://localhost:5000"

	cl := mg.NewClient(server.URL, "admin", "intel123")
	return cl
}

func TestListPeers(t *testing.T) {

	endPoint := "/api/1.0/peers"
	status := "CONNECTED"
	id := "123"
	name := "gluster01"

	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		switch request.URL.RequestURI() {
		case endPoint:
			response := response{Ok: true, Err: ""}
			peersList := []Peer{Peer{Name: name, Status: status, ID: id}}
			peerResponse := peerResponse{Data: peersList, response: response}
			jsonData, _ := json.Marshal(peerResponse)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write(jsonData)

		}
	}))
	defer testServer.Close()

	glusterClient := mg.NewClient(testServer.URL, "admin", "intel123")
	//peers := []Peer{Peer{Name: name, Status: status, ID: id}}
	icreateVol, err := glusterClient.GetPeers()
	t.Log(icreateVol, err)

}

func TestCreateGlusterVolume(t *testing.T) {

	volumeTobeCreated := "vol1"
	endPoint := "/api/1.0/volume/"
	endPointWithVolume := fmt.Sprintf("%s%s", endPoint, volumeTobeCreated)

	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		switch request.URL.RequestURI() {
		case endPointWithVolume:
			var createResponse response
			if request.Method == http.MethodPost {
				body, err := ioutil.ReadAll(request.Body)
				if err != nil {
					createResponse = response{Ok: false, Err: "Not able to ready body"}
					break
				}
				if len(body) < 0 {
					createResponse = response{Ok: false, Err: "body of the request is empty"}
					break
				}
				createResponse = response{Ok: true, Err: ""}
			} else {
				createResponse = response{Ok: false, Err: "unknown HTTP method"}
			}
			jsonData, _ := json.Marshal(createResponse)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write(jsonData)
		}
	}))
	defer testServer.Close()

	glusterClient := GlusterLibClient(testServer)
	//peers := []Peer{Peer{Name: name, Status: status, ID: id}}
	brick := fmt.Sprintf("%s:%s/%s", "test", "test path", "vol1")
	vol := mg.GlusterVolume{
		Name:   "vol1",
		Bricks: []string{brick},
		Force:  1,
	}
	if createVolErr := glusterClient.CreateGlusterVolume(vol); createVolErr != nil {
		t.Fatalf("Volume creation failed with following error %s", createVolErr)
	}
}

func TestClient_VolumeExist(t *testing.T) {
	existingVolume := "vol1"

	endPoint := "/api/1.0/volumes"

	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		switch request.URL.RequestURI() {
		case endPoint:
			response := response{Ok: true, Err: ""}
			volumesList := []volume{volume{Name: existingVolume}}
			volumeResponse := volumeResponse{Data: volumesList, response: response}
			jsonData, _ := json.Marshal(volumeResponse)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write(jsonData)

		}
	}))
	defer testServer.Close()
	glusterClient := GlusterLibClient(testServer)

	res, err := glusterClient.VolumeExist("vol1")
	if err != nil {
		t.Fatalf("Volume exists check failed  %s", err.Error())
	}
	if res == true {
		t.Log(existingVolume, "Volume exists")
	}

}

func TestClient_StopVolume(t *testing.T) {
	VolumeToBeStopped := "vol1"
	endPointWithVolume := fmt.Sprintf("/api/1.0/volume/%s/stop", VolumeToBeStopped)

	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		switch request.URL.RequestURI() {
		case endPointWithVolume:
			var createResponse response
			if request.Method == http.MethodPut {
				createResponse = response{Ok: true, Err: ""}
			} else {
				createResponse = response{Ok: true, Err: "unknow HTTP method"}
			}
			jsonData, _ := json.Marshal(createResponse)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write(jsonData)
		}
	}))
	defer testServer.Close()

	glusterClient := GlusterLibClient(testServer)

	err := glusterClient.StopVolume("vol1")
	if err != nil {
		t.Fatalf("Volume Stop failed  %s", err.Error())
	}

}

func TestClient_StartVolume(t *testing.T) {
	VolumeToBeStarted := "vol1"
	endPointWithVolume := fmt.Sprintf("/api/1.0/volume/%s/start", VolumeToBeStarted)

	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		switch request.URL.RequestURI() {
		case endPointWithVolume:
			var createResponse response
			if request.Method == http.MethodPut {
				createResponse = response{Ok: true, Err: ""}
			} else {
				createResponse = response{Ok: true, Err: "unknow HTTP method"}
			}
			jsonData, _ := json.Marshal(createResponse)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write(jsonData)
		}
	}))
	defer testServer.Close()

	glusterClient := GlusterLibClient(testServer)

	err := glusterClient.StartVolume("vol1")
	if err != nil {
		t.Fatalf("Volume Start failed  %s", err.Error())
	}

}

func TestClient_GetBricks(t *testing.T) {
	existingVolume := "vol1"

	endPoint := "/api/1.0/volumes"
	bricks := []Brickinfo{{
		HostUUID: "testUUID",
		Name:     "brick1",
	}}

	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		switch request.URL.RequestURI() {
		case endPoint:
			response := response{Ok: true, Err: ""}
			volumesList := []volume{volume{Name: existingVolume, Bricks: bricks}}
			volumeResponse := volumeResponse{Data: volumesList, response: response}
			jsonData, _ := json.Marshal(volumeResponse)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write(jsonData)

		}
	}))
	defer testServer.Close()
	glusterClient := GlusterLibClient(testServer)

	brickData, err := glusterClient.GetBricks("vol1")
	if err != nil {
		t.Fatalf("Get Bricks failed %s", err.Error())
	}
	t.Log("bricks are ", brickData)

}

func TestClient_GetPeers(t *testing.T) {

	endPoint := "/api/1.0/peers"
	status := "CONNECTED"
	id := "123"
	name := "gluster01"

	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		switch request.URL.RequestURI() {
		case endPoint:
			response := response{Ok: true, Err: ""}
			peersList := []Peer{Peer{Name: name, Status: status, ID: id}}
			peerResponse := peerResponse{Data: peersList, response: response}
			jsonData, _ := json.Marshal(peerResponse)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write(jsonData)

		}
	}))

	defer testServer.Close()
	glusterClient := GlusterLibClient(testServer)

	peers, err := glusterClient.GetPeers()
	if err != nil {
		t.Fatalf("Get Peers Failed  %s", err.Error())
	}
	t.Log("peers are", peers)
}

func TestClient_PeerDetach(t *testing.T) {

	hostname := "peer1"
	endPoint := fmt.Sprintf("/api/1.0/peer/%s", hostname)

	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		switch request.URL.RequestURI() {
		case endPoint:
			response := response{Ok: true, Err: ""}
			jsonData, _ := json.Marshal(response)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write(jsonData)

		}
	}))

	defer testServer.Close()
	glusterClient := GlusterLibClient(testServer)
	err := glusterClient.PeerDetach("peer1")
	if err != nil {
		t.Fatalf("Peer Detach Failed %s", err.Error())
	}

}

func TestClient_PeerStatus(t *testing.T) {

	endPoint := "/api/1.0/peers"
	status := "CONNECTED"
	id := "123"
	name := "gluster01"

	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		switch request.URL.RequestURI() {
		case endPoint:
			response := response{Ok: true, Err: ""}
			peersList := []Peer{Peer{Name: name, Status: status, ID: id}}
			peerResponse := peerResponse{Data: peersList, response: response}
			jsonData, _ := json.Marshal(peerResponse)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write(jsonData)

		}
	}))

	defer testServer.Close()
	glusterClient := GlusterLibClient(testServer)

	peers, err := glusterClient.PeerStatus()
	if err != nil {
		t.Fatalf("Peer Status failed %s", err.Error())
	}
	t.Log("peers are", peers)

}

func TestClient_PeerProbeWithMsg(t *testing.T) {

	hostname := "peer1"
	endPoint := fmt.Sprintf("/api/1.0/peer/%s", hostname)
	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		switch request.URL.RequestURI() {
		case endPoint:
			response := response{Ok: true, Err: ""}
			jsonData, _ := json.Marshal(response)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write(jsonData)

		}
	}))

	defer testServer.Close()
	glusterClient := GlusterLibClient(testServer)

	msg, err := glusterClient.PeerProbeWithMsg(hostname)
	if err != nil {
		t.Fatalf("Peer probe failed  %s", err.Error())
	}
	t.Log("peer probe was ", msg)

}

func TestClient_RemoveVolume(t *testing.T) {

	VolumeToBeDeleted := "vol1"
	endPointWithVolume := fmt.Sprintf("/api/1.0/volume/%s", VolumeToBeDeleted)

	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		switch request.URL.RequestURI() {
		case endPointWithVolume:
			var createResponse response
			if request.Method == http.MethodPut {
				createResponse = response{Ok: true, Err: ""}
			} else {
				createResponse = response{Ok: true, Err: "unknow HTTP method"}
			}
			jsonData, _ := json.Marshal(createResponse)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write(jsonData)
		}
	}))
	defer testServer.Close()

	glusterClient := GlusterLibClient(testServer)

	err := glusterClient.RemoveVolume(VolumeToBeDeleted)
	if err != nil {
		t.Fatalf("Volume remove failed  %s", err.Error())
	}

}

func TestClient_AddBrick(t *testing.T) {

	volumeName := "vol1"
	endPoint := fmt.Sprintf("/api/1.0/volume/%s/addbrick", volumeName)

	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		switch request.URL.RequestURI() {
		case endPoint:
			response := response{Ok: true, Err: ""}
			jsonData, _ := json.Marshal(response)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write(jsonData)

		}
	}))

	defer testServer.Close()
	glusterClient := GlusterLibClient(testServer)
	err := glusterClient.AddBrick(volumeName, "brick", 2, 2)
	if err != nil {
		t.Fatalf("Add brick failed  %s", err.Error())
	}

}

func TestClient_RemoveBrick(t *testing.T) {
	volumeName := "vol1"
	endPoint := fmt.Sprintf("/api/1.0/volume/%s/removebrick", volumeName)

	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		switch request.URL.RequestURI() {
		case endPoint:
			response := response{Ok: true, Err: ""}
			jsonData, _ := json.Marshal(response)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write(jsonData)

		}
	}))

	defer testServer.Close()
	glusterClient := GlusterLibClient(testServer)
	err := glusterClient.RemoveBrick(volumeName, "brick", 2)
	if err != nil {
		t.Fatalf(" Remove brick Failed %s", err.Error())
	}

}
