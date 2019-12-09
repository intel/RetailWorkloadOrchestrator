package main

import (
	"github.com/sapk/docker-volume-gluster/gluster"
	"github.com/sirupsen/logrus"
)

var (
	//Version version of app set by build flag
	Version string
	//Branch git branch of app set by build flag
	Branch string
	//Commit git commit of app set by build flag
	Commit string
	//BuildTime build time of app set by build flag
	BuildTime string
)

func main() {
	logrus.Print("1.0.0")
	gluster.Version = Version
	gluster.Commit = Commit
	gluster.Branch = Branch
	gluster.BuildTime = BuildTime
	gluster.Init()
}
