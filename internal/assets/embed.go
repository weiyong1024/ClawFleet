package assets

import "embed"

// DockerFS contains the Docker build context (Dockerfile, supervisord.conf, entrypoint.sh).
//
//go:embed docker
var DockerFS embed.FS
