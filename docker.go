package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	d "github.com/fsouza/go-dockerclient"
)

// ContainerWatch checks for new containers and if they exist, add the sites and it's endpoints
func ContainerWatch(containerized, healthchecks bool, healthCheckURL string, myPort int) {
	myPort64 := int64(myPort)
	address := "localhost"
	if containerized {
		address = containerizedIP()
	}

	client, err := d.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		log.Error(err.Error())
		return
	}

	for {
		if containers, err := client.ListContainers(d.ListContainersOptions{All: true}); err == nil {
			for _, container := range containers {
				name := container.Names[0][1:]
				for _, port := range container.Ports {
					if port.PublicPort != myPort64 && port.Type == "tcp" {
						log.WithFields(log.Fields{
							"Public":  port.PublicPort,
							"Private": port.PrivatePort,
						}).Debug(name)
						AddSite(name, fmt.Sprintf("http://%s:%d", address, port.PublicPort), healthchecks, healthCheckURL)
					}
				}
			}
		} else {
			log.Error("Unable to connect to docker")
			log.Error(err.Error())
		}

		// Every 5 seconds, check for new containers
		<-time.After(5 * time.Second)
	}
}

// containerizedIP returns a string with the ip address of the docker host. Localhost else ...
func containerizedIP() string {
	// Do we need to start our docker service?
	cmd := exec.Command("bash", "-c", "/sbin/ip route|awk '/default/ { print $3 }'")
	if output, err := cmd.Output(); err == nil {
		log.WithFields(log.Fields{
			"IP": strings.TrimSpace(string(output)),
		}).Info("Auto detecting docker host IP Address")
		return fmt.Sprintf("%s", strings.TrimSpace(string(output)))
	}
	return "localhost"
}
