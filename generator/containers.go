package generator

import (
	"github.com/docker/docker/api/types"
	"github.com/lucaslorentz/caddy-docker-proxy/v2/caddyfile"
	"github.com/lucaslorentz/caddy-docker-proxy/v2/docker"
	"go.uber.org/zap"
)

func (g *CaddyfileGenerator) getContainerCaddyfile(dockerClient *docker.Client, container *types.Container, logger *zap.Logger) (*caddyfile.Container, error) {
	caddyLabels := g.filterLabels(container.Labels)

	return labelsToCaddyfile(caddyLabels, container, func() ([]string, error) {
		return g.getContainerIPAddresses(dockerClient, container, logger, true)
	})
}

func (g *CaddyfileGenerator) getContainerIPAddresses(dockerClient *docker.Client, container *types.Container, logger *zap.Logger, onlyIngressIps bool) ([]string, error) {
	ips := []string{}

	ingressNetworkFromLabel, overrideNetwork := container.Labels[IngressNetworkLabel]

	for networkName, network := range container.NetworkSettings.Networks {
		include := false

		if !onlyIngressIps {
			include = true
		} else if overrideNetwork {
			include = networkName == ingressNetworkFromLabel
		} else {
			include = g.ingressNetworks[network.NetworkID]
		}

		if include {
			ips = append(ips, network.IPAddress)
		} else if networkName == "host" {
			ips = append(ips, g.dockerClientGatewayIP[*dockerClient])
		}
	}

	if len(ips) == 0 {
		logger.Warn("Container is not in same network as caddy", zap.String("container", container.ID), zap.String("container id", container.ID))

	}

	return ips, nil
}
