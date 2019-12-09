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

func Test_leaderState(t *testing.T) {
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
			if err := leaderState(); (err != nil) != tt.wantErr {
				t.Errorf("leaderState() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_workerState(t *testing.T) {
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
			if err := workerState(); (err != nil) != tt.wantErr {
				t.Errorf("workerState() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_memberUpdateCheckAndSetTag(t *testing.T) {
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
			if err := memberUpdateCheckAndSetTag(); (err != nil) != tt.wantErr {
				t.Errorf("memberUpdateCheckAndSetTag() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_manageSwarmStatus(t *testing.T) {
	type args struct {
		swarmManager bool
		statusSwarm  string
	}
	tests := []struct {
		name     string
		args     args
		sunnyday bool
		wantErr  bool
	}{
		// TODO: More tests can be added
		{name: "TC1", wantErr: false, sunnyday: true, args: args{swarmManager: true, statusSwarm: "reachable"}},
		{name: "TC2", wantErr: true, sunnyday: false, args: args{swarmManager: true, statusSwarm: "reachable"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			helpers.SunnyDay = tt.sunnyday
			if err := manageSwarmStatus(tt.args.swarmManager, tt.args.statusSwarm); (err != nil) != tt.wantErr {
				t.Errorf("manageSwarmStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_memberUpdateWorker(t *testing.T) {
	tests := []struct {
		name     string
		sunnyday bool
		wantErr  bool
	}{
		{name: "TC1", wantErr: false, sunnyday: true},
		{name: "TC2", wantErr: true, sunnyday: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			helpers.SunnyDay = tt.sunnyday
			if err := memberUpdateWorker(); (err != nil) != tt.wantErr {
				t.Errorf("memberUpdateWorker() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_memberUpdateLeader(t *testing.T) {
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
			if err := memberUpdateLeader(); (err != nil) != tt.wantErr {
				t.Errorf("memberUpdateLeader() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
