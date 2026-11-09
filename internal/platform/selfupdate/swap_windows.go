//go:build windows

package selfupdate

import "os"

// swap moves the new binary tmp over exe. Windows cannot overwrite a
// running executable but can rename it, so the old binary is moved aside
// to exe.old (removed by the next update).
func swap(tmp, exe string) error {
	old := exe + ".old"
	os.Remove(old)
	if err := os.Rename(exe, old); err != nil {
		return err
	}
	if err := os.Rename(tmp, exe); err != nil {
		os.Rename(old, exe)
		return err
	}
	return nil
}
