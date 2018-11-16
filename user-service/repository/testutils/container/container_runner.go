// Package container simplifies running a container for testing purposes.
// TODO: tidy up options and add more error checking.
package container

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/docker/go-connections/nat"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"

	"github.com/docker/docker/client"
)

type Port string

type PortMapping map[Port]Port

func (pm PortMapping) With(host, container Port) PortMapping {
	pm[host] = container
	return pm
}

type Options struct {
	Image       string
	PortMapping PortMapping
	Writer      io.Writer
	Commands    []string
}

type Option func(*Options) error

func NewManager(opts ...Option) (m *Manager, err error) {
	m = &Manager{
		opts: &Options{
			Image:       "alpine",
			Writer:      os.Stdout,
			PortMapping: make(map[Port]Port),
			Commands:    []string{},
		},
	}
	for _, opt := range opts {
		if err = opt(m.opts); err != nil {
			return
		}
	}
	m.client, err = client.NewEnvClient()
	return
}

type Manager struct {
	opts        *Options
	client      client.APIClient
	containerID string
}

func (m *Manager) Options() Options {
	return *m.opts
}

func (m *Manager) pullImage(ctx context.Context) (err error) {
	reader, err := m.client.ImagePull(ctx, m.Options().Image, types.ImagePullOptions{})
	if err != nil {
		return
	}
	if _, err = io.Copy(m.Options().Writer, reader); err != nil {
		return
	}
	return
}

func (m *Manager) createContainer(ctx context.Context) (id string, err error) {
	hostPorts, containerPorts := m.ports()
	resp, err := m.client.ContainerCreate(ctx,
		&container.Config{Cmd: m.Options().Commands, Image: m.Options().Image, Tty: true, ExposedPorts: containerPorts},
		&container.HostConfig{PortBindings: hostPorts},
		nil, "")
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (m *Manager) ports() (nat.PortMap, nat.PortSet) {
	mp := make(map[nat.Port][]nat.PortBinding)
	ps := make(map[nat.Port]struct{})

	for pc, ph := range m.Options().PortMapping {
		mp[nat.Port(ph)] = []nat.PortBinding{
			{HostIP: "0.0.0.0", HostPort: string(ph)},
		}
		ps[nat.Port(pc)] = struct{}{}
	}

	return mp, ps
}

func (m *Manager) Start(ctx context.Context) (err error) {
	if err = m.pullImage(ctx); err != nil {
		return
	}
	m.containerID, err = m.createContainer(ctx)
	if err != nil {
		return
	}
	if err = m.client.ContainerStart(ctx, m.containerID, types.ContainerStartOptions{}); err != nil {
		return
	}
	reader, err := m.client.ContainerLogs(ctx, m.containerID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		return
	}
	if _, err = io.Copy(m.Options().Writer, reader); err != nil {
		return
	}
	return
}

func (m *Manager) Stop(ctx context.Context) (err error) {
	timeout := time.Second * 10
	if err = m.client.ContainerStop(ctx, m.containerID, &timeout); err != nil {
		return
	}
	return
}

func WithPortMapping(mapping PortMapping) Option {
	return func(o *Options) (err error) {
		o.PortMapping = mapping
		return
	}
}

func WithImage(name string) Option {
	return func(o *Options) (err error) {
		o.Image = name
		return
	}
}

func WithWriter(w io.Writer) Option {
	return func(o *Options) (err error) {
		o.Writer = w
		return
	}
}
