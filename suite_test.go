// mgo - MongoDB driver for Go

package mgo_test

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"testing"
	"time"

	mgo "github.com/initial-commit-hq/mgo"
	"github.com/initial-commit-hq/mgo/bson"
	. "gopkg.in/check.v1"
)

var fast = flag.Bool("fast", false, "Skip slow tests")

type M bson.M

type cLogger C

func (c *cLogger) Output(calldepth int, s string) error {
	ns := time.Now().UnixNano()
	t := float64(ns%100e9) / 1e9
	((*C)(c)).Logf("[LOG] %.05f %s", t, s)
	return nil
}

func TestAll(t *testing.T) {
	TestingT(t)
}

type S struct {
	session *mgo.Session
	stopped bool
	build   mgo.BuildInfo
	frozen  []string
}

func (s *S) versionAtLeast(v ...int) (result bool) {
	for i := range v {
		if i == len(s.build.VersionArray) {
			return false
		}
		if s.build.VersionArray[i] != v[i] {
			return s.build.VersionArray[i] >= v[i]
		}
	}
	return true
}

var _ = Suite(&S{})

func (s *S) SetUpSuite(c *C) {
	mgo.SetDebug(true)
	mgo.SetStats(true)
	s.StartAll()

	session, err := mgo.Dial("localhost:40001")
	c.Assert(err, IsNil)
	s.build, err = session.BuildInfo()
	c.Check(err, IsNil)
	session.Close()
}

func (s *S) SetUpTest(c *C) {
	err := run("mongo --nodb harness/mongojs/dropall.js")
	if err != nil {
		panic(err.Error())
	}
	mgo.SetLogger((*cLogger)(c))
	mgo.ResetStats()
}

func (s *S) TearDownTest(c *C) {
	if s.stopped {
		s.Stop(":40201")
		s.Stop(":40202")
		s.Stop(":40203")
		s.StartAll()
	}
	for _, host := range s.frozen {
		if host != "" {
			s.Thaw(host)
		}
	}
	var stats mgo.Stats
	for i := 0; ; i++ {
		stats = mgo.GetStats()
		if stats.SocketsInUse == 0 && stats.SocketsAlive == 0 {
			break
		}
		if i == 20 {
			c.Fatal("Test left sockets in a dirty state")
		}
		c.Logf("Waiting for sockets to die: %d in use, %d alive", stats.SocketsInUse, stats.SocketsAlive)
		time.Sleep(500 * time.Millisecond)
	}
	for i := 0; ; i++ {
		stats = mgo.GetStats()
		if stats.Clusters == 0 {
			break
		}
		if i == 60 {
			c.Fatal("Test left clusters alive")
		}
		c.Logf("Waiting for clusters to die: %d alive", stats.Clusters)
		time.Sleep(1 * time.Second)
	}
}

func (s *S) Stop(host string) {
	// Give a moment for slaves to sync and avoid getting rollback issues.
	panicOnWindows()
	time.Sleep(2 * time.Second)
	err := run("svc -d _harness/daemons/" + supvName(host))
	if err != nil {
		panic(err)
	}
	s.stopped = true
}

func (s *S) pid(host string) int {
	// Note recent releases of lsof force 'f' to be present in the output (WTF?).
	cmd := exec.Command("lsof", "-iTCP:"+hostPort(host), "-sTCP:LISTEN", "-Fpf")
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	pidstr := string(bytes.Fields(output[1:])[0])
	pid, err := strconv.Atoi(pidstr)
	if err != nil {
		panic(fmt.Errorf("cannot convert pid to int: %q, command line: %q", pidstr, cmd.Args))
	}
	return pid
}

func (s *S) Freeze(host string) {
	err := stop(s.pid(host))
	if err != nil {
		panic(err)
	}
	s.frozen = append(s.frozen, host)
}

func (s *S) Thaw(host string) {
	err := cont(s.pid(host))
	if err != nil {
		panic(err)
	}
	for i, frozen := range s.frozen {
		if frozen == host {
			s.frozen[i] = ""
		}
	}
}

func (s *S) StartAll() {
	if s.stopped {
		// Restart any stopped nodes.
		run("svc -u _harness/daemons/*")
		err := run("mongo --nodb harness/mongojs/wait.js")
		if err != nil {
			panic(err)
		}
		s.stopped = false
	}
}

func run(command string) error {
	var output []byte
	var err error
	if runtime.GOOS == "windows" {
		output, err = exec.Command("cmd", "/C", command).CombinedOutput()
	} else {
		output, err = exec.Command("/bin/sh", "-c", command).CombinedOutput()
	}

	if err != nil {
		msg := fmt.Sprintf("Failed to execute: %s: %s\n%s", command, err.Error(), string(output))
		return errors.New(msg)
	}
	return nil
}

var supvNames = map[string]string{
	"40001": "db1",
	"40002": "db2",
	"40011": "rs1a",
	"40012": "rs1b",
	"40013": "rs1c",
	"40021": "rs2a",
	"40022": "rs2b",
	"40023": "rs2c",
	"40031": "rs3a",
	"40032": "rs3b",
	"40033": "rs3c",
	"40041": "rs4a",
	"40101": "cfg1",
	"40102": "cfg2",
	"40103": "cfg3",
	"40201": "s1",
	"40202": "s2",
	"40203": "s3",
}

// supvName returns the daemon name for the given host address.
func supvName(host string) string {
	host, port, err := net.SplitHostPort(host)
	if err != nil {
		panic(err)
	}
	name, ok := supvNames[port]
	if !ok {
		panic("Unknown host: " + host)
	}
	return name
}

func hostPort(host string) string {
	_, port, err := net.SplitHostPort(host)
	if err != nil {
		panic(err)
	}
	return port
}

func panicOnWindows() {
	if runtime.GOOS == "windows" {
		panic("the test suite is not yet fully supported on Windows")
	}
}
