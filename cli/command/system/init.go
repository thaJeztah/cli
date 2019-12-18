package system

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/project"
	"github.com/spf13/cobra"
)

const dockerDir = ".docker"

type initOptions struct {
	template string
}

// NewInitCommand creates a new cobra.Command for `docker init`
func NewInitCommand(dockerCli command.Cli) *cobra.Command {
	var opts initOptions

	cmd := &cobra.Command{
		Use:   "init [OPTIONS]",
		Short: "Initialize a docker project",
		Args:  cli.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.template == "" {
				opts.template = project.SelectTemplate()
			}
			return runInit(cmd, dockerCli, &opts)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.template, "template", "t", "", "Template to use for initializing docker")

	return cmd
}

func runInit(cmd *cobra.Command, dockerCli command.Cli, opts *initOptions) error {
	if opts.template != "" {
		_ = os.Mkdir(dockerDir, 0777)
		err := CopyDirectory(opts.template, dockerDir)
		if err != nil {
			panic(err)
		}
	}
	return nil
}

func CopyDirectory(scrDir, dest string) error {
	entries, err := ioutil.ReadDir(scrDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(scrDir, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.Name() == ".DS_Store" {
			continue
		}

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		stat, ok := fileInfo.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", sourcePath)
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := CreateIfNotExists(destPath, 0755); err != nil {
				return err
			}
			if err := CopyDirectory(sourcePath, destPath); err != nil {
				return err
			}
		case os.ModeSymlink:
			if err := CopySymLink(sourcePath, destPath); err != nil {
				return err
			}
		default:
			if err := Copy(sourcePath, destPath); err != nil {
				return err
			}
		}

		if err := os.Lchown(destPath, int(stat.Uid), int(stat.Gid)); err != nil {
			return err
		}

		isSymlink := entry.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err := os.Chmod(destPath, entry.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

func Copy(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	defer out.Close()
	if err != nil {
		return err
	}

	in, err := os.Open(srcFile)
	defer in.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

func CreateIfNotExists(dir string, perm os.FileMode) error {
	if Exists(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}

	return nil
}

func CopySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return err
	}
	return os.Symlink(link, dest)
}
