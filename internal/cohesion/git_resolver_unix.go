//go:build darwin || linux

//nolint:modernize,noctx // preserve platform contract
package cohesion

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"crypto/sha1" // Git object identity is SHA-1 for this locked repository set.
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

const maximumResolvedGitEntries = 4096

type looseGitRepository struct {
	rootFD int
	gitFD  int
	root   string
}

type resolvedGitEntry struct {
	mode   uint32
	object string
}

func resolveGitSourceFiles(root, repository, revision string, requests []gitSourceRequest) ([][]byte, error) {
	repo, err := openLooseGitRepository(root, repository)
	if err != nil {
		return nil, err
	}
	defer repo.close()
	if err := repo.requireHEAD(revision); err != nil {
		return nil, err
	}
	commit, err := repo.readObject(revision, "commit", maximumSchemaV3ArtifactBytes)
	if err != nil {
		return nil, err
	}
	treeID, err := commitTreeID(commit)
	if err != nil {
		return nil, err
	}
	entries := make(map[string]resolvedGitEntry)
	if err := repo.walkTree(treeID, "", entries); err != nil {
		return nil, err
	}
	if err := repo.requireCleanIndexAndWorktree(entries); err != nil {
		return nil, err
	}
	contents := make([][]byte, 0, len(requests))
	for _, request := range requests {
		if !safeRelativePath(request.path) || request.maximumBytes < 0 {
			return nil, errors.New("Git source request is invalid")
		}
		entry, exists := entries[request.path]
		if !exists || (entry.mode != 0o100644 && entry.mode != 0o100755) {
			return nil, errors.New("Git source path is not a regular blob")
		}
		content, err := repo.readObject(entry.object, "blob", request.maximumBytes)
		if err != nil {
			return nil, err
		}
		contents = append(contents, content)
	}
	if err := repo.requireHEAD(revision); err != nil {
		return nil, err
	}
	if err := repo.requireCleanIndexAndWorktree(entries); err != nil {
		return nil, errors.New("Git source changed during resolution")
	}
	return contents, nil
}

func verifyGitReleaseTag(root, repository, release, tagObject, peeledCommit string) error {
	repo, err := openLooseGitRepository(root, repository)
	if err != nil {
		return err
	}
	defer repo.close()
	if err := repo.requireHEAD(peeledCommit); err != nil {
		return err
	}
	content, err := repo.readObject(tagObject, "tag", maximumSchemaV3ArtifactBytes)
	if err != nil {
		return errors.New("release tag object is not an exact annotated tag")
	}
	object, objectType, tag, err := annotatedTagIdentity(content)
	if err != nil || object != peeledCommit || objectType != "commit" || tag != release {
		return errors.New("release annotated tag identity does not match")
	}
	if _, err := repo.readObject(peeledCommit, "commit", maximumSchemaV3ArtifactBytes); err != nil {
		return errors.New("release tag peel does not resolve to its commit")
	}
	return nil
}

func openLooseGitRepository(root, repository string) (*looseGitRepository, error) {
	rootFD, err := openAbsoluteDirectoryNoFollow(root)
	if err != nil {
		return nil, errors.New("open Git source root")
	}
	gitFD, err := unix.Openat(rootFD, ".git", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		_ = unix.Close(rootFD)
		return nil, errors.New("Git source must contain a descriptor-local .git directory")
	}
	repo := &looseGitRepository{rootFD: rootFD, gitFD: gitFD, root: root}
	if err := repo.rejectExternalObjectSources(); err != nil {
		repo.close()
		return nil, err
	}
	config, _, err := readRegularAt(repo.gitFD, "config", 1<<20)
	if err != nil {
		repo.close()
		return nil, errors.New("read Git source configuration")
	}
	remote, err := originURLFromGitConfig(config)
	if err != nil || !gitRemoteMatchesRepository(remote, repository) {
		repo.close()
		return nil, errors.New("Git source repository identity does not match")
	}
	return repo, nil
}

func (repo *looseGitRepository) close() {
	if repo.gitFD >= 0 {
		_ = unix.Close(repo.gitFD)
		repo.gitFD = -1
	}
	if repo.rootFD >= 0 {
		_ = unix.Close(repo.rootFD)
		repo.rootFD = -1
	}
}

func (repo *looseGitRepository) rejectExternalObjectSources() error {
	if entryExistsAt(repo.gitFD, "objects/info/alternates") {
		return errors.New("Git alternates are not permitted")
	}
	if entryExistsAt(repo.gitFD, "refs/replace") {
		return errors.New("Git replace references are not permitted")
	}
	if packed, _, err := readRegularAt(repo.gitFD, "packed-refs", 32<<20); err == nil && bytes.Contains(packed, []byte(" refs/replace/")) {
		return errors.New("packed Git replace references are not permitted")
	}
	config, _, err := readRegularAt(repo.gitFD, "config", 1<<20)
	if err != nil {
		return errors.New("read Git source configuration")
	}
	lower := bytes.ToLower(config)
	for _, forbidden := range [][]byte{[]byte("[include"), []byte("promisor"), []byte("partialclone"), []byte("objectformat"), []byte("worktreeconfig"), []byte("core.worktree")} {
		if bytes.Contains(lower, forbidden) {
			return errors.New("Git configuration selects an unsupported external object source")
		}
	}
	packFD, err := openDirectoryAt(repo.gitFD, "objects/pack")
	if err == nil {
		defer unix.Close(packFD)
		duplicate, err := unix.Dup(packFD)
		if err != nil {
			return errors.New("inspect Git pack directory")
		}
		reader := os.NewFile(uintptr(duplicate), "git-pack-directory-reader")
		defer reader.Close()
		entries, err := reader.ReadDir(-1)
		if err != nil {
			return errors.New("inspect Git pack directory")
		}
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".promisor") {
				return errors.New("Git promisor packs are not permitted")
			}
		}
	}
	return nil
}

func (repo *looseGitRepository) requireHEAD(revision string) error {
	head, _, err := readRegularAt(repo.gitFD, "HEAD", 4096)
	if err != nil {
		return errors.New("read Git HEAD")
	}
	value := strings.TrimSpace(string(head))
	if strings.HasPrefix(value, "ref: ") {
		ref := strings.TrimPrefix(value, "ref: ")
		if !safeGitRef(ref) {
			return errors.New("Git HEAD reference is invalid")
		}
		data, _, readErr := readRegularAt(repo.gitFD, ref, 128)
		if readErr != nil {
			packed, _, packedErr := readRegularAt(repo.gitFD, "packed-refs", 32<<20)
			if packedErr != nil {
				return errors.New("Git HEAD reference is unavailable")
			}
			data, readErr = packedReference(packed, ref)
			if readErr != nil {
				return errors.New("Git HEAD reference is unavailable")
			}
		}
		value = strings.TrimSpace(string(data))
	}
	if value != revision {
		return errors.New("Git source HEAD does not match the locked revision")
	}
	return nil
}

func packedReference(data []byte, ref string) ([]byte, error) {
	for _, line := range bytes.Split(data, []byte{'\n'}) {
		fields := bytes.Fields(line)
		if len(fields) == 2 && string(fields[1]) == ref && resolutionMapGitSHAPattern.Match(fields[0]) {
			return append([]byte(nil), fields[0]...), nil
		}
	}
	return nil, errors.New("packed reference is unavailable")
}

func (repo *looseGitRepository) readObject(id, expectedType string, maximumBytes int64) ([]byte, error) {
	if !resolutionMapGitSHAPattern.MatchString(id) || maximumBytes < 0 {
		return nil, errors.New("Git object identity is invalid")
	}
	compressed, _, err := openRegularAt(repo.gitFD, "objects/"+id[:2]+"/"+id[2:])
	if err != nil {
		return repo.readPackedObject(id, expectedType, maximumBytes)
	}
	file := os.NewFile(uintptr(compressed), "git-loose-object")
	if file == nil {
		_ = unix.Close(compressed)
		return nil, errors.New("adopt Git object descriptor")
	}
	defer file.Close()
	reader, err := zlib.NewReader(file)
	if err != nil {
		return nil, errors.New("decompress Git object")
	}
	defer reader.Close()
	raw, err := io.ReadAll(io.LimitReader(reader, maximumBytes+4097))
	if err != nil || int64(len(raw)) > maximumBytes+4096 {
		return nil, errors.New("Git object exceeds its maximum byte length")
	}
	want, _ := hex.DecodeString(id)
	digest := sha1.Sum(raw)
	if !bytes.Equal(digest[:], want) {
		return nil, errors.New("Git object content does not match its identity")
	}
	header, content, found := bytes.Cut(raw, []byte{0})
	if !found {
		return nil, errors.New("Git object header is invalid")
	}
	kind, sizeText, found := strings.Cut(string(header), " ")
	size, sizeErr := strconv.ParseInt(sizeText, 10, 64)
	if !found || sizeErr != nil || size != int64(len(content)) || kind != expectedType || size > maximumBytes {
		return nil, errors.New("Git object type or size does not match")
	}
	return content, nil
}

func (repo *looseGitRepository) readPackedObject(id, expectedType string, maximumBytes int64) ([]byte, error) {
	command := exec.Command("/usr/bin/git", "-C", repo.root, "cat-file", "--batch")
	command.Env = []string{"GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_OPTIONAL_LOCKS=0", "PATH=/usr/bin:/bin"}
	stdin, err := command.StdinPipe()
	if err != nil {
		return nil, errors.New("open Git object reader")
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		return nil, errors.New("open Git object output")
	}
	if err := command.Start(); err != nil {
		return nil, errors.New("start Git object reader")
	}
	_, _ = io.WriteString(stdin, id+"\n")
	_ = stdin.Close()
	reader := bufio.NewReader(io.LimitReader(stdout, maximumBytes+128))
	header, err := reader.ReadString('\n')
	if err != nil {
		_ = command.Wait()
		return nil, errors.New("read Git packed object header")
	}
	fields := strings.Fields(header)
	if len(fields) != 3 || fields[0] != id || fields[1] != expectedType {
		_ = command.Wait()
		return nil, errors.New("Git packed object type or identity does not match")
	}
	size, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil || size < 0 || size > maximumBytes {
		_ = command.Wait()
		return nil, errors.New("Git packed object size is invalid")
	}
	content, err := io.ReadAll(io.LimitReader(reader, size))
	if err != nil || int64(len(content)) != size {
		_ = command.Wait()
		return nil, errors.New("read Git packed object")
	}
	if err := command.Wait(); err != nil {
		return nil, errors.New("Git packed object reader failed")
	}
	raw := append([]byte(expectedType+" "+strconv.FormatInt(size, 10)+"\x00"), content...)
	digest := sha1.Sum(raw)
	want, _ := hex.DecodeString(id)
	if !bytes.Equal(digest[:], want) {
		return nil, errors.New("Git packed object content does not match its identity")
	}
	return content, nil
}

func (repo *looseGitRepository) walkTree(id, prefix string, result map[string]resolvedGitEntry) error {
	content, err := repo.readObject(id, "tree", maximumSchemaV3ArtifactBytes)
	if err != nil {
		return err
	}
	for len(content) > 0 {
		space := bytes.IndexByte(content, ' ')
		nul := bytes.IndexByte(content, 0)
		if space <= 0 || nul <= space+1 || len(content) < nul+21 {
			return errors.New("Git tree entry is invalid")
		}
		mode64, err := strconv.ParseUint(string(content[:space]), 8, 32)
		name := string(content[space+1 : nul])
		if err != nil || !safeGitTreeName(name) {
			return errors.New("Git tree entry identity is invalid")
		}
		object := hex.EncodeToString(content[nul+1 : nul+21])
		content = content[nul+21:]
		path := name
		if prefix != "" {
			path = prefix + "/" + name
		}
		mode := uint32(mode64)
		switch mode {
		case 0o40000:
			if err := repo.walkTree(object, path, result); err != nil {
				return err
			}
		case 0o100644, 0o100755:
			if len(result) >= maximumResolvedGitEntries {
				return errors.New("Git tree exceeds its supported entry bound")
			}
			if _, duplicate := result[path]; duplicate {
				return errors.New("Git tree contains a duplicate path")
			}
			if _, err := repo.readObject(object, "blob", maximumSchemaV3ArtifactBytes); err != nil {
				return err
			}
			result[path] = resolvedGitEntry{mode: mode, object: object}
		default:
			return errors.New("Git tree contains an unsupported entry mode")
		}
	}
	return nil
}

func (repo *looseGitRepository) requireCleanIndexAndWorktree(tree map[string]resolvedGitEntry) error {
	index, _, err := readRegularAt(repo.gitFD, "index", 32<<20)
	if err != nil {
		return errors.New("read Git index")
	}
	indexed, err := parseGitIndexV2(index)
	if err != nil || len(indexed) != len(tree) {
		return errors.New("Git index does not match the locked tree")
	}
	for path, expected := range tree {
		actual, exists := indexed[path]
		if !exists || actual != expected {
			return errors.New("Git index does not match the locked tree")
		}
		content, mode, err := readRegularAt(repo.rootFD, path, maximumSchemaV3ArtifactBytes)
		if err != nil {
			return errors.New("Git worktree does not match the locked tree")
		}
		worktreeMode := uint32(0o100644)
		if mode&0o111 != 0 {
			worktreeMode = 0o100755
		}
		if worktreeMode != expected.mode || gitBlobID(content) != expected.object {
			return errors.New("Git worktree does not match the locked tree")
		}
	}
	return nil
}

func parseGitIndexV2(data []byte) (map[string]resolvedGitEntry, error) {
	if len(data) < 32 || string(data[:4]) != "DIRC" || binary.BigEndian.Uint32(data[4:8]) != 2 {
		return nil, errors.New("Git index format is unsupported")
	}
	digest := sha1.Sum(data[:len(data)-20])
	if !bytes.Equal(digest[:], data[len(data)-20:]) {
		return nil, errors.New("Git index checksum does not match")
	}
	count := int(binary.BigEndian.Uint32(data[8:12]))
	if count > maximumResolvedGitEntries {
		return nil, errors.New("Git index exceeds its supported entry bound")
	}
	result := make(map[string]resolvedGitEntry, count)
	offset := 12
	previous := ""
	for range count {
		start := offset
		if offset+62 > len(data)-20 {
			return nil, errors.New("Git index entry is truncated")
		}
		mode := binary.BigEndian.Uint32(data[offset+24 : offset+28])
		object := hex.EncodeToString(data[offset+40 : offset+60])
		flags := binary.BigEndian.Uint16(data[offset+60 : offset+62])
		if flags&0x7000 != 0 {
			return nil, errors.New("Git index stage or extended flags are unsupported")
		}
		offset += 62
		nul := bytes.IndexByte(data[offset:len(data)-20], 0)
		if nul < 1 {
			return nil, errors.New("Git index path is invalid")
		}
		path := string(data[offset : offset+nul])
		if !safeRelativePath(path) || (previous != "" && path <= previous) {
			return nil, errors.New("Git index paths are invalid")
		}
		previous = path
		declaredLength := int(flags & 0x0fff)
		if declaredLength != 0x0fff && declaredLength != len(path) {
			return nil, errors.New("Git index path length does not match")
		}
		if _, duplicate := result[path]; duplicate {
			return nil, errors.New("Git index contains a duplicate path")
		}
		result[path] = resolvedGitEntry{mode: mode, object: object}
		entryLength := 62 + nul + 1
		offset = start + (entryLength+7)&^7
	}
	for offset < len(data)-20 {
		if offset+8 > len(data)-20 {
			return nil, errors.New("Git index extension is truncated")
		}
		signature := data[offset : offset+4]
		if signature[0] < 'A' || signature[0] > 'Z' {
			return nil, errors.New("required Git index extensions are unsupported")
		}
		size := int(binary.BigEndian.Uint32(data[offset+4 : offset+8]))
		offset += 8
		if size < 0 || offset+size > len(data)-20 {
			return nil, errors.New("Git index extension length is invalid")
		}
		offset += size
	}
	return result, nil
}

func commitTreeID(content []byte) (string, error) {
	line, _, found := bytes.Cut(content, []byte{'\n'})
	if !found || !bytes.HasPrefix(line, []byte("tree ")) {
		return "", errors.New("Git commit tree identity is invalid")
	}
	id := string(bytes.TrimPrefix(line, []byte("tree ")))
	if !resolutionMapGitSHAPattern.MatchString(id) {
		return "", errors.New("Git commit tree identity is invalid")
	}
	return id, nil
}

func annotatedTagIdentity(content []byte) (string, string, string, error) {
	header, _, found := bytes.Cut(content, []byte("\n\n"))
	if !found {
		return "", "", "", errors.New("annotated tag header is invalid")
	}
	values := map[string]string{}
	for _, line := range strings.Split(string(header), "\n") {
		name, value, ok := strings.Cut(line, " ")
		if ok && (name == "object" || name == "type" || name == "tag") {
			if _, duplicate := values[name]; duplicate {
				return "", "", "", errors.New("annotated tag header is duplicated")
			}
			values[name] = value
		}
	}
	if len(values) != 3 {
		return "", "", "", errors.New("annotated tag identity is incomplete")
	}
	return values["object"], values["type"], values["tag"], nil
}

func originURLFromGitConfig(data []byte) (string, error) {
	section := ""
	remote := ""
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		if section == `remote "origin"` && strings.EqualFold(strings.TrimSpace(key), "url") {
			if remote != "" {
				return "", errors.New("Git origin URL is duplicated")
			}
			remote = strings.TrimSpace(value)
		}
	}
	if remote == "" {
		return "", errors.New("Git origin URL is missing")
	}
	return remote, nil
}

func gitRemoteMatchesRepository(remote, repository string) bool {
	path := strings.TrimPrefix(repository, "github.com/")
	return remote == "https://github.com/"+path || remote == "https://github.com/"+path+".git" ||
		remote == "git@github.com:"+path || remote == "git@github.com:"+path+".git" ||
		remote == "ssh://git@github.com/"+path || remote == "ssh://git@github.com/"+path+".git"
}

func gitBlobID(content []byte) string {
	header := []byte(fmt.Sprintf("blob %d%c", len(content), 0))
	digest := sha1.Sum(append(header, content...))
	return hex.EncodeToString(digest[:])
}

func safeGitTreeName(name string) bool {
	return name != "" && name != "." && name != ".." && !strings.ContainsAny(name, "/\\\x00")
}

func safeGitRef(ref string) bool {
	return strings.HasPrefix(ref, "refs/heads/") && safeRelativePath(ref) && !strings.Contains(ref, "..") && !strings.HasSuffix(ref, ".lock")
}

func openAbsoluteDirectoryNoFollow(path string) (int, error) {
	current, err := unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return -1, err
	}
	for _, component := range strings.Split(strings.TrimPrefix(path, "/"), "/") {
		if component == "" || component == "." || component == ".." {
			_ = unix.Close(current)
			return -1, errors.New("unsafe directory component")
		}
		next, err := unix.Openat(current, component, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		_ = unix.Close(current)
		if err != nil {
			return -1, err
		}
		current = next
	}
	return current, nil
}

func openDirectoryAt(rootFD int, path string) (int, error) {
	return openAtComponents(rootFD, path, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW)
}

func openRegularAt(rootFD int, path string) (int, uint32, error) {
	fd, err := openAtComponents(rootFD, path, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW)
	if err != nil {
		return -1, 0, err
	}
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil || stat.Mode&unix.S_IFMT != unix.S_IFREG {
		_ = unix.Close(fd)
		return -1, 0, errors.New("resolved path is not a regular file")
	}
	// Stat_t.Mode is uint16 on Darwin and uint32 on Linux. Normalize through
	// uint64 so both platform builds retain an explicit width-normalizing cast.
	mode := uint64(stat.Mode)
	return fd, uint32(mode), nil
}

func openAtComponents(rootFD int, path string, finalFlags int) (int, error) {
	if !safeRelativePath(path) {
		return -1, errors.New("unsafe relative path")
	}
	current, err := unix.Dup(rootFD)
	if err != nil {
		return -1, err
	}
	components := strings.Split(path, "/")
	for index, component := range components {
		flags := unix.O_RDONLY | unix.O_DIRECTORY | unix.O_CLOEXEC | unix.O_NOFOLLOW
		if index == len(components)-1 {
			flags = finalFlags
		}
		next, err := unix.Openat(current, component, flags, 0)
		_ = unix.Close(current)
		if err != nil {
			return -1, err
		}
		current = next
	}
	return current, nil
}

func readRegularAt(rootFD int, path string, maximumBytes int64) ([]byte, uint32, error) {
	fd, mode, err := openRegularAt(rootFD, path)
	if err != nil {
		return nil, 0, err
	}
	file := os.NewFile(uintptr(fd), "descriptor-rooted-file")
	if file == nil {
		_ = unix.Close(fd)
		return nil, 0, errors.New("adopt descriptor-rooted file")
	}
	defer file.Close()
	content, err := io.ReadAll(io.LimitReader(file, maximumBytes+1))
	if err != nil || int64(len(content)) > maximumBytes {
		return nil, 0, errors.New("descriptor-rooted file exceeds its bound")
	}
	return content, mode, nil
}

func entryExistsAt(rootFD int, path string) bool {
	parent, name := "", path
	if slash := strings.LastIndexByte(path, '/'); slash >= 0 {
		parent, name = path[:slash], path[slash+1:]
	}
	current, err := unix.Dup(rootFD)
	if err != nil {
		return false
	}
	if parent != "" {
		next, openErr := openDirectoryAt(current, parent)
		if openErr != nil {
			_ = unix.Close(current)
			return false
		}
		_ = unix.Close(current)
		current = next
	}
	defer unix.Close(current)
	var stat unix.Stat_t
	return unix.Fstatat(current, name, &stat, unix.AT_SYMLINK_NOFOLLOW) == nil
}
