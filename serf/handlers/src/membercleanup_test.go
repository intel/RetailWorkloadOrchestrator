//Copyright 2019, Intel Corporation

package main

import (
	"fmt"
	"helpers"
	"testing"
)

var (
	rwolog helpers.Logger
)

func Test_memberCleanup(t *testing.T) {
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
			if err := memberCleanup(); (err != nil) != tt.wantErr {
				t.Errorf("memberCleanup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_removeMember(t *testing.T) {
	type args struct {
		nodeName string
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		sunnyday bool
	}{
		{name: "TC1", wantErr: false, sunnyday: true, args: args{nodeName: "V1"}},
		{name: "TC2", wantErr: true, sunnyday: false, args: args{nodeName: "V1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.SunnyDay = tt.sunnyday
			client = helpers.GlusterLibClient()
			fmt.Println(tt.args.nodeName)
			if err := removeMember(tt.args.nodeName); (err != nil) != tt.wantErr {
				t.Errorf("removeMember() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_removeDockerNodes(t *testing.T) {
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

			if err := removeDockerNodes(); (err != nil) != tt.wantErr {
				t.Errorf("removeDockerNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_glusterCleanup(t *testing.T) {
	type args struct {
		node string
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		sunnyday bool
	}{
		{name: "TC1", wantErr: false, sunnyday: true, args: args{node: "V1"}},
		{name: "TC2", wantErr: true, sunnyday: false, args: args{node: "V1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.SunnyDay = tt.sunnyday
			client = helpers.GlusterLibClient()

			if err := glusterCleanup(tt.args.node); (err != nil) != tt.wantErr {
				t.Errorf("glusterCleanup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_removeNode(t *testing.T) {
	type args struct {
		node string
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		sunnyday bool
	}{
		{name: "TC1", wantErr: false, sunnyday: true, args: args{node: "V1"}},
		{name: "TC2", wantErr: true, sunnyday: false, args: args{node: "V1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.SunnyDay = tt.sunnyday
			client = helpers.GlusterLibClient()
			if err := removeNode(tt.args.node); (err != nil) != tt.wantErr {
				t.Errorf("removeNode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_removeGlusterBrick(t *testing.T) {
	type args struct {
		glusterHostToDetach string
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		sunnyday bool
	}{
		{name: "TC1", wantErr: false, sunnyday: true, args: args{glusterHostToDetach: "V1"}},
		//	{name: "TC2", wantErr: true, sunnyday: false, args: args{glusterHostToDetach: "V1"}},
		// TODO: Fix the second case
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.SunnyDay = tt.sunnyday
			client = helpers.GlusterLibClient()
			if err := removeGlusterBrick(tt.args.glusterHostToDetach); (err != nil) != tt.wantErr {
				t.Errorf("removeGlusterBrick() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
