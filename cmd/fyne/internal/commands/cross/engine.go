package cross

import (
	"context"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	mobyclient "github.com/moby/moby/client"
)

const (
	// containerWorkDir is the working directory inside the container
	containerWorkDir = "/app"
	// containerCacheMount is the path where the host cache directory is mounted in the container
	containerCacheMount = "/go"
	// containerGoCachePath is the Go build cache path inside the container
	containerGoCachePath = "/go/go-build"
	// containerZigCachePath is the Zig compiler cache path inside the container
	containerZigCachePath = "/go/zig-cache"
	// stripFlags are linker flags to remove debug symbols and DWARF info
	stripFlags = "-s -w"
)

// Engine manages container-based cross-compilation using Docker/Moby.
// It handles container configuration, image selection, volume mounts, and command execution.
type Engine struct {
	client  *mobyclient.Client
	env     map[string]string
	image   string
	mounts  map[string]string
	goarch  string
	goos    string
	verbose bool
}

// NewEngine creates a new Engine configured for the specified target OS and architecture.
// It initializes the Docker client, selects the appropriate container image, and sets up
// the cross-compilation environment variables. Returns an error if the target is unsupported
// or if the Docker client cannot be created.
func NewEngine(goos, goarch string) (*Engine, error) {
	image, err := TargetImage(goos)
	if err != nil {
		return nil, err
	}

	env, err := TargetEnv(goos, goarch)
	if err != nil {
		return nil, err
	}

	client, err := mobyclient.New(mobyclient.FromEnv)
	if err != nil {
		return nil, err
	}

	return &Engine{
		client: client,
		env:    env,
		image:  image,
		mounts: make(map[string]string),
		goarch: goarch,
		goos:   goos,
	}, nil
}

// Close closes the underlying Docker client connection.
func (e *Engine) Close() error {
	return e.client.Close()
}

// Debug returns a formatted string representation of the engine configuration
// including target OS/arch, image, mounts, and environment variables.
// Useful for verbose output and debugging container setup.
func (e *Engine) Debug() string {
	var buf strings.Builder

	buf.WriteString("Target:      " + e.goos + "/" + e.goarch + "\n")
	buf.WriteString("Image:       " + e.image + "\n")

	if len(e.mounts) > 0 {
		buf.WriteString("Mounts:\n")
		for src, tgt := range e.mounts {
			buf.WriteString(fmt.Sprintf("  %s -> %s\n", src, tgt))
		}
	}

	if len(e.env) > 0 {
		buf.WriteString("Environment:\n")
		// Sort env vars for consistent output
		var envKeys []string
		for k := range e.env {
			envKeys = append(envKeys, k)
		}
		sort.Strings(envKeys)
		for _, k := range envKeys {
			buf.WriteString(fmt.Sprintf("  %s=%s\n", k, e.env[k]))
		}
	}

	return strings.TrimSuffix(buf.String(), "\n")
}

// Exec executes the given command in a new container with the configured environment and mounts.
// The container is automatically removed after execution. Returns an error if the command fails
// or exits with a non-zero status code.
func (e *Engine) Exec(ctx context.Context, cmd Command) error {
	env := make([]string, 0, len(e.env))
	for k, v := range e.env {
		env = append(env, k+"="+v)
	}

	mounts := make([]mount.Mount, 0, len(e.mounts))
	for src, tgt := range e.mounts {
		if !filepath.IsAbs(src) {
			return fmt.Errorf("mount source %q must be an absolute path", src)
		}
		if _, err := os.Stat(src); os.IsNotExist(err) {
			return fmt.Errorf("cannot mount src %q path with target %q. src does not exists", src, tgt)
		}
		mounts = append(mounts, mount.Mount{Type: mount.TypeBind, Source: src, Target: tgt})
	}

	// Get current user UID/GID for running container
	uid := fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid())

	opts := mobyclient.ContainerCreateOptions{
		Config: &container.Config{
			Image:        e.Image(),
			Cmd:          cmd.Command(),
			Env:          env,
			WorkingDir:   containerWorkDir,
			User:         uid,
			AttachStdout: true,
			AttachStderr: true,
		},
		HostConfig: &container.HostConfig{
			AutoRemove: true,
			Mounts:     mounts,
		},
	}

	resp, err := e.client.ContainerCreate(ctx, opts)
	if err != nil {
		return fmt.Errorf("create container: %w", err)
	}

	attach, err := e.client.ContainerAttach(ctx, resp.ID, mobyclient.ContainerAttachOptions{
		Stream: true, Stdout: true, Stderr: true,
	})
	if err != nil {
		return fmt.Errorf("attach: %w", err)
	}
	defer attach.Close()

	_, err = e.client.ContainerStart(ctx, resp.ID, mobyclient.ContainerStartOptions{})
	if err != nil {
		return fmt.Errorf("start container: %w", err)
	}

	// Stream output synchronously (blocks until container exits and closes the stream)
	stdcopy.StdCopy(os.Stdout, os.Stderr, attach.Reader)

	// Get final status via inspect
	info, err := e.client.ContainerInspect(ctx, resp.ID, mobyclient.ContainerInspectOptions{})
	if err != nil {
		return fmt.Errorf("inspect: %w", err)
	}

	state := info.Container.State
	if state.ExitCode != 0 {
		return fmt.Errorf("ext code: %d - error: %s", state.ExitCode, state.Error)
	}

	return nil
}

// Image returns the container image name that will be used for builds.
func (e *Engine) Image() string {
	return e.image
}

// PullImage pulls the latest version of the container image from the registry.
func (e *Engine) PullImage(ctx context.Context) error {
	out, err := e.client.ImagePull(ctx, e.Image(), mobyclient.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(io.Discard, out)
	return err
}

// WithCache enables caching with the specified cache directory.
// Sets up mount and environment variables for Go and Zig caches.
func (e *Engine) WithCache(cacheDir string) {
	e.WithMount(cacheDir, containerCacheMount)
	e.env["GOCACHE"] = containerGoCachePath
	e.env["ZIG_GLOBAL_CACHE_DIR"] = containerZigCachePath
}

// WithEnv merges additional environment variables into the engine's environment.
func (e *Engine) WithEnv(env map[string]string) {
	maps.Copy(e.env, env)
}

// WithImage sets a custom container image, overriding the default image for the target OS.
// If the provided image string is empty, the default image is kept.
func (e *Engine) WithImage(image string) {
	if image != "" {
		e.image = image
	}
}

// WithMount adds a bind mount from the host path (src) to the container path (target).
func (e *Engine) WithMount(src, target string) {
	e.mounts[src] = target
}

// WithSourceMount mounts the specified host directory as the working directory in the container.
func (e *Engine) WithSourceMount(hostPath string) {
	e.WithMount(hostPath, containerWorkDir)
}

// WithStripDebug enables stripping of debug symbols and DWARF info.
// For release builds, this passes strip flags directly to the C linker
// since Go's -ldflags="-s -w" only affects the internal Go linker,
// not the external linker used in CGO cross-compilation with zig cc.
func (e *Engine) WithStripDebug() {
	e.appendCGOLDFlags(stripFlags)
}

// WithVerbose enables or disables verbose output during container operations.
func (e *Engine) WithVerbose(verbose bool) {
	e.verbose = verbose
}

// appendCGOLDFlags appends flags to the CGO_LDFLAGS environment variable.
func (e *Engine) appendCGOLDFlags(flags string) {
	if existing, ok := e.env["CGO_LDFLAGS"]; ok {
		e.env["CGO_LDFLAGS"] = existing + " " + flags
		return
	}
	e.env["CGO_LDFLAGS"] = flags
}
