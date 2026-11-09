package selfupdate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
	"strings"
)

// goos and goarch name the platform; tests replace them.
var goos, goarch = runtime.GOOS, runtime.GOARCH

// Asset is the name of the bare release binary for this platform.
func Asset() string {
	name := "custos_" + goos + "_" + goarch
	if goos == "windows" {
		name += ".exe"
	}
	return name
}

// AssetBase is where the assets of version are downloaded from:
// $CUSTOS_DOWNLOAD_URL (a mirror URL or a local directory), else the
// GitHub release.
func AssetBase(version string) string {
	if base := getenv("CUSTOS_DOWNLOAD_URL"); base != "" {
		return strings.TrimRight(base, "/")
	}
	return releasesURL + "/download/v" + version
}

// Replace downloads the release binary of version, checks it against the
// release's SHA256SUMS and puts it in place of exe.
func Replace(ctx context.Context, exe, version string) error {
	base, name := AssetBase(version), Asset()
	sums, err := fetch(ctx, base+"/SHA256SUMS")
	if err != nil {
		return err
	}
	want := ""
	for line := range strings.Lines(string(sums)) {
		f := strings.Fields(line)
		if len(f) == 2 && strings.TrimPrefix(f[1], "*") == name {
			want = strings.ToLower(f[0])
		}
	}
	if want == "" {
		return fmt.Errorf("%s is not listed in %s/SHA256SUMS", name, base)
	}
	bin, err := fetch(ctx, base+"/"+name)
	if err != nil {
		return err
	}
	if sum := sha256.Sum256(bin); hex.EncodeToString(sum[:]) != want {
		return fmt.Errorf("%s/%s: %w", base, name, errChecksum)
	}
	info, err := os.Stat(exe)
	if err != nil {
		return err
	}
	tmp := fmt.Sprintf("%s.%d.tmp", exe, os.Getpid())
	if err := os.WriteFile(tmp, bin, info.Mode().Perm()|0o100); err != nil {
		return fmt.Errorf("cannot write next to %s (%w); reinstall custos or rerun with write access there", exe, err)
	}
	if err := swap(tmp, exe); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("cannot replace %s: %w", exe, err)
	}
	return nil
}
