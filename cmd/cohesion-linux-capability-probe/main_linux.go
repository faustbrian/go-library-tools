//go:build linux

// Command cohesion-linux-capability-probe fails closed unless a hosted Linux
// runner provides the bounded kernel primitives checked by this preflight.
package main

import (
	"context"
	"debug/elf"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	overflowID = 65534

	landlockRulesetVersion  = 1
	landlockRulePathBeneath = 1

	seccompSetModeFilter = 2
	seccompReturnAllow   = 0x7fff0000
	seccompReturnErrno   = 0x00050000

	bpfLoadWordAbsolute = 0x20
	bpfJumpEqual        = 0x15
	bpfReturn           = 0x06

	securebitsLocked = 1<<0 | 1<<1 | 1<<2 | 1<<3 | 1<<5 | 1<<6 | 1<<7
)

var landlockHandled = uint64(
	unix.LANDLOCK_ACCESS_FS_EXECUTE |
		unix.LANDLOCK_ACCESS_FS_WRITE_FILE |
		unix.LANDLOCK_ACCESS_FS_READ_FILE |
		unix.LANDLOCK_ACCESS_FS_READ_DIR |
		unix.LANDLOCK_ACCESS_FS_REMOVE_DIR |
		unix.LANDLOCK_ACCESS_FS_REMOVE_FILE |
		unix.LANDLOCK_ACCESS_FS_MAKE_CHAR |
		unix.LANDLOCK_ACCESS_FS_MAKE_DIR |
		unix.LANDLOCK_ACCESS_FS_MAKE_REG |
		unix.LANDLOCK_ACCESS_FS_MAKE_SOCK |
		unix.LANDLOCK_ACCESS_FS_MAKE_FIFO |
		unix.LANDLOCK_ACCESS_FS_MAKE_BLOCK |
		unix.LANDLOCK_ACCESS_FS_MAKE_SYM |
		unix.LANDLOCK_ACCESS_FS_REFER |
		unix.LANDLOCK_ACCESS_FS_TRUNCATE)

type landlockRulesetAttr struct {
	HandledAccessFS  uint64
	HandledAccessNet uint64
	Scoped           uint64
}

type landlockPathBeneathAttr struct {
	AllowedAccess uint64
	ParentFD      int32
	_             uint32
}

type fOwnerEx struct {
	Type int32
	PID  int32
}

func main() {
	var err error
	switch strings.Join(os.Args[1:], " ") {
	case "pidfd-sender":
		err = writeByteToFD(3)
	case "lease-writer":
		err = overwrite(os.Getenv("LEASE_PATH"))
	case "landlock-child":
		err = probeLandlockChild(os.Getenv("LANDLOCK_ROOT"))
	case "seccomp-child":
		err = probeSeccompChild()
	case "namespace-cgroup-child":
		ctx, stop := signal.NotifyContext(context.Background(), unix.SIGINT, unix.SIGTERM)
		err = probeNamespaceCgroupChild(ctx)
		stop()
	case "cgroup-leaf-child":
		err = probeCgroupLeafChild()
	case "":
		ctx, stop := signal.NotifyContext(context.Background(), unix.SIGINT, unix.SIGTERM)
		err = run(ctx)
		stop()
	default:
		err = fmt.Errorf("unknown probe mode")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "unsupported:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) (resultErr error) {
	kernelRelease, err := probeKernelAndIdentity()
	if err != nil {
		return err
	}
	if err := probeStaticExecutable(ctx); err != nil {
		return err
	}
	if err := probePIDFDCredentials(ctx); err != nil {
		return err
	}
	root, err := os.MkdirTemp("", "golib-linux-capability-probe.")
	if err != nil {
		return err
	}
	cleaned := false
	defer func() {
		if cleaned {
			return
		}
		if cleanupErr := removeAndVerify(root); cleanupErr != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("task cleanup: %w", cleanupErr))
		}
	}()
	if err := os.Chmod(root, 0o700); err != nil {
		return err
	}
	if err := probeLandlock(ctx, root); err != nil {
		return err
	}
	if err := probeLease(ctx, root); err != nil {
		return err
	}
	if err := probeSeccomp(ctx); err != nil {
		return err
	}
	if err := probeCgroupAndNamespaces(ctx, root); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("supervisor cancelled: %w", err)
	}
	if err := removeAndVerify(root); err != nil {
		return fmt.Errorf("task cleanup: %w", err)
	}
	cleaned = true
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("supervisor cancelled: %w", err)
	}
	fmt.Printf("supported-primitives: linux/%s kernel=%s kernel>=6.9 userns mountns netns pidns landlock-abi>=4 pidfd cgroup-v2 lease signalfd static-elf seccomp cleanup\n", runtime.GOARCH, kernelRelease)
	return nil
}

func probeKernelAndIdentity() (string, error) {
	if runtime.GOOS != "linux" || (runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64") {
		return "", fmt.Errorf("platform must be linux/amd64 or linux/arm64")
	}
	var name unix.Utsname
	if err := unix.Uname(&name); err != nil {
		return "", fmt.Errorf("uname: %w", err)
	}
	release := unix.ByteSliceToString(name.Release[:])
	parts := strings.SplitN(release, ".", 3)
	if len(parts) < 2 {
		return "", fmt.Errorf("unrecognized kernel release %q", release)
	}
	major, majorErr := strconv.Atoi(parts[0])
	minor, minorErr := strconv.Atoi(parts[1])
	if majorErr != nil || minorErr != nil || major < 6 || (major == 6 && minor < 9) {
		return "", fmt.Errorf("kernel %q is older than 6.9", release)
	}
	uid, gid := os.Geteuid(), os.Getegid()
	if uid == 0 || gid == 0 || uid == overflowID || gid == overflowID {
		return "", fmt.Errorf("outer identity uid=%d gid=%d is root or reserved", uid, gid)
	}
	status, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return "", fmt.Errorf("read capability state: %w", err)
	}
	for _, field := range []string{"CapPrm:", "CapEff:"} {
		value, ok := statusField(string(status), field)
		if !ok || value != "0000000000000000" {
			return "", fmt.Errorf("%s must be empty, got %q", field, value)
		}
	}
	return release, nil
}

func probeStaticExecutable(parent context.Context) error {
	path, err := os.Executable()
	if err != nil {
		return err
	}
	file, err := elf.Open(path)
	if err != nil {
		return fmt.Errorf("open executing ELF: %w", err)
	}
	defer file.Close()
	wantMachine := map[string]elf.Machine{"amd64": elf.EM_X86_64, "arm64": elf.EM_AARCH64}[runtime.GOARCH]
	if file.Machine != wantMachine {
		return fmt.Errorf("ELF machine=%v want=%v", file.Machine, wantMachine)
	}
	for _, program := range file.Progs {
		if program.Type == elf.PT_INTERP {
			return fmt.Errorf("executing ELF has an interpreter")
		}
	}
	libraries, err := file.ImportedLibraries()
	if err != nil {
		return fmt.Errorf("read ELF dependencies: %w", err)
	}
	if len(libraries) != 0 {
		return fmt.Errorf("executing ELF has dynamic dependencies")
	}
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, path, "pidfd-sender")
	configureGracefulCancel(command)
	left, right, err := socketPair()
	if err != nil {
		return err
	}
	defer left.Close()
	defer right.Close()
	command.ExtraFiles = []*os.File{right}
	if err := command.Start(); err != nil {
		return fmt.Errorf("execute static ELF: %w", err)
	}
	defer stopAndReap(command)
	_ = right.Close()
	if err := left.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return err
	}
	if _, err := left.Read(make([]byte, 1)); err != nil {
		_ = stopAndReap(command)
		return fmt.Errorf("static ELF child handshake: %w", err)
	}
	if err := command.Wait(); err != nil {
		return fmt.Errorf("static ELF child: %w", err)
	}
	return nil
}

func probePIDFDCredentials(parent context.Context) error {
	receiver, sender, err := socketPair()
	if err != nil {
		return err
	}
	defer receiver.Close()
	defer sender.Close()
	if err := unix.SetsockoptInt(int(receiver.Fd()), unix.SOL_SOCKET, unix.SO_PASSPIDFD, 1); err != nil {
		return fmt.Errorf("SO_PASSPIDFD: %w", err)
	}
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "pidfd-sender")
	configureGracefulCancel(command)
	command.ExtraFiles = []*os.File{sender}
	if err := command.Start(); err != nil {
		return err
	}
	defer stopAndReap(command)
	_ = sender.Close()
	if err := receiver.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return err
	}
	data := make([]byte, 1)
	control := make([]byte, unix.CmsgSpace(4))
	count, controlCount, recvErr := recvmsgBounded(int(receiver.Fd()), data, control, 5*time.Second)
	if recvErr != nil || count != 1 {
		_ = stopAndReap(command)
		return fmt.Errorf("receive SCM_PIDFD: bytes=%d err=%w", count, recvErr)
	}
	messages, err := unix.ParseSocketControlMessage(control[:controlCount])
	if err != nil || len(messages) != 1 || messages[0].Header.Level != unix.SOL_SOCKET || messages[0].Header.Type != unix.SCM_PIDFD || len(messages[0].Data) != 4 {
		_ = stopAndReap(command)
		return fmt.Errorf("SCM_PIDFD control shape is unavailable")
	}
	peerFD := int(*(*int32)(unsafe.Pointer(&messages[0].Data[0])))
	defer unix.Close(peerFD)
	unix.CloseOnExec(peerFD)
	openedFD, err := unix.PidfdOpen(command.Process.Pid, 0)
	if err != nil {
		_ = stopAndReap(command)
		return fmt.Errorf("pidfd_open: %w", err)
	}
	defer unix.Close(openedFD)
	if err := sameFileIdentity(peerFD, openedFD); err != nil {
		_ = stopAndReap(command)
		return fmt.Errorf("stable pidfd identity: %w", err)
	}
	if err := command.Wait(); err != nil {
		return err
	}
	return nil
}

func probeLandlock(parent context.Context, root string) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "landlock-child")
	configureGracefulCancel(command)
	command.Env = []string{"LANDLOCK_ROOT=" + root}
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("Landlock REFER/TRUNCATE probe: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func probeLandlockChild(root string) error {
	abi, _, errno := unix.Syscall6(unix.SYS_LANDLOCK_CREATE_RULESET, 0, 0, landlockRulesetVersion, 0, 0, 0)
	if errno != 0 || abi < 4 {
		return fmt.Errorf("Landlock ABI=%d errno=%v", abi, errno)
	}
	allowed := filepath.Join(root, "landlock-allowed")
	denied := filepath.Join(root, "landlock-denied")
	if err := os.MkdirAll(allowed, 0o700); err != nil {
		return err
	}
	if err := os.MkdirAll(denied, 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(allowed, "source"), []byte("source"), 0o600); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(denied, "target"), []byte("target"), 0o600); err != nil {
		return err
	}
	attr := landlockRulesetAttr{HandledAccessFS: landlockHandled}
	ruleset, _, errno := unix.Syscall6(unix.SYS_LANDLOCK_CREATE_RULESET, uintptr(unsafe.Pointer(&attr)), unsafe.Sizeof(attr), 0, 0, 0, 0)
	if errno != 0 {
		return fmt.Errorf("landlock_create_ruleset: %v", errno)
	}
	defer unix.Close(int(ruleset))
	allowedFD, err := unix.Open(allowed, unix.O_PATH|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return err
	}
	defer unix.Close(allowedFD)
	if err := addLandlockPathRule(int(ruleset), allowedFD, landlockHandled); err != nil {
		return err
	}
	deniedFD, err := unix.Open(denied, unix.O_PATH|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return err
	}
	defer unix.Close(deniedFD)
	if err := addLandlockPathRule(int(ruleset), deniedFD, landlockHandled&^(unix.LANDLOCK_ACCESS_FS_REFER|unix.LANDLOCK_ACCESS_FS_TRUNCATE)); err != nil {
		return err
	}
	if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
		return err
	}
	_, _, errno = unix.Syscall6(unix.SYS_LANDLOCK_RESTRICT_SELF, ruleset, 0, 0, 0, 0, 0)
	if errno != 0 {
		return fmt.Errorf("landlock_restrict_self: %v", errno)
	}
	if err := os.Rename(filepath.Join(allowed, "source"), filepath.Join(denied, "moved")); !errors.Is(err, syscall.EXDEV) && !errors.Is(err, syscall.EACCES) {
		return fmt.Errorf("Landlock REFER denial=%v", err)
	}
	if err := os.Truncate(filepath.Join(denied, "target"), 0); !errors.Is(err, syscall.EACCES) {
		return fmt.Errorf("Landlock TRUNCATE denial=%v", err)
	}
	if err := os.Truncate(filepath.Join(allowed, "source"), 0); err != nil {
		return fmt.Errorf("Landlock permitted truncate: %w", err)
	}
	return nil
}

func addLandlockPathRule(ruleset, parentFD int, access uint64) error {
	pathRule := landlockPathBeneathAttr{AllowedAccess: access, ParentFD: int32(parentFD)}
	_, _, errno := unix.Syscall6(unix.SYS_LANDLOCK_ADD_RULE, uintptr(ruleset), landlockRulePathBeneath, uintptr(unsafe.Pointer(&pathRule)), 0, 0, 0)
	if errno != 0 {
		return fmt.Errorf("landlock_add_rule: %v", errno)
	}
	return nil
}

func probeLease(parent context.Context, root string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	const leaseSignal = unix.Signal(40)
	path := filepath.Join(root, "lease")
	if err := os.WriteFile(path, []byte("leased"), 0o600); err != nil {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	owner := fOwnerEx{Type: 0, PID: int32(unix.Gettid())}
	if _, _, errno := unix.Syscall(unix.SYS_FCNTL, file.Fd(), unix.F_SETOWN_EX, uintptr(unsafe.Pointer(&owner))); errno != 0 {
		return fmt.Errorf("F_SETOWN_EX/F_OWNER_TID: %w", errno)
	}
	if _, err := unix.FcntlInt(file.Fd(), unix.F_SETSIG, int(leaseSignal)); err != nil {
		return fmt.Errorf("F_SETSIG: %w", err)
	}
	var mask unix.Sigset_t
	mask.Val[(leaseSignal-1)/64] |= 1 << ((leaseSignal - 1) % 64)
	var previousMask unix.Sigset_t
	if err := unix.PthreadSigmask(unix.SIG_BLOCK, &mask, &previousMask); err != nil {
		return fmt.Errorf("block lease signal: %w", err)
	}
	defer unix.PthreadSigmask(unix.SIG_SETMASK, &previousMask, nil)
	signalFD, err := unix.Signalfd(-1, &mask, unix.SFD_CLOEXEC|unix.SFD_NONBLOCK)
	if err != nil {
		return fmt.Errorf("signalfd: %w", err)
	}
	defer unix.Close(signalFD)
	if _, err := unix.FcntlInt(file.Fd(), unix.F_SETLEASE, unix.F_RDLCK); err != nil {
		return fmt.Errorf("F_SETLEASE read: %w", err)
	}
	leaseHeld := true
	defer func() {
		if leaseHeld {
			_, _ = unix.FcntlInt(file.Fd(), unix.F_SETLEASE, unix.F_UNLCK)
		}
	}()
	lease, err := unix.FcntlInt(file.Fd(), unix.F_GETLEASE, 0)
	if err != nil || lease != unix.F_RDLCK {
		return fmt.Errorf("F_GETLEASE=%d err=%v", lease, err)
	}
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "lease-writer")
	configureGracefulCancel(command)
	command.Env = []string{"LEASE_PATH=" + path}
	if err := command.Start(); err != nil {
		return err
	}
	defer stopAndReap(command)
	deadline := time.Now().Add(5 * time.Second)
	for {
		if err := parent.Err(); err != nil {
			return fmt.Errorf("lease supervisor cancelled: %w", err)
		}
		buffer := make([]byte, unsafe.Sizeof(unix.SignalfdSiginfo{}))
		if count, err := unix.Read(signalFD, buffer); err == nil {
			info := (*unix.SignalfdSiginfo)(unsafe.Pointer(&buffer[0]))
			if count != len(buffer) || info.Signo != uint32(leaseSignal) || info.Fd != int32(file.Fd()) {
				return fmt.Errorf("unexpected signalfd lease-break record")
			}
			lease, leaseErr := unix.FcntlInt(file.Fd(), unix.F_GETLEASE, 0)
			if leaseErr != nil || lease != unix.F_RDLCK {
				return fmt.Errorf("lease state after break=%d err=%v", lease, leaseErr)
			}
			break
		} else if !errors.Is(err, unix.EAGAIN) || time.Now().After(deadline) {
			_ = stopAndReap(command)
			return fmt.Errorf("lease-break signalfd: %w", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if _, err := unix.FcntlInt(file.Fd(), unix.F_SETLEASE, unix.F_UNLCK); err != nil {
		_ = stopAndReap(command)
		return fmt.Errorf("release lease: %w", err)
	}
	leaseHeld = false
	if err := command.Wait(); err != nil {
		return fmt.Errorf("lease-break writer: %w", err)
	}
	lease, err = unix.FcntlInt(file.Fd(), unix.F_GETLEASE, 0)
	if err != nil || lease != unix.F_UNLCK {
		return fmt.Errorf("final lease state=%d err=%v", lease, err)
	}
	return nil
}

func probeSeccomp(parent context.Context) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "seccomp-child")
	configureGracefulCancel(command)
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("seccomp install: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func probeSeccompChild() error {
	filters := []unix.SockFilter{
		{Code: bpfLoadWordAbsolute, K: 0},
		{Code: bpfJumpEqual, Jt: 0, Jf: 1, K: unix.SYS_CLONE3},
		{Code: bpfReturn, K: seccompReturnErrno | uint32(unix.EPERM)},
		{Code: bpfReturn, K: seccompReturnAllow},
	}
	program := unix.SockFprog{Len: uint16(len(filters)), Filter: &filters[0]}
	if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
		return err
	}
	if err := unix.Prctl(unix.PR_SET_SECCOMP, seccompSetModeFilter, uintptr(unsafe.Pointer(&program)), 0, 0); err != nil {
		return err
	}
	_, _, errno := unix.Syscall6(unix.SYS_CLONE3, 0, 0, 0, 0, 0, 0)
	if errno != unix.EPERM {
		return fmt.Errorf("seccomp clone3 denial=%v", errno)
	}
	return nil
}

func probeCgroupAndNamespaces(parent context.Context, taskRoot string) (resultErr error) {
	cgroupRoot, err := currentCgroupDirectory()
	if err != nil {
		return err
	}
	leaf := filepath.Join(cgroupRoot, "golib-probe-"+strconv.Itoa(os.Getpid()))
	if err := os.Mkdir(leaf, 0o755); err != nil {
		return fmt.Errorf("delegated writable cgroup v2 leaf: %w", err)
	}
	var command *exec.Cmd
	defer func() {
		if cleanupErr := cleanupCgroup(leaf, command); cleanupErr != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("cgroup cleanup: %w", cleanupErr))
		}
	}()
	limits := map[string]string{
		"pids.max":        "512",
		"memory.max":      "2147483648",
		"memory.swap.max": "0",
		"cpu.max":         "400000 100000",
	}
	for name, value := range limits {
		if err := os.WriteFile(filepath.Join(leaf, name), []byte(value), 0o600); err != nil {
			return fmt.Errorf("write %s: %w", name, err)
		}
	}
	if err := verifyLimits(leaf, limits); err != nil {
		return err
	}
	cgroupFD, err := unix.Open(leaf, unix.O_DIRECTORY|unix.O_RDONLY|unix.O_CLOEXEC, 0)
	if err != nil {
		return err
	}
	cgroupFile := os.NewFile(uintptr(cgroupFD), "cgroup")
	if cgroupFile == nil {
		_ = unix.Close(cgroupFD)
		return fmt.Errorf("own cgroup descriptor")
	}
	cgroupFileClosed := false
	defer func() {
		if !cgroupFileClosed {
			_ = cgroupFile.Close()
		}
	}()
	parentSock, childSock, err := socketPair()
	if err != nil {
		return err
	}
	defer parentSock.Close()
	defer childSock.Close()
	ctx, cancel := context.WithTimeout(parent, 30*time.Second)
	defer cancel()
	command = exec.CommandContext(ctx, os.Args[0], "namespace-cgroup-child")
	configureGracefulCancel(command)
	command.ExtraFiles = []*os.File{cgroupFile, childSock}
	command.Env = []string{"PROBE_MOUNT_ROOT=" + taskRoot}
	command.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:                 unix.CLONE_NEWUSER | unix.CLONE_NEWNS | unix.CLONE_NEWNET | unix.CLONE_NEWPID,
		UidMappings:                []syscall.SysProcIDMap{{ContainerID: overflowID, HostID: os.Geteuid(), Size: 1}},
		GidMappings:                []syscall.SysProcIDMap{{ContainerID: overflowID, HostID: os.Getegid(), Size: 1}},
		GidMappingsEnableSetgroups: false,
		Credential:                 &syscall.Credential{Uid: overflowID, Gid: overflowID, NoSetGroups: true},
		AmbientCaps:                []uintptr{unix.CAP_NET_ADMIN, unix.CAP_SETPCAP, unix.CAP_SYS_ADMIN},
	}
	if err := command.Start(); err != nil {
		return fmt.Errorf("NEWUSER+NEWNS+NEWNET+NEWPID exact mapping: %w", err)
	}
	if err := cgroupFile.Close(); err != nil {
		return fmt.Errorf("close transferred cgroup descriptor: %w", err)
	}
	cgroupFileClosed = true
	if err := verifyNamespaceIsolation(command.Process.Pid); err != nil {
		return err
	}
	_ = childSock.Close()
	if err := parentSock.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return err
	}
	message := make([]byte, 16)
	control := make([]byte, unix.CmsgSpace(4))
	n, controlN, recvErr := recvmsgBounded(int(parentSock.Fd()), message, control, 5*time.Second)
	if recvErr != nil || string(message[:n]) != "ready" {
		return fmt.Errorf("mapped clone3 readiness: bytes=%q err=%v", message[:n], recvErr)
	}
	messages, err := unix.ParseSocketControlMessage(control[:controlN])
	if err != nil || len(messages) != 1 {
		return fmt.Errorf("mapped clone3 pidfd transfer")
	}
	fds, err := unix.ParseUnixRights(&messages[0])
	if err != nil || len(fds) != 1 {
		return fmt.Errorf("mapped clone3 pidfd rights")
	}
	childPIDFD := fds[0]
	defer unix.Close(childPIDFD)
	unix.CloseOnExec(childPIDFD)
	processes, err := oneCgroupPID(leaf)
	if err != nil {
		return err
	}
	comparisonFD, err := unix.PidfdOpen(processes, 0)
	if err != nil {
		return err
	}
	if err := sameFileIdentity(childPIDFD, comparisonFD); err != nil {
		unix.Close(comparisonFD)
		return err
	}
	unix.Close(comparisonFD)
	if err := verifyLimits(leaf, limits); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(leaf, "cgroup.kill"), []byte("1"), 0o600); err != nil {
		return fmt.Errorf("cgroup.kill: %w", err)
	}
	if err := command.Wait(); err != nil {
		return fmt.Errorf("namespace init reap: %w", err)
	}
	if _, err := os.Stat(filepath.Join("/proc", strconv.Itoa(command.Process.Pid))); !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("namespace init cleanup left pid %d", command.Process.Pid)
	}
	if _, err := oneCgroupPID(leaf); err == nil || !strings.Contains(err.Error(), "zero") {
		return fmt.Errorf("cgroup leaf did not drain: %v", err)
	}
	return nil
}

func verifyNamespaceIsolation(pid int) error {
	for _, namespace := range []string{"user", "mnt", "net", "pid"} {
		outer, err := os.Readlink(filepath.Join("/proc/self/ns", namespace))
		if err != nil {
			return err
		}
		inner, err := os.Readlink(filepath.Join("/proc", strconv.Itoa(pid), "ns", namespace))
		if err != nil {
			return fmt.Errorf("read %s namespace: %w", namespace, err)
		}
		if inner == outer {
			return fmt.Errorf("%s namespace was not isolated", namespace)
		}
	}
	return nil
}

func probeNamespaceCgroupChild(parent context.Context) error {
	ruid, euid, suid := unix.Getresuid()
	rgid, egid, sgid := unix.Getresgid()
	if ruid != overflowID || euid != overflowID || suid != overflowID || rgid != overflowID || egid != overflowID || sgid != overflowID {
		return fmt.Errorf("inner identity uid=%d/%d/%d gid=%d/%d/%d", ruid, euid, suid, rgid, egid, sgid)
	}
	groups, err := os.Getgroups()
	if err != nil || len(groups) != 0 {
		return fmt.Errorf("supplementary groups=%v err=%v", groups, err)
	}
	uidMap, err := os.ReadFile("/proc/self/uid_map")
	if err != nil || strings.Join(strings.Fields(string(uidMap)), " ") != fmt.Sprintf("%d %d 1", overflowID, outerID("uid")) {
		return fmt.Errorf("uid_map=%q err=%v", uidMap, err)
	}
	gidMap, err := os.ReadFile("/proc/self/gid_map")
	if err != nil || strings.Join(strings.Fields(string(gidMap)), " ") != fmt.Sprintf("%d %d 1", overflowID, outerID("gid")) {
		return fmt.Errorf("gid_map=%q err=%v", gidMap, err)
	}
	setgroups, err := os.ReadFile("/proc/self/setgroups")
	if err != nil || strings.TrimSpace(string(setgroups)) != "deny" {
		return fmt.Errorf("setgroups=%q err=%v", setgroups, err)
	}
	if err := probeNamespaceMountAndNetwork(os.Getenv("PROBE_MOUNT_ROOT")); err != nil {
		return err
	}
	if err := dropCapabilities(); err != nil {
		return err
	}
	statusBytes, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return err
	}
	for _, field := range []string{"CapInh:", "CapPrm:", "CapEff:", "CapBnd:", "CapAmb:"} {
		value, ok := statusField(string(statusBytes), field)
		if !ok || value != "0000000000000000" {
			return fmt.Errorf("%s=%q", field, value)
		}
	}
	var noNewPrivileges uintptr
	if value, ok := statusField(string(statusBytes), "NoNewPrivs:"); !ok || value != "0" {
		return fmt.Errorf("unexpected initial no_new_privs=%q", value)
	}
	if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
		return err
	}
	if _, _, errno := unix.Syscall6(unix.SYS_PRCTL, unix.PR_GET_NO_NEW_PRIVS, 0, 0, 0, 0, 0); errno != 0 {
		return errno
	} else {
		noNewPrivileges = 1
	}
	if noNewPrivileges != 1 {
		return fmt.Errorf("no_new_privs not set")
	}
	if os.Getpid() != 1 {
		return fmt.Errorf("PID namespace init pid=%d", os.Getpid())
	}
	unix.CloseOnExec(3)
	unix.CloseOnExec(4)
	pidfd := -1
	ctx, cancel := context.WithTimeout(parent, 20*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "cgroup-leaf-child")
	configureGracefulCancel(command)
	command.Env = []string{}
	command.SysProcAttr = &syscall.SysProcAttr{UseCgroupFD: true, CgroupFD: 3, PidFD: &pidfd}
	if err := command.Start(); err != nil {
		return fmt.Errorf("clone3(CLONE_PIDFD|CLONE_INTO_CGROUP): %w", err)
	}
	defer stopAndReap(command)
	if pidfd < 0 {
		return fmt.Errorf("clone3 did not return pidfd")
	}
	defer unix.Close(int(pidfd))
	rights := unix.UnixRights(int(pidfd))
	if err := unix.Sendmsg(4, []byte("ready"), rights, nil, 0); err != nil {
		return fmt.Errorf("send clone pidfd: %w", err)
	}
	waitErr := command.Wait()
	if waitErr == nil {
		return fmt.Errorf("cgroup child exited without cgroup.kill")
	}
	var exitError *exec.ExitError
	if !errors.As(waitErr, &exitError) {
		return fmt.Errorf("reap cgroup child: %w", waitErr)
	}
	status, ok := exitError.Sys().(syscall.WaitStatus)
	if !ok || !status.Signaled() || status.Signal() != unix.SIGKILL {
		return fmt.Errorf("cgroup child status=%v", status)
	}
	return nil
}

func probeCgroupLeafChild() error {
	select {
	case <-time.After(20 * time.Second):
		return fmt.Errorf("cgroup.kill was not observed")
	}
}

func probeNamespaceMountAndNetwork(taskRoot string) (resultErr error) {
	if taskRoot == "" {
		return fmt.Errorf("missing task-owned mount root")
	}
	if err := bringLoopbackUp(); err != nil {
		return err
	}
	interfaces, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("enumerate network namespace: %w", err)
	}
	if len(interfaces) != 1 || interfaces[0].Name != "lo" || interfaces[0].Flags&(net.FlagLoopback|net.FlagUp) != net.FlagLoopback|net.FlagUp {
		return fmt.Errorf("network namespace interfaces=%v", interfaces)
	}
	if err := unix.Mount("", "/", "", unix.MS_REC|unix.MS_PRIVATE, ""); err != nil {
		return fmt.Errorf("make mount namespace private: %w", err)
	}
	mountpoint := filepath.Join(taskRoot, "mount-probe")
	if err := os.Mkdir(mountpoint, 0o700); err != nil {
		return err
	}
	mounted := false
	defer func() {
		if mounted {
			if err := unix.Unmount(mountpoint, 0); err != nil {
				resultErr = errors.Join(resultErr, fmt.Errorf("unmount task tmpfs: %w", err))
			}
		}
		if err := os.Remove(mountpoint); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("remove mountpoint: %w", err))
		}
	}()
	if err := unix.Mount("tmpfs", mountpoint, "tmpfs", unix.MS_NODEV|unix.MS_NOSUID|unix.MS_NOEXEC, "size=1048576"); err != nil {
		return fmt.Errorf("mount task tmpfs: %w", err)
	}
	mounted = true
	path := filepath.Join(mountpoint, "probe")
	if err := os.WriteFile(path, []byte("mounted"), 0o600); err != nil {
		return fmt.Errorf("write task tmpfs: %w", err)
	}
	content, err := os.ReadFile(path)
	if err != nil || string(content) != "mounted" {
		return fmt.Errorf("read task tmpfs=%q err=%v", content, err)
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	if err := unix.Unmount(mountpoint, 0); err != nil {
		return fmt.Errorf("unmount task tmpfs: %w", err)
	}
	mounted = false
	return nil
}

func bringLoopbackUp() error {
	socket, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return fmt.Errorf("open loopback control socket: %w", err)
	}
	defer unix.Close(socket)
	request, err := unix.NewIfreq("lo")
	if err != nil {
		return fmt.Errorf("create loopback request: %w", err)
	}
	if err := unix.IoctlIfreq(socket, unix.SIOCGIFFLAGS, request); err != nil {
		return fmt.Errorf("read loopback flags: %w", err)
	}
	request.SetUint16(request.Uint16() | unix.IFF_UP)
	if err := unix.IoctlIfreq(socket, unix.SIOCSIFFLAGS, request); err != nil {
		return fmt.Errorf("bring loopback up: %w", err)
	}
	if err := unix.IoctlIfreq(socket, unix.SIOCGIFFLAGS, request); err != nil {
		return fmt.Errorf("re-read loopback flags: %w", err)
	}
	if request.Uint16()&(unix.IFF_LOOPBACK|unix.IFF_UP) != unix.IFF_LOOPBACK|unix.IFF_UP {
		return fmt.Errorf("loopback flags=%#x", request.Uint16())
	}
	return nil
}

func dropCapabilities() error {
	if err := unix.Prctl(unix.PR_SET_SECUREBITS, securebitsLocked, 0, 0, 0); err != nil {
		return fmt.Errorf("lock securebits: %w", err)
	}
	securebits, _, errno := unix.Syscall6(unix.SYS_PRCTL, unix.PR_GET_SECUREBITS, 0, 0, 0, 0, 0)
	if errno != 0 || securebits != securebitsLocked {
		return fmt.Errorf("securebits=%#x want=%#x err=%v", securebits, securebitsLocked, errno)
	}
	if err := unix.Prctl(unix.PR_CAP_AMBIENT, unix.PR_CAP_AMBIENT_CLEAR_ALL, 0, 0, 0); err != nil {
		return fmt.Errorf("clear ambient capabilities: %w", err)
	}
	for capability := 0; capability < 64; capability++ {
		if err := unix.Prctl(unix.PR_CAPBSET_DROP, uintptr(capability), 0, 0, 0); err != nil && !errors.Is(err, unix.EINVAL) {
			return fmt.Errorf("drop bounding capability %d: %w", capability, err)
		}
	}
	header := unix.CapUserHeader{Version: unix.LINUX_CAPABILITY_VERSION_3, Pid: 0}
	data := [2]unix.CapUserData{}
	if err := unix.Capset(&header, &data[0]); err != nil {
		return fmt.Errorf("clear capability sets: %w", err)
	}
	return nil
}

func cleanupCgroup(leaf string, command *exec.Cmd) error {
	var cleanupErr error
	populated, err := cgroupPopulated(leaf)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		cleanupErr = errors.Join(cleanupErr, err)
	}
	if populated {
		if err := os.WriteFile(filepath.Join(leaf, "cgroup.kill"), []byte("1"), 0o600); err != nil {
			cleanupErr = errors.Join(cleanupErr, fmt.Errorf("cgroup.kill: %w", err))
		}
	}
	if command != nil && command.Process != nil && command.ProcessState == nil {
		if err := stopAndReap(command); err != nil {
			cleanupErr = errors.Join(cleanupErr, fmt.Errorf("reap namespace init: %w", err))
		}
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		populated, err = cgroupPopulated(leaf)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return cleanupErr
			}
			cleanupErr = errors.Join(cleanupErr, err)
			break
		}
		if !populated {
			break
		}
		if time.Now().After(deadline) {
			cleanupErr = errors.Join(cleanupErr, fmt.Errorf("cgroup remained populated"))
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err := os.Remove(leaf); err != nil {
		cleanupErr = errors.Join(cleanupErr, err)
	}
	if _, err := os.Stat(leaf); !errors.Is(err, os.ErrNotExist) {
		cleanupErr = errors.Join(cleanupErr, fmt.Errorf("cgroup leaf still exists"))
	}
	return cleanupErr
}

func cgroupPopulated(leaf string) (bool, error) {
	data, err := os.ReadFile(filepath.Join(leaf, "cgroup.events"))
	if err != nil {
		return false, err
	}
	value, ok := statusField(string(data), "populated")
	if !ok || (value != "0" && value != "1") {
		return false, fmt.Errorf("invalid cgroup.events populated state")
	}
	return value == "1", nil
}

func currentCgroupDirectory() (string, error) {
	data, err := os.ReadFile("/proc/self/cgroup")
	if err != nil {
		return "", err
	}
	var path string
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if strings.HasPrefix(line, "0::") {
			path = strings.TrimPrefix(line, "0::")
		}
	}
	if path == "" || strings.Contains(path, "..") {
		return "", fmt.Errorf("unified cgroup v2 membership unavailable")
	}
	root := filepath.Join("/sys/fs/cgroup", filepath.Clean("/"+path))
	if _, err := os.Stat(filepath.Join(root, "cgroup.controllers")); err != nil {
		return "", fmt.Errorf("cgroup v2 controllers: %w", err)
	}
	return root, nil
}

func verifyLimits(leaf string, limits map[string]string) error {
	for name, want := range limits {
		data, err := os.ReadFile(filepath.Join(leaf, name))
		if err != nil || strings.TrimSpace(string(data)) != want {
			return fmt.Errorf("%s=%q want=%q err=%v", name, strings.TrimSpace(string(data)), want, err)
		}
	}
	return nil
}

func oneCgroupPID(leaf string) (int, error) {
	data, err := os.ReadFile(filepath.Join(leaf, "cgroup.procs"))
	if err != nil {
		return 0, err
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return 0, fmt.Errorf("zero cgroup processes")
	}
	if len(fields) != 1 {
		return 0, fmt.Errorf("cgroup processes=%d", len(fields))
	}
	return strconv.Atoi(fields[0])
}

func socketPair() (*os.File, *os.File, error) {
	fds, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_SEQPACKET|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return nil, nil, err
	}
	return os.NewFile(uintptr(fds[0]), "socket-parent"), os.NewFile(uintptr(fds[1]), "socket-child"), nil
}

func recvmsgBounded(fd int, payload, control []byte, timeout time.Duration) (int, int, error) {
	milliseconds := int(timeout / time.Millisecond)
	if milliseconds <= 0 {
		return 0, 0, fmt.Errorf("receive timeout must be positive")
	}
	pollFDs := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
	count, err := unix.Poll(pollFDs, milliseconds)
	if err != nil {
		return 0, 0, fmt.Errorf("poll receive socket: %w", err)
	}
	if count != 1 || pollFDs[0].Revents&unix.POLLIN == 0 {
		return 0, 0, fmt.Errorf("receive socket timed out or closed: events=%#x", pollFDs[0].Revents)
	}
	payloadN, controlN, flags, _, err := unix.Recvmsg(fd, payload, control, unix.MSG_DONTWAIT)
	if err != nil {
		return 0, 0, err
	}
	if flags&(unix.MSG_TRUNC|unix.MSG_CTRUNC) != 0 {
		return 0, 0, fmt.Errorf("truncated receive message: flags=%#x", flags)
	}
	return payloadN, controlN, nil
}

func writeByteToFD(fd int) error {
	_, err := unix.Write(fd, []byte{'x'})
	return err
}

func overwrite(path string) error {
	return os.WriteFile(path, []byte("changed"), 0o600)
}

func configureGracefulCancel(command *exec.Cmd) {
	command.Cancel = func() error {
		return command.Process.Signal(unix.SIGTERM)
	}
	command.WaitDelay = 2 * time.Second
}

func stopAndReap(command *exec.Cmd) error {
	if command == nil || command.Process == nil || command.ProcessState != nil {
		return nil
	}
	_ = command.Process.Kill()
	done := make(chan error, 1)
	go func() {
		done <- command.Wait()
	}()
	select {
	case <-done:
		return nil
	case <-time.After(2 * time.Second):
		return fmt.Errorf("process reap timed out")
	}
}

func removeAndVerify(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("path still exists")
	}
	return nil
}

func sameFileIdentity(left, right int) error {
	var leftStat, rightStat unix.Stat_t
	if err := unix.Fstat(left, &leftStat); err != nil {
		return err
	}
	if err := unix.Fstat(right, &rightStat); err != nil {
		return err
	}
	if leftStat.Dev != rightStat.Dev || leftStat.Ino != rightStat.Ino || leftStat.Ino == 0 {
		return fmt.Errorf("pidfd identities differ")
	}
	return nil
}

func statusField(status, name string) (string, bool) {
	for _, line := range strings.Split(status, "\n") {
		if strings.HasPrefix(line, name) {
			return strings.TrimSpace(strings.TrimPrefix(line, name)), true
		}
	}
	return "", false
}

func outerID(kind string) int {
	data, err := os.ReadFile("/proc/self/" + kind + "_map")
	if err != nil {
		return -1
	}
	fields := strings.Fields(string(data))
	if len(fields) != 3 {
		return -1
	}
	value, err := strconv.Atoi(fields[1])
	if err != nil {
		return -1
	}
	return value
}
