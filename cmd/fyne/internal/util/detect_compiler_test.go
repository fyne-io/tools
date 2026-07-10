package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DetectCompiler(t *testing.T) {
	tests := []struct {
		unameInfo string
		osName    string
		ccVersion string
		wantInfo  *CompilerInfo
		wantError error
	}{
		// darwin
		{
			"Darwin 24.6.0 arm",
			"darwin",
			"Apple clang version 17.0.0 (clang-1700.0.13.5)",
			&CompilerInfo{"clang", "17.0.0", "darwin"},
			nil,
		},
		{
			"Darwin 25.4.0 arm",
			"darwin",
			"Apple clang version 21.0.0 (clang-2100.1.1.101)",
			&CompilerInfo{"clang", "21.0.0", "darwin"},
			nil,
		},
		{
			"Darwin 25.5.0 arm",
			"darwin",
			"Apple clang version 21.0.0 (clang-2100.1.1.101)",
			&CompilerInfo{"clang", "21.0.0", "darwin"},
			nil,
		},
		// freebsd
		{
			"FreeBSD 14.2-RELEASE amd64",
			"freebsd",
			"FreeBSD clang version 18.1.6 (https://github.com/llvm/llvm-project.git llvmorg-18.1.6-0-g1118c2e05e67)",
			&CompilerInfo{"clang", "18.1.6", "freebsd"},
			nil,
		},
		{
			"FreeBSD 15.1-RELEASE-p1 amd64",
			"freebsd",
			"FreeBSD clang version 19.1.7 (https://github.com/llvm/llvm-project.git llvmorg-19.1.7-0-gcd708029e0b2)",
			&CompilerInfo{"clang", "19.1.7", "freebsd"},
			nil,
		},
		// linux
		{
			"Linux 6.12.0-160000.35-default x86_64",
			"linux",
			"cc (SUSE Linux) 15.3.0",
			&CompilerInfo{"gcc", "15.3.0", "suse"},
			nil,
		},
		{
			"Linux 6.12.86+deb13-arm64 unknown",
			"linux",
			"cc (Debian 14.2.0-19) 14.2.0",
			&CompilerInfo{"gcc", "14.2.0", "debian"},
			nil,
		},
		{
			"Linux 6.12.88+deb13-arm64 unknown",
			"linux",
			"cc (Debian 14.2.0-19) 14.2.0",
			&CompilerInfo{"gcc", "14.2.0", "debian"},
			nil,
		},
		{
			"Linux 6.12.93+rpt-rpi-v8 unknown",
			"linux",
			"cc (Debian 12.2.0-14+deb12u1) 12.2.0  (PI4)",
			&CompilerInfo{"gcc", "12.2.0", "debian"},
			nil,
		},
		{
			"Linux 6.18.35-gentoo-dist AMD Ryzen 7 7800X3D 8-Core Processor",
			"linux",
			"cc (Gentoo 15.3.0 p8) 15.3.0",
			&CompilerInfo{"gcc", "15.3.0", "gentoo"},
			nil,
		},
		{
			"Linux 7.0.3-arch1-2 unknown",
			"linux",
			"cc (GCC) 16.1.1 20260430",
			&CompilerInfo{"gcc", "16.1.1", "linux"},
			nil,
		},
		{
			"Linux 7.0.12-arch1-1 unknown",
			"linux",
			"cc (GCC) 16.1.1 20260430",
			&CompilerInfo{"gcc", "16.1.1", "linux"},
			nil,
		},
		{
			"Linux 7.1.2-arch3-1 unknown",
			"linux",
			"cc (GCC) 16.1.1 20260625",
			&CompilerInfo{"gcc", "16.1.1", "linux"},
			nil,
		},
		{
			"??? Ubuntu ....",
			"linux",
			"cc (Ubuntu 11.4.0-1ubuntu1~22.04.3) 11.4.0",
			&CompilerInfo{"gcc", "11.4.0", "ubuntu"},
			nil,
		},
		// windows
		{
			"Microsoft Windows 11 Pro 10.0.26200.0 AMD64",
			"windows",
			"clang version 22.1.8 (https://github.com/llvm/llvm-project.git ca7933e47d3a3451d81e72ac174dcb5aa28b59d1)",
			&CompilerInfo{"clang", "22.1.8", "windows"},
			nil,
		},
		{
			"MSYS_NT-10.0-26200 3.6.7-f2802c5f.x86_64 unknown",
			"windows",
			"cc.exe (Rev13, Built by MSYS2 project) 15.2.0",
			&CompilerInfo{"gcc", "15.2.0", "msys2"},
			nil,
		},
		// openbsd
		{
			"OpenBSD 7.3 aarch64",
			"openbsd",
			"OpenBSD clang version 13.0.0",
			&CompilerInfo{"clang", "13.0.0", "openbsd"},
			nil,
		},
		// errors
		{
			"unknown",
			"random",
			"clank version 1.0.0",
			nil,
			ErrNoCompilerFound,
		},
		{
			"unknown",
			"random",
			"tcc ( ,-a-b, -y -z ) 1.0.0",
			nil,
			ErrNoCompilerFound,
		},
		{
			"empty",
			"",
			"",
			nil,
			ErrNoCompilerFound,
		},
		{
			"invalid gcc version",
			"random",
			"cc (Haha) 1.2.3.4",
			nil,
			ErrInvalidVersion,
		},
		{
			"invalid clang version",
			"random",
			"Hehe clang version 1.2.3.4",
			nil,
			ErrInvalidVersion,
		},
	}

	for _, test := range tests {
		info, err := DetectCompiler(test.ccVersion, test.osName)
		if test.wantError == nil {
			assert.NoError(t, err)
			assert.Equal(t, test.wantInfo.Name, info.Name, "name: %s: %q", test.unameInfo, test.ccVersion)
			assert.Equal(t, test.wantInfo.Version, info.Version, "version: %s: %q", test.unameInfo, test.ccVersion)
			assert.Equal(t, test.wantInfo.OS, info.OS, "os: %s: %q", test.unameInfo, test.ccVersion)
		} else if assert.Error(t, err) {
			assert.Equal(t, test.wantError, err)
		}
	}
}
