/*
	This file contains test helper functions to mock linux sysfs directory.
	If a package need to access system sysfs it should call CreateTmpSysFs() before test
	then call RemoveTmpSysFs() once test is done for clean up.
*/

package utils

import (
	"os"
	"path/filepath"
)

const (
	OwnerReadWriteExecuteOthersReadExecuteAttrs = 0755
	OwnerReadWriteOthersReadAttrs               = 0644
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type tmpSysFs struct {
	dirRoot      string
	dirList      []string
	fileList     map[string][]byte
	netSymlinks  map[string]string
	devSymlinks  map[string]string
	vfSymlinks   map[string]string
	originalRoot *os.File
}

var ts = tmpSysFs{
	dirList: []string{
		"sys/class/net",
		"sys/bus/pci/devices",
		"sys/devices/pci0000:ae/0000:ae:00.0/0000:af:00.1/net/ib0",
		"sys/devices/pci0000:ae/0000:ae:00.0/0000:af:06.0/net/ib1",
		"sys/devices/pci0000:ae/0000:ae:00.0/0000:af:06.1/net/ib2",
		"sys/devices/pci0000:00/0000:00:02.0/0000:05:00.0/net/ib3",
		"sys/devices/pci0000:00/0000:00:02.0/0000:05:00.0/net/ib4",
	},
	fileList: map[string][]byte{
		"sys/devices/pci0000:ae/0000:ae:00.0/0000:af:00.1/sriov_numvfs": []byte("2"),
		"sys/devices/pci0000:00/0000:00:02.0/0000:05:00.0/sriov_numvfs": []byte("0"),
	},
	netSymlinks: map[string]string{
		"sys/class/net/ib0": "sys/devices/pci0000:ae/0000:ae:00.0/0000:af:00.1/net/ib0",
		"sys/class/net/ib1": "sys/devices/pci0000:ae/0000:ae:00.0/0000:af:06.0/net/ib1",
		"sys/class/net/ib2": "sys/devices/pci0000:ae/0000:ae:00.0/0000:af:06.1/net/ib2",
		"sys/class/net/ib3": "sys/devices/pci0000:00/0000:00:02.0/0000:05:00.0/net/ib3",
		"sys/class/net/ib4": "sys/devices/pci0000:00/0000:00:02.0/0000:05:00.0/net/ib4",
	},
	devSymlinks: map[string]string{
		"sys/class/net/ib0/device": "sys/devices/pci0000:ae/0000:ae:00.0/0000:af:00.1",
		"sys/class/net/ib1/device": "sys/devices/pci0000:ae/0000:ae:00.0/0000:af:06.0",
		"sys/class/net/ib2/device": "sys/devices/pci0000:ae/0000:ae:00.0/0000:af:06.1",
		"sys/class/net/ib3/device": "sys/devices/pci0000:00/0000:00:02.0/0000:05:00.0",
		"sys/class/net/ib4/device": "sys/devices/pci0000:00/0000:00:02.0/0000:05:00.0",

		"sys/bus/pci/devices/0000:af:00.1": "sys/devices/pci0000:ae/0000:ae:00.0/0000:af:00.1",
		"sys/bus/pci/devices/0000:af:06.0": "sys/devices/pci0000:ae/0000:ae:00.0/0000:af:06.0",
		"sys/bus/pci/devices/0000:af:06.1": "sys/devices/pci0000:ae/0000:ae:00.0/0000:af:06.1",
		"sys/bus/pci/devices/0000:05:00.0": "sys/devices/pci0000:00/0000:00:02.0/0000:05:00.0",
	},
	vfSymlinks: map[string]string{
		"sys/devices/pci0000:ae/0000:ae:00.0/0000:af:00.1/virtfn0": "sys/devices/pci0000:ae/0000:ae:00.0/0000:af:06.0",
		"sys/devices/pci0000:ae/0000:ae:00.0/0000:af:06.0/physfn":  "sys/devices/pci0000:ae/0000:ae:00.0/0000:af:00.1",

		"sys/devices/pci0000:ae/0000:ae:00.0/0000:af:00.1/virtfn1": "sys/devices/pci0000:ae/0000:ae:00.0/0000:af:06.1",
		"sys/devices/pci0000:ae/0000:ae:00.0/0000:af:06.1/physfn":  "sys/devices/pci0000:ae/0000:ae:00.0/0000:af:00.1",
	},
}

// CreateTmpSysFs create mock sysfs for testing
func CreateTmpSysFs() error {
	originalRoot, _ := os.Open("/")
	ts.originalRoot = originalRoot

	tmpdir, ioErr := os.MkdirTemp("/tmp", "ib-sriov-cni-plugin-testfiles-")
	if ioErr != nil {
		return ioErr
	}

	ts.dirRoot = tmpdir

	for _, dir := range ts.dirList {
		if err := os.MkdirAll(filepath.Join(ts.dirRoot, dir), OwnerReadWriteExecuteOthersReadExecuteAttrs); err != nil {
			return err
		}
	}
	for filename, body := range ts.fileList {
		if err := os.WriteFile(filepath.Join(ts.dirRoot, filename), body, OwnerReadWriteOthersReadAttrs); err != nil {
			return err
		}
	}

	for link, target := range ts.netSymlinks {
		if err := createSymlinks(filepath.Join(ts.dirRoot, link), filepath.Join(ts.dirRoot, target)); err != nil {
			return err
		}
	}

	for link, target := range ts.devSymlinks {
		if err := createSymlinks(filepath.Join(ts.dirRoot, link), filepath.Join(ts.dirRoot, target)); err != nil {
			return err
		}
	}

	for link, target := range ts.vfSymlinks {
		if err := createSymlinks(filepath.Join(ts.dirRoot, link), filepath.Join(ts.dirRoot, target)); err != nil {
			return err
		}
	}

	SysBusPci = filepath.Join(ts.dirRoot, SysBusPci)
	NetDirectory = filepath.Join(ts.dirRoot, NetDirectory)
	return nil
}

func createSymlinks(link, target string) error {
	if err := os.MkdirAll(target, OwnerReadWriteExecuteOthersReadExecuteAttrs); err != nil {
		return err
	}
	if err := os.Symlink(target, link); err != nil {
		return err
	}
	return nil
}

// RemoveTmpSysFs removes mocked sysfs
func RemoveTmpSysFs() error {
	if err := ts.originalRoot.Chdir(); err != nil {
		return err
	}
	if err := ts.originalRoot.Close(); err != nil {
		return err
	}
	if err := os.RemoveAll(ts.dirRoot); err != nil {
		return err
	}
	return nil
}
