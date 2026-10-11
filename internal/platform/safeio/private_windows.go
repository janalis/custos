//go:build windows

package safeio

import (
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

// PrivateDirectory checks the Windows ACL instead of POSIX permission bits.
// Only the current user, SYSTEM and administrators may have allow entries.
func PrivateDirectory(path string, info os.FileInfo) bool {
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return false
	}
	sd, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.OWNER_SECURITY_INFORMATION)
	if err != nil {
		return false
	}
	owner, _, err := sd.Owner()
	if err != nil || owner == nil {
		return false
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil || !windows.EqualSid(owner, user.User.Sid) {
		return false
	}
	acl, _, err := sd.DACL()
	if err != nil || acl == nil {
		return false
	}
	for i := uint32(0); i < uint32(acl.AceCount); i++ {
		var ace *windows.ACCESS_ALLOWED_ACE
		if windows.GetAce(acl, i, &ace) != nil {
			return false
		}
		if ace.Header.AceType == windows.ACCESS_DENIED_ACE_TYPE {
			continue
		}
		if ace.Header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE {
			return false
		}
		sid := (*windows.SID)(unsafe.Pointer(&ace.SidStart))
		if !windows.EqualSid(sid, user.User.Sid) && sid.String() != "S-1-5-18" && sid.String() != "S-1-5-32-544" {
			return false
		}
	}
	return true
}
