# TODO

## Ubuntu Image Option

Support an optional `--image-flavor` flag on `clawfleet create` to choose between
the default Alpine-based image and an Ubuntu-based image. Ubuntu may offer better
compatibility for certain OpenClaw plugins and browser automation workloads.

## OpenClaw Version Pinning

The Dockerfile currently installs `openclaw@latest`. Consider pinning to a specific
version (e.g. `openclaw@2026.3.7`) to ensure reproducible builds and avoid breaking
changes when OpenClaw releases a new version.
