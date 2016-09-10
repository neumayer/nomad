package driver

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/nomad/client/config"
	"github.com/hashicorp/nomad/client/fingerprint"
	"github.com/hashicorp/nomad/helper/fields"
	"github.com/hashicorp/nomad/nomad/structs"
	"github.com/mitchellh/mapstructure"

	dstructs "github.com/hashicorp/nomad/client/driver/structs"
	cstructs "github.com/hashicorp/nomad/client/structs"
	lxc "gopkg.in/lxc/go-lxc.v2"
)

type LxcDriver struct {
	DriverContext
	fingerprint.StaticFingerprinter
}

type LxcDriverConfig struct {
	Template             string
	Distro               string
	Release              string
	Arch                 string
	ImageVariant         string   "mapstructure:`image_variant`"
	ImageServer          string   "mapstructure:`image_server`"
	GPGKeyID             string   "mapstructure:`gpg_key_id`"
	GPGKeyServer         string   "mapstructure:`gpg_key_server`"
	DisableGPGValidation bool     "mapstructure:`disable_gpg`"
	FlushCache           bool     "mapstructure:`flush_cache`"
	ForceCache           bool     "mapstructure:`force_cache`"
	TemplateArgs         []string "mapstructure:`template_args`"
}

func NewLxcDriver(ctx *DriverContext) Driver {
	return &LxcDriver{DriverContext: *ctx}
}

func (d *LxcDriver) Validate(config map[string]interface{}) error {
	fd := &fields.FieldData{
		Raw: config,
		Schema: map[string]*fields.FieldSchema{
			"template": &fields.FieldSchema{
				Type:     fields.TypeString,
				Required: true,
			},
			"distro": &fields.FieldSchema{
				Type:     fields.TypeString,
				Required: false,
			},
			"release": &fields.FieldSchema{
				Type:     fields.TypeString,
				Required: false,
			},
			"arch": &fields.FieldSchema{
				Type:     fields.TypeString,
				Required: false,
			},
			"image_variant": &fields.FieldSchema{
				Type:     fields.TypeString,
				Required: false,
			},
			"image_server": &fields.FieldSchema{
				Type:     fields.TypeString,
				Required: false,
			},
			"gpg_key_id": &fields.FieldSchema{
				Type:     fields.TypeString,
				Required: false,
			},
			"gpg_key_server": &fields.FieldSchema{
				Type:     fields.TypeString,
				Required: false,
			},
			"disable_gpg": &fields.FieldSchema{
				Type:     fields.TypeString,
				Required: false,
			},
			"flush_cache": &fields.FieldSchema{
				Type:     fields.TypeString,
				Required: false,
			},
			"force_cache": &fields.FieldSchema{
				Type:     fields.TypeString,
				Required: false,
			},
			"template_args": &fields.FieldSchema{
				Type:     fields.TypeArray,
				Required: false,
			},
		},
	}

	if err := fd.Validate(); err != nil {
		return err
	}
	return nil
}

func (d *LxcDriver) Fingerprint(cfg *config.Config, node *structs.Node) (bool, error) {
	version := lxc.Version()
	if version == "" {
		return false, nil
	}
	node.Attributes["driver.lxc.version"] = version
	node.Attributes["driver.lxc"] = "1"
	return true, nil
}

func (d *LxcDriver) Start(ctx *ExecContext, task *structs.Task) (DriverHandle, error) {
	var driverConfig LxcDriverConfig
	if err := mapstructure.WeakDecode(task.Config, &driverConfig); err != nil {
		return nil, err
	}
	lxcPath := lxc.DefaultConfigPath()
	if path := d.config.Read("lxc.path"); path != "" {
		lxcPath = path
	}

	containerName := fmt.Sprintf("%s-%s", task.Name, ctx.AllocID)
	c, err := lxc.NewContainer(containerName, lxcPath)
	if err != nil {
		return nil, fmt.Errorf("unable to create container: %v", err)
	}
	c.SetVerbosity(lxc.Verbose)

	options := lxc.TemplateOptions{
		Template:             driverConfig.Template,
		Distro:               driverConfig.Distro,
		Release:              driverConfig.Release,
		Arch:                 driverConfig.Arch,
		FlushCache:           driverConfig.FlushCache,
		DisableGPGValidation: driverConfig.DisableGPGValidation,
	}

	if err := c.Create(options); err != nil {
		return nil, fmt.Errorf("unable to create container: %v", err)
	}

	handle := lxcDriverHandle{
		container:   c,
		logger:      d.logger,
		killTimeout: GetKillTimeout(task.KillTimeout, d.DriverContext.config.MaxKillTimeout),
	}

	return &handle, nil
}

func (d *LxcDriver) Open(ctx *ExecContext, handleID string) (DriverHandle, error) {
	return nil, nil
}

type lxcDriverHandle struct {
	container   *lxc.Container
	logger      *log.Logger
	killTimeout time.Duration
	waitCh      chan *dstructs.WaitResult
}

func (h *lxcDriverHandle) ID() string {
	return h.container.Name()
}

func (h *lxcDriverHandle) WaitCh() chan *dstructs.WaitResult {
	return h.waitCh
}

func (h *lxcDriverHandle) Update(task *structs.Task) error {
	h.killTimeout = GetKillTimeout(task.KillTimeout, h.killTimeout)
	return nil
}

func (h *lxcDriverHandle) Kill() error {
	return nil
}

func (h *lxcDriverHandle) Stats() (*cstructs.TaskResourceUsage, error) {
	return nil, nil
}
