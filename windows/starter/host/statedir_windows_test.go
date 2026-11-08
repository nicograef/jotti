package host

import (
	"encoding/binary"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"unsafe"
)

const (
	aceTypeAllowed        = 0
	aceFlagsDirInherit    = 0x3  // OBJECT_INHERIT_ACE | CONTAINER_INHERIT_ACE
	aceFlagInherited      = 0x10 // INHERITED_ACE
	seDaclProtected       = 0x1000
	maskFullAccess        = 0x1f01ff
	maskModify            = 0x1301bf
	sidUsers              = "S-1-5-32-545"
	errorNotAllAssigned   = syscall.Errno(1300)
	sePrivilegeEnabled    = 0x2
	tokenPrivilegesLength = 16 // PrivilegeCount + ein LUID_AND_ATTRIBUTES
)

var (
	procGetSecurityDescriptorControl = advapi32.NewProc("GetSecurityDescriptorControl")
	procLookupPrivilegeValueW        = advapi32.NewProc("LookupPrivilegeValueW")
	procAdjustTokenPrivileges        = advapi32.NewProc("AdjustTokenPrivileges")
)

type ace struct {
	typ   byte
	flags byte
	mask  uint32
	sid   string
}

// TestPrepareStateDirACL prueft die Rechte, die Windows tatsaechlich setzt, unter
// ProgramData, dessen geerbte ACL jedem Benutzer Lesezugriff gaebe.
func TestPrepareStateDirACL(t *testing.T) {
	user, _, err := stateDirUsers()
	if err != nil {
		t.Fatalf("stateDirUsers: %v", err)
	}
	wantDir := []ace{
		{aceTypeAllowed, aceFlagsDirInherit, maskFullAccess, "S-1-5-18"},
		{aceTypeAllowed, aceFlagsDirInherit, maskFullAccess, "S-1-5-32-544"},
		{aceTypeAllowed, aceFlagsDirInherit, maskModify, user},
	}
	wantFile := []ace{
		{aceTypeAllowed, aceFlagInherited, maskFullAccess, "S-1-5-18"},
		{aceTypeAllowed, aceFlagInherited, maskFullAccess, "S-1-5-32-544"},
		{aceTypeAllowed, aceFlagInherited, maskModify, user},
	}

	tests := []struct {
		name  string
		setup func(t *testing.T, dir string)
	}{
		{"neuer Ordner", func(*testing.T, string) {}},
		{"vorhandener Ordner mit geerbter ACL und .env", func(t *testing.T, dir string) {
			if err := os.Mkdir(dir, 0o750); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("JWT_SECRET=x\n"), 0o600); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := filepath.Join(programDataTempDir(t), "jotti")
			tt.setup(t, dir)

			// Zweimal wie bei jedem weiteren Start des Starters.
			for range 2 {
				if err := PrepareStateDir(dir); err != nil {
					t.Fatalf("PrepareStateDir: %v", err)
				}
			}

			protected, got := readDACL(t, dir)
			if !protected {
				t.Error("DACL des Ordners erbt noch von ProgramData")
			}
			assertACEs(t, dir, got, wantDir)

			env := filepath.Join(dir, ".env")
			if err := os.WriteFile(env, []byte("JWT_SECRET=y\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			_, got = readDACL(t, env)
			assertACEs(t, env, got, wantFile)
		})
	}
}

func TestPrepareStateDirRefusesForeignOwner(t *testing.T) {
	dir := filepath.Join(programDataTempDir(t), "jotti")
	if err := os.Mkdir(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	setOwner(t, dir, sidUsers)

	err := PrepareStateDir(dir)
	if err == nil || !strings.Contains(err.Error(), "gehoert") {
		t.Fatalf("PrepareStateDir = %v, want refusal for the owner %s", err, sidUsers)
	}
	assertUntouched(t, dir)
}

func TestPrepareStateDirRefusesJunction(t *testing.T) {
	base := programDataTempDir(t)
	target := filepath.Join(base, "ziel")
	if err := os.Mkdir(target, 0o750); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(base, "jotti")
	if out, err := exec.Command("cmd", "/c", "mklink", "/J", link, target).CombinedOutput(); err != nil {
		t.Fatalf("mklink /J: %v: %s", err, out)
	}

	err := PrepareStateDir(link)
	if err == nil || !strings.Contains(err.Error(), "kein gewoehnlicher Ordner") {
		t.Fatalf("PrepareStateDir = %v, want refusal of the junction", err)
	}
	assertUntouched(t, target)
}

func programDataTempDir(t *testing.T) string {
	t.Helper()
	base, err := os.MkdirTemp(os.Getenv("PROGRAMDATA"), "jotti-acl-test-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(base); err != nil { //nolint:gosec // base comes from os.MkdirTemp
			t.Errorf("cleanup: %v", err)
		}
	})
	return base
}

// assertUntouched prueft, dass ein verweigerter Ordner noch die geerbte ACL traegt.
func assertUntouched(t *testing.T, path string) {
	t.Helper()
	protected, got := readDACL(t, path)
	if protected {
		t.Errorf("%s: DACL protected after the refusal", path)
	}
	for _, a := range got {
		if a.flags&aceFlagInherited == 0 {
			t.Errorf("%s: explicit ACE %+v after the refusal", path, a)
		}
	}
}

func assertACEs(t *testing.T, path string, got, want []ace) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: ACEs = %+v, want %+v", path, got, want)
	}
	for _, w := range want {
		found := false
		for _, g := range got {
			found = found || g == w
		}
		if !found {
			t.Errorf("%s: ACEs = %+v, missing %+v", path, got, w)
		}
	}
}

// readDACL liest die DACL ueber GetNamedSecurityInfoW und zerlegt sie nach dem
// dokumentierten ACL-/ACE-Layout.
func readDACL(t *testing.T, path string) (protected bool, aces []ace) {
	t.Helper()
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	var dacl *byte
	var sd uintptr
	r, _, _ := procGetNamedSecurityInfoW.Call(uintptr(unsafe.Pointer(p)), seFileObject, daclSecurityInformation,
		0, 0, uintptr(unsafe.Pointer(&dacl)), 0, uintptr(unsafe.Pointer(&sd)))
	if r != 0 {
		t.Fatalf("GetNamedSecurityInfoW(%s): %v", path, syscall.Errno(r))
	}
	defer localFree(sd)

	var control uint16
	var revision uint32
	r, _, e := procGetSecurityDescriptorControl.Call(sd, uintptr(unsafe.Pointer(&control)), uintptr(unsafe.Pointer(&revision)))
	if r == 0 {
		t.Fatalf("GetSecurityDescriptorControl(%s): %v", path, e)
	}
	if dacl == nil {
		t.Fatalf("%s: no DACL", path)
	}

	// ACL: Revision, Sbz1, AclSize uint16, AceCount uint16, Sbz2; danach die ACEs.
	header := unsafe.Slice(dacl, 8)
	acl := unsafe.Slice(dacl, binary.LittleEndian.Uint16(header[2:]))
	off := 8
	for range binary.LittleEndian.Uint16(header[4:]) {
		// ACE: AceType, AceFlags, AceSize uint16, Mask uint32, dann die SID.
		a := ace{typ: acl[off], flags: acl[off+1], mask: binary.LittleEndian.Uint32(acl[off+4:])}
		if a.typ == aceTypeAllowed {
			sid, err := (*syscall.SID)(unsafe.Pointer(&acl[off+8])).String()
			if err != nil {
				t.Fatal(err)
			}
			a.sid = sid
		}
		aces = append(aces, a)
		off += int(binary.LittleEndian.Uint16(acl[off+2:]))
	}
	return control&seDaclProtected != 0, aces
}

// setOwner setzt einen fremden Besitzer; dafuer braucht der Prozess
// SeRestorePrivilege, das ein elevierter Administrator besitzt.
func setOwner(t *testing.T, path, ownerSID string) {
	t.Helper()
	enableRestorePrivilege(t)
	sid, err := syscall.StringToSid(ownerSID)
	if err != nil {
		t.Fatal(err)
	}
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		t.Fatal(err)
	}
	r, _, _ := procSetNamedSecurityInfoW.Call(uintptr(unsafe.Pointer(p)), seFileObject, ownerSecurityInformation,
		uintptr(unsafe.Pointer(sid)), 0, 0, 0)
	if r != 0 {
		t.Fatalf("SetNamedSecurityInfoW(owner %s): %v", ownerSID, syscall.Errno(r))
	}
}

func enableRestorePrivilege(t *testing.T) {
	t.Helper()
	process, err := syscall.GetCurrentProcess()
	if err != nil {
		t.Fatal(err)
	}
	var token syscall.Token
	if err := syscall.OpenProcessToken(process, syscall.TOKEN_ADJUST_PRIVILEGES|syscall.TOKEN_QUERY, &token); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = token.Close() })

	name, err := syscall.UTF16PtrFromString("SeRestorePrivilege")
	if err != nil {
		t.Fatal(err)
	}
	// TOKEN_PRIVILEGES mit einem Eintrag: PrivilegeCount, LUID (8 Byte), Attributes.
	var privileges, previous [tokenPrivilegesLength / 4]uint32
	privileges[0] = 1
	r, _, e := procLookupPrivilegeValueW.Call(0, uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(&privileges[1])))
	if r == 0 {
		t.Fatalf("LookupPrivilegeValueW: %v", e)
	}
	privileges[3] = sePrivilegeEnabled
	var returned uint32
	r, _, e = procAdjustTokenPrivileges.Call(uintptr(token), 0, uintptr(unsafe.Pointer(&privileges[0])),
		tokenPrivilegesLength, uintptr(unsafe.Pointer(&previous[0])), uintptr(unsafe.Pointer(&returned)))
	if r == 0 {
		t.Fatalf("AdjustTokenPrivileges: %v", e)
	}
	if errors.Is(e, errorNotAllAssigned) {
		t.Fatal("SeRestorePrivilege fehlt: den Test in einer elevierten Shell ausfuehren, wie jotti-start.exe")
	}
	t.Cleanup(func() {
		_, _, _ = procAdjustTokenPrivileges.Call(uintptr(token), 0, uintptr(unsafe.Pointer(&previous[0])), 0, 0, 0)
	})
}
