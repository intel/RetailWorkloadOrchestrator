//Copyright 2019, Intel Corporation

package main

import (
	"helpers"
	"testing"
)

func TestInitHandler(t *testing.T) {
	tests := []struct {
		name          string
		expectedError bool
		sunnyday      bool
	}{
		{name: "T1", expectedError: false, sunnyday: true},
		//		{name: "T2", expectedError: true, sunnyday: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedError := false
			helpers.SunnyDay = tt.sunnyday

			err := InitHandler()

			if err != nil {
				expectedError = true
			}

			if expectedError != tt.expectedError {
				t.Errorf("%s:%v ", tt.name, tt.expectedError)
			}

		})
	}
}

func TestManageLeaderWorker(t *testing.T) {
	//memberJoinManageLeaderWorker()

	tests := []struct {
		name          string
		expectedError bool
		sunnyday      bool
	}{
		{name: "T1", expectedError: false, sunnyday: true},
		{name: "T2", expectedError: true, sunnyday: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedError := false

			helpers.SunnyDay = tt.sunnyday

			err := memberJoinManageLeaderWorker()
			if err != nil {
				expectedError = true
			}
			if expectedError != tt.expectedError {
				t.Errorf("%s:%v ", tt.name, tt.expectedError)
			}
		})
	}
}

func Test_handleRoleAndSwarm(t *testing.T) {
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
			client = helpers.GlusterLibClient()

			if err := handleRoleAndSwarm(); (err != nil) != tt.wantErr {
				t.Errorf("handleRoleAndSwarm() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_memberJoinManageLeaderWorker(t *testing.T) {
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
			client = helpers.GlusterLibClient()

			if err := memberJoinManageLeaderWorker(); (err != nil) != tt.wantErr {
				t.Errorf("memberJoinManageLeaderWorker() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_memberJoinAssignLeaderWorker(t *testing.T) {
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
			client = helpers.GlusterLibClient()

			if err := memberJoinAssignLeaderWorker(); (err != nil) != tt.wantErr {
				t.Errorf("memberJoinAssignLeaderWorker() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_checkSwarmAndGlusterStatus(t *testing.T) {
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
			client = helpers.GlusterLibClient()

			if err := checkSwarmAndGlusterStatus(); (err != nil) != tt.wantErr {
				t.Errorf("checkSwarmAndGlusterStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
