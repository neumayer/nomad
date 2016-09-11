package driver

import (
	"testing"
	"time"

	"github.com/hashicorp/nomad/client/config"
	"github.com/hashicorp/nomad/nomad/structs"
	"github.com/hashicorp/nomad/testutil"
)

func TestLxcDriver_Fingerprint(t *testing.T) {
	task := &structs.Task{
		Name:      "foo",
		Resources: structs.DefaultResources(),
	}

	driverCtx, execCtx := testDriverContexts(task)
	defer execCtx.AllocDir.Destroy()
	d := NewLxcDriver(driverCtx)
	node := &structs.Node{
		Attributes: map[string]string{},
	}
	apply, err := d.Fingerprint(&config.Config{}, node)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !apply {
		t.Fatalf("should apply")
	}
	if node.Attributes["driver.lxc"] == "" {
		t.Fatalf("missing driver")
	}
}

func TestLxcDriver_Start_Wait(t *testing.T) {
	task := &structs.Task{
		Name: "foo",
		Config: map[string]interface{}{
			"template": "/usr/share/lxc/templates/lxc-busybox",
		},
		Resources: structs.DefaultResources(),
	}

	driverCtx, execCtx := testDriverContexts(task)
	defer execCtx.AllocDir.Destroy()
	d := NewLxcDriver(driverCtx)

	handle, err := d.Start(execCtx, task)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if handle == nil {
		t.Fatalf("missing handle")
	}

	select {
	case res := <-handle.WaitCh():
		if !res.Successful() {
			t.Fatalf("err: %v", res)
		}
	case <-time.After(time.Duration(testutil.TestMultiplier()*5) * time.Second):
		t.Fatalf("timeout")
	}
}
