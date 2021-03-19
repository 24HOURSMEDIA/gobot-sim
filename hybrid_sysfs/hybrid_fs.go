package hybrid_sysfs

import (
	"github.com/rs/zerolog/log"
	"gobot.io/x/gobot/sysfs"
	"os"
	"syscall"
)

// MockSyscall represents the hybrid sys call
type HybridSyscall struct {
	Impl func(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno)
}

// Syscall implements the SystemCaller interface
func (sys *HybridSyscall) Syscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err syscall.Errno) {
	if sys.Impl != nil {
		return sys.Impl(trap, a1, a2, a3)
	}
	return 0, 0, 0
}

// HybridFs delegates between to filesystems based on paths
type HybridFs struct {
	nativeFs      sysfs.Filesystem
	nativeSysCall sysfs.NativeSyscall
	mockFs        *sysfs.MockFilesystem
	mockSysCall   sysfs.MockSyscall
	mockablePaths map[string]bool
}

func NewHybridFs(nativeFs sysfs.Filesystem, mockFs *sysfs.MockFilesystem) *HybridFs {
	if len(mockFs.Files) > 0 {
		panic("mockFs cannot contain files and must be empty when injected")
	}
	fs := &HybridFs{
		nativeFs:      nativeFs,
		mockFs:        mockFs,
		mockablePaths: make(map[string]bool),
	}
	return fs
}

// AddMockablePath sets a file path that will be delegated to the mock file system
// instead of the native file system
func (hfs *HybridFs) AddMockablePath(name string) {
	hfs.mockablePaths[name] = true
	hfs.mockFs.Add(name)
}

func (hfs *HybridFs) OpenFile(name string, flag int, perm os.FileMode) (file sysfs.File, err error) {
	return hfs.selectFs(name).OpenFile(name, flag, perm)
}

func (hfs *HybridFs) Stat(name string) (os.FileInfo, error) {
	return hfs.selectFs(name).Stat(name)
}

// selectFs selects the appropriate filesystem based on a path
func (hfs *HybridFs) selectFs(name string) sysfs.Filesystem {
	mockable, found := hfs.mockablePaths[name]
	if found && mockable {
		log.Trace().Str("path", name).Msg("delegate to mock fs")
		return hfs.mockFs
	}
	log.Trace().Str("path", name).Msg("delegate to native fs")
	return hfs.nativeFs
}
