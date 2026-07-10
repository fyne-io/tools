package commands

import (
	"os"
	"os/exec"

	"golang.org/x/mod/semver"

	intutil "fyne.io/tools/cmd/fyne/internal/util"
)

const (
	hardeningCFLAGS        = "-D_FORTIFY_SOURCE=3 -fstack-protector-strong"
	hardeningLDFLAGSLinux  = "-Wl,-z,relro,-z,now -Wl,--as-needed"
	hardeningLDFLAGSDarwin = "-Wl,-dead_strip_dylibs"
)

type hardeningFlags struct {
	os, arch, cc, minVer, maxVer, cflags string
}

var hardeningFlagsTable = []hardeningFlags{
	{"ubuntu", "amd64", "gcc", "0", "11.4.0", "-fstack-protector-strong"}, // https://github.com/fyne-io/tools/issues/137
}

func ccVersion() string {
	cc, ok := os.LookupEnv("CC")
	if !ok {
		cc = "cc"
	}

	cmd := exec.Command(cc, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}

	return string(out)
}

func hardeningFlagsLookup(out, goos, arch string) string {
	info, err := intutil.DetectCompiler(out, goos)
	if err != nil {
		return ""
	}
	for _, e := range hardeningFlagsTable {
		if e.os != info.OS {
			continue
		}
		if e.arch != arch {
			continue
		}
		if semver.Compare("v"+info.Version, "v"+e.minVer) < 0 {
			continue
		}
		if semver.Compare("v"+info.Version, "v"+e.maxVer) > 0 {
			continue
		}
		return e.cflags
	}
	return hardeningCFLAGS
}
