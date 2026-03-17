package container

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"strings"

	docker "github.com/fsouza/go-dockerclient"

	"github.com/weiyong1024/clawfleet/internal/assets"
)

func ImageExists(cli *docker.Client, imageRef string) (bool, error) {
	images, err := cli.ListImages(docker.ListImagesOptions{All: false})
	if err != nil {
		return false, fmt.Errorf("listing images: %w", err)
	}
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == imageRef {
				return true, nil
			}
		}
	}
	return false, nil
}

func Build(cli *docker.Client, imageRef string, out io.Writer) error {
	buildCtx, err := createBuildContext()
	if err != nil {
		return fmt.Errorf("creating build context: %w", err)
	}

	err = cli.BuildImage(docker.BuildImageOptions{
		Name:           imageRef,
		InputStream:    buildCtx,
		OutputStream:   out,
		RmTmpContainer: true,
	})
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}
	return nil
}

// TagImage adds an additional tag to an already-built image.
func TagImage(cli *docker.Client, existingRef, repo, tag string) error {
	return cli.TagImage(existingRef, docker.TagImageOptions{
		Repo:  repo,
		Tag:   tag,
		Force: true,
	})
}

// PullImage pulls an image from a remote registry.
func PullImage(cli *docker.Client, repo, tag string, out io.Writer) error {
	return cli.PullImage(docker.PullImageOptions{
		Repository:   repo,
		Tag:          tag,
		OutputStream: out,
	}, docker.AuthConfiguration{})
}

func createBuildContext() (io.Reader, error) {
	buf := &bytes.Buffer{}
	tw := tar.NewWriter(buf)

	err := fs.WalkDir(assets.DockerFS, "docker", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		content, err := assets.DockerFS.ReadFile(path)
		if err != nil {
			return err
		}
		name := strings.TrimPrefix(path, "docker/")
		mode := int64(0644)
		if name == "entrypoint.sh" {
			mode = 0755
		}
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: mode, Size: int64(len(content))}); err != nil {
			return err
		}
		_, err = tw.Write(content)
		return err
	})
	if err != nil {
		return nil, err
	}
	tw.Close()
	return buf, nil
}
