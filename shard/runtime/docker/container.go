package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"prismarine/shard/runtime"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
)

func (i *Instance) Attach(ctx context.Context) error {
	if i.IsAttached() {
		return nil
	}

	opts := container.AttachOptions{
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Stream: true,
	}

	// Set the stream again with the container.
	if st, err := i.client.ContainerAttach(ctx, i.Cfg.Uuid, opts); err != nil {
		return errors.Wrap(err, "runtime/docker: error while attaching to container")
	} else {
		i.SetStream(&st)
	}

	go func() {
		// pollCtx, cancel := context.WithCancel(context.Background())
		// defer cancel()
		defer i.stream.Close()
		defer func() {
			i.SetState(runtime.ProcessOfflineState)
			i.SetStream(nil)
		}()

		//go func() {
		//	if err := i.
		//}()
	}()

	return nil
}

func (i *Instance) Create() error {
	log.
		With("runtime", "docker").
		With("Instance", i.Id).
		Debug("Creating Instance")

	ctx := context.Background()

	if _, err := i.ContainerInspect(ctx); err == nil {
		return nil
	} else if !client.IsErrNotFound(err) {
		return errors.Wrap(err, "runtime/docker: failed to inspect")
	}

	if err := i.ensureImageExists(i.Cfg.Container.Image); err != nil {
		return errors.WithStack(err)
	}

	conf := &container.Config{
		Hostname:        "",
		Domainname:      "",
		User:            "",
		AttachStdin:     true,
		AttachStdout:    true,
		AttachStderr:    true,
		ExposedPorts:    nil,
		Tty:             true,
		OpenStdin:       true,
		StdinOnce:       false,
		Env:             nil,
		Cmd:             nil,
		Healthcheck:     nil,
		ArgsEscaped:     false,
		Image:           strings.TrimPrefix(i.Cfg.Container.Image, "~"),
		Volumes:         nil,
		WorkingDir:      "",
		Entrypoint:      nil,
		NetworkDisabled: false,
		OnBuild:         nil,
		Labels:          nil,
		StopSignal:      "",
		StopTimeout:     nil,
		Shell:           nil,
	}

	// Set the user running the container properly depending on what mode we are operating in.

	// Network config

	hostConf := &container.HostConfig{
		Binds:           nil,
		ContainerIDFile: "",
		LogConfig:       container.LogConfig{},
		NetworkMode:     "",
		PortBindings:    nil,
		RestartPolicy:   container.RestartPolicy{},
		AutoRemove:      false,
		VolumeDriver:    "",
		VolumesFrom:     nil,
		ConsoleSize:     [2]uint{},
		Annotations:     nil,
		CapAdd:          nil,
		CapDrop:         nil,
		CgroupnsMode:    "",
		DNS:             nil,
		DNSOptions:      nil,
		DNSSearch:       nil,
		ExtraHosts:      nil,
		GroupAdd:        nil,
		IpcMode:         "",
		Cgroup:          "",
		Links:           nil,
		OomScoreAdj:     0,
		PidMode:         "",
		Privileged:      false,
		PublishAllPorts: false,
		ReadonlyRootfs:  false,
		SecurityOpt:     nil,
		StorageOpt:      nil,
		Tmpfs:           nil,
		UTSMode:         "",
		UsernsMode:      "",
		ShmSize:         0,
		Sysctls:         nil,
		Runtime:         "",
		Isolation:       "",
		Resources:       container.Resources{},
		Mounts:          nil,
		MaskedPaths:     nil,
		ReadonlyPaths:   nil,
		Init:            nil,
	}

	if _, err := i.client.ContainerCreate(ctx, conf, hostConf, nil, nil, i.Cfg.Uuid); err != nil {
		return errors.Wrap(err, "runtime/docker: failed to create container")
	}

	return nil
}

// Pulls the image from Docker. If there is an error while pulling the image
// from the source but the image already exists locally, we will report that
// error to the logger but continue with the process.
//
// The reasoning behind this is that Quay has had some serious outages as of
// late, and we don't need to block all the servers from booting just because
// of that. I'd imagine in a lot of cases an outage shouldn't affect users too
// badly. It'll at least keep existing servers working correctly if anything.
func (i *Instance) ensureImageExists(image string) error {
	i.Events().Publish(runtime.DockerImagePullStarted, "")
	defer i.Events().Publish(runtime.DockerImagePullCompleted, "")

	// Images prefixed with a ~ are local images we do not need to pull
	if strings.HasPrefix(image, "~") {
		return nil
	}

	// Give up to 15 minutes to pull the image... probably a better way to do this
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*15)
	defer cancel()

	out, err := i.client.ImagePull(ctx, image, types.ImagePullOptions{All: false})
	if err != nil {
		images, ierr := i.client.ImageList(ctx, types.ImageListOptions{})
		if ierr != nil {
			return errors.Wrap(ierr, "runtime/docker: failed to list images")
		}

		for _, img := range images {
			for _, t := range img.RepoTags {
				if t != image {
					continue
				}

				log.
					With("image", image).
					With("container_id", i.Id).
					Warn("unable to pull requested image from remote server, however image exists locally")

				return nil
			}
		}

		return errors.Wrap(err, "runtime/docker: failed to pull image")
	}

	defer out.Close()

	log.With("image", image).Debug("Pulling docker image... this may take a while")

	// Not sure if this is the best way to do this... blocks execution until the image is
	// done being pulled
	scanner := bufio.NewScanner(out)

	for scanner.Scan() {
		b := scanner.Bytes()
		m := make(map[string]interface{})
		err := json.Unmarshal(b, &m)
		if err != nil {
			return err
		}
		status := m["status"]
		progress := m["progress"]

		i.Events().Publish(runtime.DockerImagePullStatus, progress)
		log.
			With("runtime", "docker").
			With("image", image).
			With("status", fmt.Sprintf("%s", status)).
			With("progress", fmt.Sprintf("%s", progress)).
			Debug("Pulling Image")

	}

	if err := scanner.Err(); err != nil {
		return err
	}

	log.With("image", image).Debug("completed docker image pull")

	return nil
}
