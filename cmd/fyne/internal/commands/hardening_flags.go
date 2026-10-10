package commands

import (
	"os"
	"os/exec"
	"regexp"
	"strings"

	"golang.org/x/mod/semver"

	"fyne.io/tools/cmd/fyne/internal/util"
)

const (
	hardeningCFLAGS        = "-D_FORTIFY_SOURCE=3 -fcf-protection -fstack-protector-strong"
	hardeningLDFLAGSLinux  = "-Wl,-z,relro,-z,now -Wl,--as-needed"
	hardeningLDFLAGSDarwin = "-Wl,-dead_strip_dylibs"
)

type hardeningFlags struct {
	os, arch, cc, minVer, maxVer, cflags string
}

// specific flags go first, generic flags last
var hardeningFlagsTable = []hardeningFlags{
	//revive:disable:add-constant
	{"ubuntu", "amd64", "gcc", "*", "11.4.0", "-fcf-protection -fstack-protector-strong"}, // Ubuntu 22.04/gcc 11.4.0 fails with _FORTIFY_SOURCE redefined error
	{"windows", "*", "gcc", "*", "*", "-D_FORTIFY_SOURCE=3 -fstack-protector-strong"},     // mingw doesn't support -fcf-protection -- XXX: double check for better conditions
	{"*", "arm64", "*", "*", "*", "-D_FORTIFY_SOURCE=3 -fstack-protector-strong"},         // -fcf-protection unsupported on arm64
	//revive:enable:add-constant
}

func ccProg() string {
	cc, ok := os.LookupEnv("CC")
	if !ok {
		return "cc"
	}
	return cc
}

func ccVersion() string {
	cc := ccProg()
	cmd := exec.Command(cc, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}

	return string(out)
}

const ccTestFlagsCode = `
int main(int argc, char **argv) {
	return 0;
}
`

func ccVersionAndDefaultFlags() (string, error) {
	f, err := os.CreateTemp("", "fyne-check-cc-*.c")
	if err != nil {
		return "", err
	}
	inFile := f.Name()
	outFile := strings.TrimSuffix(inFile, ".c")
	defer func() {
		_ = os.Remove(inFile)
		_ = os.Remove(outFile)
	}()

	if _, err := f.WriteString(ccTestFlagsCode); err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}

	cmd := exec.Command(ccProg(), "-Q", "-v", "-o", outFile, inFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func gccDefaultFlags(s string) []string {
	re := regexp.MustCompile(`(?ms)^options enabled:((?:\s+\S+\n?)*)`)
	m := re.FindStringSubmatch(s)
	if len(m) < 2 {
		return nil
	}
	return strings.Fields(m[1])
}

func hardeningCFlagsLookup(out, goos, arch string) string {
	info, err := util.DetectCompiler(out, goos)
	if err != nil {
		return ""
	}

	var defs []string
	if info.Name == "gcc" {
		defs = gccDefaultFlags(out)
	}

	for _, e := range hardeningFlagsTable {
		//revive:disable:add-constant
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
		return dedupeFlags(e.cflags, defs)
		//revive:enable:add-constant
	}
	return dedupeFlags(hardeningCFLAGS, defs)
}

func dedupeFlags(s string, defs []string) string {
	if len(defs) == 0 {
		return s
	}

	seen := make(map[string]struct{})
	for _, flag := range defs {
		seen[flag] = struct{}{}
	}

	r := []string{}
	for _, flag := range strings.Fields(s) {
		if _, found := seen[flag]; found {
			continue
		}
		seen[flag] = struct{}{}
		r = append(r, flag)
	}
	return util.JoinSpace(r)
}
