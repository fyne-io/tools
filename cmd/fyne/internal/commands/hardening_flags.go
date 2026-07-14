package commands

import (
	"os"
	"os/exec"

	"golang.org/x/mod/semver"

	intutil "fyne.io/tools/cmd/fyne/internal/util"
)

const (
	hardeningCFLAGS        = "-D_FORTIFY_SOURCE=3 -fcf-protection -fstack-protector-strong"
	hardeningLDFLAGSLinux  = "-Wl,-z,relro,-z,now -Wl,--as-needed"
	hardeningLDFLAGSDarwin = "-Wl,-dead_strip_dylibs"
)

type hardeningFlags struct {
	os, arch, cc, minVer, maxVer, cflags string
}

var hardeningFlagsTable = []hardeningFlags{
	{"ubuntu", "amd64", "gcc", "*", "11.4.0", "-fcf-protection -fstack-protector-strong"},  // Ubuntu 22.04/gcc 11.4.0 fails with _FORTIFY_SOURCE redefined error
	{"windows", "*", "gcc", "*", "*", "-D_FORTIFY_SOURCE=3 -fstack-protector-strong"},      // mingw doesn't support -fcf-protection -- XXX: double check for better conditions
	{"darwin", "arm64", "clang", "*", "*", "-D_FORTIFY_SOURCE=3 -fstack-protector-strong"}, // -fcf-protection unsupported on this target
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

func hardeningCFlagsLookup(out, goos, arch string) string {
	info, err := intutil.DetectCompiler(out, goos)
	if err != nil {
		return ""
	}
	for _, e := range hardeningFlagsTable {
		if e.cc != "*" && e.cc != info.Name {
			continue
		}
		if e.os != "*" && e.os != info.OS {
			continue
		}
		if e.arch != "*" && e.arch != arch {
			continue
		}
		if e.minVer != "*" && semver.Compare("v"+info.Version, "v"+e.minVer) < 0 {
			continue
		}
		if e.maxVer != "*" && semver.Compare("v"+info.Version, "v"+e.maxVer) > 0 {
			continue
		}
		return e.cflags
	}
	return hardeningCFLAGS
}
