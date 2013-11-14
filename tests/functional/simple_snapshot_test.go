package test

import (
	"io/ioutil"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/coreos/go-etcd/etcd"
)

// This test creates a single node and then set a value to it to trigger snapshot
func TestSimpleSnapshot(t *testing.T) {
	procAttr := new(os.ProcAttr)
	procAttr.Files = []*os.File{nil, os.Stdout, os.Stderr}
	args := []string{"etcd", "-n=node1", "-d=/tmp/node1", "-snapshot=true", "-snapshotCount=500"}

	process, err := os.StartProcess(EtcdBinPath, append(args, "-f"), procAttr)
	if err != nil {
		t.Fatal("start process failed:" + err.Error())
	}
	defer process.Kill()

	time.Sleep(time.Second)

	c := etcd.NewClient(nil)

	c.SyncCluster()
	// issue first 501 commands
	for i := 0; i < 501; i++ {
		result, err := c.Set("foo", "bar", 100)

		if err != nil || result.Key != "/foo" || result.Value != "bar" || result.TTL < 95 {
			if err != nil {
				t.Fatal(err)
			}

			t.Fatalf("Set failed with %s %s %v", result.Key, result.Value, result.TTL)
		}
	}

	// wait for a snapshot interval
	time.Sleep(3 * time.Second)

	snapshots, err := ioutil.ReadDir("/tmp/node1/snapshot")

	if err != nil {
		t.Fatal("list snapshot failed:" + err.Error())
	}

	if len(snapshots) != 1 {
		t.Fatal("wrong number of snapshot :[1/", len(snapshots), "]")
	}

	index, _ := strconv.Atoi(snapshots[0].Name()[2:5])

	if index < 507 || index > 510 {
		t.Fatal("wrong name of snapshot :", snapshots[0].Name())
	}

	// issue second 501 commands
	for i := 0; i < 501; i++ {
		result, err := c.Set("foo", "bar", 100)

		if err != nil || result.Key != "/foo" || result.Value != "bar" || result.TTL < 95 {
			if err != nil {
				t.Fatal(err)
			}

			t.Fatalf("Set failed with %s %s %v", result.Key, result.Value, result.TTL)
		}
	}

	// wait for a snapshot interval
	time.Sleep(3 * time.Second)

	snapshots, err = ioutil.ReadDir("/tmp/node1/snapshot")

	if err != nil {
		t.Fatal("list snapshot failed:" + err.Error())
	}

	if len(snapshots) != 1 {
		t.Fatal("wrong number of snapshot :[1/", len(snapshots), "]")
	}

	index, _ = strconv.Atoi(snapshots[0].Name()[2:6])

	if index < 1015 || index > 1018 {
		t.Fatal("wrong name of snapshot :", snapshots[0].Name())
	}
}
