/* Copyright (C) 2021 Intel Corporation
 * SPDX-License-Identifier: Apache-2.0
 */

package bpf

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang -target bpf -cflags "-D__TARGET_ARCH_x86" redir   lib/redir.c
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang -target bpf -cflags "-D__TARGET_ARCH_x86" sockops lib/sockops.c
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang -target bpf -cflags "-D__TARGET_ARCH_x86" redir_disable lib/redir_disable.c

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	v2 "github.com/containers/common/pkg/cgroupv2"
	"golang.org/x/sys/unix"
)

const (
	FilesystemTypeBPFFS = unix.BPF_FS_MAGIC
	MapsRoot            = "/sys/fs/bpf"
	MapsPinpath         = "/sys/fs/bpf/tc/globals"
)

type BypassProgram struct {
	sockops_Obj   sockopsObjects
	redir_Obj     redirObjects
	SockopsCgroup link.Link
}

func SetLimit() error {
	var err error = nil

	err = unix.Setrlimit(unix.RLIMIT_MEMLOCK,
		&unix.Rlimit{
			Cur: unix.RLIM_INFINITY,
			Max: unix.RLIM_INFINITY,
		})
	if err != nil {
		fmt.Println("failed to set rlimit:", err)
	}

	return err
}

func getCgroupPath() (string, error) {
	var err error = nil
	cgroupPath := "/sys/fs/cgroup"

	enabled, err := v2.Enabled()
	if !enabled {
		cgroupPath = filepath.Join(cgroupPath, "unified")
	}
	return cgroupPath, err
}

func LoadProgram(prog BypassProgram) (BypassProgram, error) {
	var err error
	var options ebpf.CollectionOptions

	err = os.MkdirAll(MapsPinpath, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}

	options.Maps.PinPath = MapsPinpath

	if err = loadRedirObjects(&prog.redir_Obj, &options); err != nil {
		fmt.Println("Error load objects:", err)
	}

	if err = link.RawAttachProgram(link.RawAttachProgramOptions{
		Target:  prog.redir_Obj.redirMaps.MapRedir.FD(),
		Program: prog.redir_Obj.redirPrograms.BpfRedirProxy,
		Attach:  ebpf.AttachSkMsgVerdict,
	}); err != nil {
		fmt.Printf("Error attaching to sockmap: %s\n", err)
	}

	if err = loadSockopsObjects(&prog.sockops_Obj, &options); err != nil {
		fmt.Println("Error load objects:", err)
	}

	if cgroupPath, err := getCgroupPath(); err == nil {
		prog.SockopsCgroup, err = link.AttachCgroup(link.CgroupOptions{
			Path:    cgroupPath,
			Attach:  ebpf.AttachCGroupSockOps,
			Program: prog.sockops_Obj.sockopsPrograms.BpfSockmap,
		})
		if err != nil {
			fmt.Printf("Error attaching sockops to cgroup: %s", err)
		}

	}

	return prog, err
}

func CloseProgram(prog BypassProgram) {
	var err error

	if prog.SockopsCgroup != nil {
		fmt.Printf("Closing sockops cgroup...\n")
		prog.SockopsCgroup.Close()
	}

	if prog.redir_Obj.redirPrograms.BpfRedirProxy != nil {
		err = link.RawDetachProgram(link.RawDetachProgramOptions{
			Target:  prog.redir_Obj.redirMaps.MapRedir.FD(),
			Program: prog.redir_Obj.redirPrograms.BpfRedirProxy,
			Attach:  ebpf.AttachSkMsgVerdict,
		})
		if err != nil {
			fmt.Printf("Error detaching '%s'\n", err)
		}

		fmt.Printf("Closing redirect prog...\n")
	}

	if prog.sockops_Obj.sockopsMaps.MapActiveEstab != nil {
		prog.sockops_Obj.sockopsMaps.MapActiveEstab.Unpin()
		prog.sockops_Obj.sockopsMaps.MapActiveEstab.Close()
	}

	if prog.sockops_Obj.sockopsMaps.MapProxy != nil {
		prog.sockops_Obj.sockopsMaps.MapProxy.Unpin()
		prog.sockops_Obj.sockopsMaps.MapProxy.Close()
	}

	if prog.sockops_Obj.sockopsMaps.MapRedir != nil {
		prog.sockops_Obj.sockopsMaps.MapRedir.Unpin()
		prog.sockops_Obj.sockopsMaps.MapRedir.Close()
	}

	if prog.sockops_Obj.sockopsMaps.DebugMap != nil {
		prog.sockops_Obj.sockopsMaps.DebugMap.Unpin()
		prog.sockops_Obj.sockopsMaps.DebugMap.Close()
	}

}

func CheckOrMountBPFFSDefault() error {
	var err error

	_, err = os.Stat(MapsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(MapsRoot, 0755); err != nil {
				return fmt.Errorf("unable to create bpf mount directory: %s", err)
			}
		}
	}

	fst := unix.Statfs_t{}
	err = unix.Statfs(MapsRoot, &fst)
	if err != nil {
		return &os.PathError{Op: "statfs", Path: MapsRoot, Err: err}
	} else if fst.Type == FilesystemTypeBPFFS {
		return nil
	}

	err = unix.Mount(MapsRoot, MapsRoot, "bpf", 0, "")
	if err != nil {
		return fmt.Errorf("failed to mount %s: %s", MapsRoot, err)
	}

	return nil
}
