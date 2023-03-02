package main

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func main() {
	// create client
	c, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	// create
	res, err := c.ContainerCreate(ctx,
		&container.Config{
			Image: "mongo",
			ExposedPorts: nat.PortSet{
				"27017/tcp": {},
			},
		},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				"27017/tcp": []nat.PortBinding{
					{
						HostIP:   "127.0.0.1",
						HostPort: "0",
					},
				},
			},
		},
		nil, "")
	if err != nil {
		panic(err)
	}
	// start
	err = c.ContainerStart(ctx, res.ID, types.ContainerStartOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Println("container started")
	time.Sleep(5 * time.Second)
	// listen
	inspRes, err := c.ContainerInspect(ctx, res.ID)
	if err != nil {
		panic(err)
	}
	fmt.Printf("listening at %+v\n",
		inspRes.NetworkSettings.Ports["27017/tcp"][0])
	// remove
	fmt.Println("killing container")
	err = c.ContainerRemove(ctx, res.ID, types.ContainerRemoveOptions{
		Force: true,
	})
	if err != nil {
		panic(err)
	}
}
