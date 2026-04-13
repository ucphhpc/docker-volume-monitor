package main

import (
	"flag"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/volume"
	"github.com/moby/moby/client"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

func currentTime() string {
	return time.Now().Format(time.RFC3339)
}

func getVolumes(ctx context.Context, apiClient *client.Client) ([]volume.Volume, error) {
	volumes, err := apiClient.VolumeList(ctx, client.VolumeListOptions{})
	if err != nil {
		return nil, err
	}
	return volumes.Items, nil
}

func getContainers(ctx context.Context, apiClient *client.Client) ([]container.Summary, error) {
	containers, err := apiClient.ContainerList(ctx, client.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}
	return containers.Items, nil
}

func volumeInUse(ctx context.Context, apiClient *client.Client, volume volume.Volume, debug bool) bool {
	containers, err := getContainers(ctx, apiClient)
	if err != nil || containers == nil {
		return false
	}
	inUse := false

	if debug {
		log.Debugf("%s - Checking whether volume: %s is in use", currentTime(), volume.Name)
	}


	for _, container := range containers {
		for _, mount := range container.Mounts {
			if debug {
				log.Debugf("%s - Checking if mount: %s uses volume: %s", currentTime(), mount.Name, volume.Name)
			}

			if mount.Name == volume.Name {
				inUse = true
			}
		}
	}

	return inUse
}

func removeVolumes(ctx context.Context, apiClient *client.Client, volumes []volume.Volume, debug bool) error {
	for _, v := range volumes {
		if !volumeInUse(ctx, apiClient, v, debug) {
			if debug {
				log.Debugf("%s - Removing Volume %s", time.Now().Format(time.RFC3339), v.Name)
			}
			if _, err := apiClient.VolumeRemove(ctx, v.Name, client.VolumeRemoveOptions{Force: true}); err != nil {
				log.Errorf("%s - Failed to remove Volume %s Reason %s", currentTime(), v.Name, err)
			} else {
				if debug {
					log.Debugf("%s - Removed Volume %s", currentTime(), v.Name)
				}
			}
		}
	}
	return nil
}

func run() {
	// Setup parameters
	var interval int
	var pruneUnused bool
	var debug bool

	flag.IntVar(&interval, "interval",
		10, "How often the monitor should check for unused volumes")
	flag.BoolVar(&pruneUnused, "prune-unused", true,
		"Whether the monitor should prune unused volumes")
	flag.BoolVar(&debug, "debug", false,
		"Set the debug flag to run the monitor in debug mode")
	flag.Parse()

	ctx := context.Background()
	cli, err := client.New(client.FromEnv)
	if err != nil {
		panic(err)
	}

	if debug {
		log.SetLevel(log.DebugLevel)
		log.Debugf("%s - Debug logging enabled: %v", currentTime())
	}

	for {
		log.Infof("%s - Checking for volumes", currentTime())
		volumes, err := getVolumes(ctx, cli)
		if err != nil {
			panic(err)
		}

		if debug {
			log.Debugf("%s - Found %d volumes", currentTime(), len(volumes))
		}
		if pruneUnused {
			removeVolumes(ctx, cli, volumes, debug)
		}
		log.Infof("%s - Finished checking for volumes", currentTime())
		time.Sleep(time.Duration(interval) * time.Minute)
	}
}
