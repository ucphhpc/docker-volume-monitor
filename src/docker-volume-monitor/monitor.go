package main

import (
	"flag"
	"time"

	containertypes "github.com/docker/docker/api/types/container"
	volumetypes "github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

func currentTime() string {
	return time.Now().Format(time.RFC3339)
}


func getVolumes(ctx context.Context, client *client.Client) (volumetypes.ListResponse, error) {
	volumes, err := client.VolumeList(ctx, volumetypes.ListOptions{})
	if err != nil {
		return volumetypes.ListResponse{}, err
	}
	return volumes, nil
}

func volumeInUse(ctx context.Context, client *client.Client, volume *volumetypes.Volume, debug bool) bool {
	containers, err := client.ContainerList(ctx, containertypes.ListOptions{All: true})
	if err != nil {
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

func removeVolumes(ctx context.Context, client *client.Client, volumes []*volumetypes.Volume, debug bool) error {
	for _, v := range volumes {
		if !volumeInUse(ctx, client, v, debug) {
			if debug {
				log.Debugf("%s - Removing Volume %s", time.Now().Format(time.RFC3339), v.Name)
			}
			if err := client.VolumeRemove(ctx, v.Name, true); err != nil {
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
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	if debug {
		log.SetLevel(log.DebugLevel)
	}

	for {
		log.Infof("%s - Checking for volumes", currentTime())
		volumes, err := getVolumes(ctx, cli)
		if err != nil {
			panic(err)
		}

		if debug {
			log.Debugf("%s - Found %d volumes", currentTime(), len(volumes.Volumes))
		}
		if pruneUnused {
			removeVolumes(ctx, cli, volumes.Volumes, debug)
		}
		log.Infof("%s - Finished checking for volumes", currentTime())
		time.Sleep(time.Duration(interval) * time.Minute)
	}
}
