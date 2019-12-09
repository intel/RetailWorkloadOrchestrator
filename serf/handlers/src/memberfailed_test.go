//Copyright 2019, Intel Corporation

package main

import (
	"helpers"
	"testing"
)

func TestInitHandler(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
		{name: "T1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InitHandler()
		})
	}
}

func Test_handleSwarmAsReachable(t *testing.T) {
	type args struct {
		aliveMembers int
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		sunnyday bool
	}{
		{name: "TC1", wantErr: false, sunnyday: true, args: args{aliveMembers: 2}},
		{name: "TC2", wantErr: true, sunnyday: false, args: args{aliveMembers: 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.SunnyDay = tt.sunnyday

			if err := handleSwarmAsReachable(tt.args.aliveMembers); (err != nil) != tt.wantErr {
				t.Errorf("handleSwarmAsReachable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_handleSwarmAsLeader(t *testing.T) {
	tests := []struct {
		name     string
		wantErr  bool
		sunnyday bool
	}{
		{name: "TC1", wantErr: false, sunnyday: true},
		{name: "TC2", wantErr: true, sunnyday: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.SunnyDay = tt.sunnyday

			if err := handleSwarmAsLeader(); (err != nil) != tt.wantErr {
				t.Errorf("handleSwarmAsLeader() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manageSwarm(t *testing.T) {
	tests := []struct {
		name     string
		wantErr  bool
		sunnyday bool
	}{
		{name: "TC1", wantErr: false, sunnyday: true},
		{name: "TC2", wantErr: true, sunnyday: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.SunnyDay = tt.sunnyday

			if err := manageSwarm(); (err != nil) != tt.wantErr {
				t.Errorf("manageSwarm() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_setTagAsNodeID(t *testing.T) {
	tests := []struct {
		name     string
		wantErr  bool
		sunnyday bool
	}{
		{name: "TC1", wantErr: false, sunnyday: true},
		{name: "TC2", wantErr: true, sunnyday: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.SunnyDay = tt.sunnyday

			if err := setTagAsNodeID(); (err != nil) != tt.wantErr {
				t.Errorf("setTagAsNodeID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_handleWorker(t *testing.T) {
	tests := []struct {
		name     string
		wantErr  bool
		sunnyday bool
	}{
		{name: "TC1", wantErr: false, sunnyday: true},
		{name: "TC2", wantErr: true, sunnyday: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.SunnyDay = tt.sunnyday

			if err := handleWorker(); (err != nil) != tt.wantErr {
				t.Errorf("handleWorker() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_checkSwarmID(t *testing.T) {
	tests := []struct {
		name     string
		wantErr  bool
		sunnyday bool
	}{
		{name: "TC1", wantErr: false, sunnyday: true},
		{name: "TC2", wantErr: true, sunnyday: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.SunnyDay = tt.sunnyday

			if ID := checkSwarmID(); (len(ID) == 0) != tt.wantErr {
				t.Errorf("handleWorker() error = %v, wantErr %v", (len(ID) == 0), tt.wantErr)
			}
		})
	}
}

func Test_initializeSwarm(t *testing.T) {
	tests := []struct {
		name     string
		wantErr  bool
		sunnyday bool
	}{
		{name: "TC1", wantErr: false, sunnyday: true},
		{name: "TC2", wantErr: true, sunnyday: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.SunnyDay = tt.sunnyday

			if err := initializeSwarm(); (err != nil) != tt.wantErr {
				t.Errorf("initializeSwarm() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_swarmInit(t *testing.T) {
	tests := []struct {
		name     string
		wantErr  bool
		sunnyday bool
	}{
		{name: "TC1", wantErr: false, sunnyday: true},
		{name: "TC2", wantErr: true, sunnyday: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.SunnyDay = tt.sunnyday

			if err := swarmInit(); (err != nil) != tt.wantErr {
				t.Errorf("swarmInit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_performCleanup(t *testing.T) {
	tests := []struct {
		name     string
		wantErr  bool
		sunnyday bool
	}{
		{name: "TC1", wantErr: false, sunnyday: true},
		{name: "TC2", wantErr: true, sunnyday: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.SunnyDay = tt.sunnyday

			if err := performCleanup(); (err != nil) != tt.wantErr {
				t.Errorf("performCleanup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_setSwarmTag(t *testing.T) {
	tests := []struct {
		name     string
		wantErr  bool
		sunnyday bool
	}{
		{name: "TC1", wantErr: false, sunnyday: true},
		{name: "TC2", wantErr: true, sunnyday: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.SunnyDay = tt.sunnyday

			if err := setSwarmTag(); (err != nil) != tt.wantErr {
				t.Errorf("setSwarmTag() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
