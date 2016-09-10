package driver

import (
	"testing"

	"github.com/hashicorp/nomad/client/config"
	"github.com/hashicorp/nomad/nomad/structs"
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
