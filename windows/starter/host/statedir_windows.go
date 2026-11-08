package host

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/nicograef/jotti/windows/starter/core"
)

const (
	seFileObject                     = 1
	ownerSecurityInformation         = 0x1
	daclSecurityInformation          = 0x4
	protectedDaclSecurityInformation = 0x80000000
	sddlRevision1                    = 1
	wtsCurrentSession                = 0xFFFFFFFF
	wtsUserName                      = 5
	wtsDomainName                    = 7
)

// advapi32 und kernel32 laedt Go nur aus System32. wtsapi32 kommt per absolutem Pfad,
// sonst koennte eine DLL neben der Exe Code in den elevierten Prozess einschleusen.
var (
	advapi32                      = syscall.NewLazyDLL("advapi32.dll")
	procGetNamedSecurityInfoW     = advapi32.NewProc("GetNamedSecurityInfoW")
	procSetNamedSecurityInfoW     = advapi32.NewProc("SetNamedSecurityInfoW")
	procConvertStringSDToSD       = advapi32.NewProc("ConvertStringSecurityDescriptorToSecurityDescriptorW")
	procGetSecurityDescriptorDacl = advapi32.NewProc("GetSecurityDescriptorDacl")
	procGetSystemDirectoryW       = syscall.NewLazyDLL("kernel32.dll").NewProc("GetSystemDirectoryW")
)

// PrepareStateDir legt %PROGRAMDATA%\jotti an und sperrt es fuer andere lokale
// Benutzer: die geerbte ProgramData-ACL gaebe jedem Benutzer Lesezugriff auf .env
// und Backups. Ein vorgefundener Ordner wird nur mit unbedenklichem Besitzer uebernommen.
func PrepareStateDir(dir string) error {
	interactive, process, err := stateDirUsers()
	if err != nil {
		return errors.New(core.DiagnoseZustandsverzeichnisRechte(dir, err))
	}
	err = prepareProtectedDir(dir, core.StateDirSDDL(interactive), []string{interactive, process})
	var foreign *foreignDirError
	switch {
	case errors.As(err, &foreign) && foreign.owner == "":
		return errors.New(core.DiagnoseZustandsverzeichnisFremd(dir, foreign.Error()))
	case errors.As(err, &foreign):
		return errors.New(core.DiagnoseZustandsverzeichnisFremd(dir,
			"gehoert "+accountName(foreign.owner)+" statt SYSTEM, den Administratoren oder dem angemeldeten Benutzer"))
	case err != nil:
		return errors.New(core.DiagnoseZustandsverzeichnisRechte(dir, err))
	}
	return nil
}

// PreparePrivateDir legt einen Unterordner des Zustandsverzeichnisses mit eigener,
// nicht erbender DACL an (core.PrivateDirSDDL). Ein vorgefundener Ordner muss SYSTEM,
// den Administratoren oder dem Prozessbenutzer gehoeren: Hat der angemeldete Benutzer
// ihn angelegt, koennte er darin Konfiguration fuer den Postmaster hinterlegt haben.
func PreparePrivateDir(dir string) error {
	process, err := processUserSID()
	if err != nil {
		return fmt.Errorf("ermitteln des Prozessbenutzers: %w", err)
	}
	err = prepareProtectedDir(dir, core.PrivateDirSDDL(process), []string{process})
	var foreign *foreignDirError
	switch {
	case errors.As(err, &foreign) && foreign.owner == "":
		return fmt.Errorf("%s %s", dir, foreign.Error())
	case errors.As(err, &foreign):
		return fmt.Errorf("%s gehoert %s statt SYSTEM, den Administratoren oder %s", dir,
			accountName(foreign.owner), accountName(process))
	case err != nil:
		return fmt.Errorf("setzen der Zugriffsrechte fuer %s: %w", dir, err)
	}
	return nil
}

// foreignDirError meldet einen vorgefundenen Ordner, der nicht uebernommen wird:
// owner ist der unzulaessige Besitzer, leer bei einer Verknuepfung oder Datei.
type foreignDirError struct {
	owner string
}

func (e *foreignDirError) Error() string {
	if e.owner == "" {
		return "ist kein gewoehnlicher Ordner, sondern eine Verknuepfung oder Datei"
	}
	return "gehoert " + e.owner
}

// prepareProtectedDir legt dir mit der DACL aus sddl an oder setzt sie auf einem
// vorgefundenen Ordner, dessen Besitzer SYSTEM, die Administratoren oder einer von owners ist.
func prepareProtectedDir(dir, sddl string, owners []string) error {
	sd, err := securityDescriptor(sddl)
	if err != nil {
		return err
	}
	defer localFree(sd)

	info, err := os.Lstat(dir)
	switch {
	case errors.Is(err, os.ErrNotExist):
		// Mit der DACL anlegen statt nachtraeglich setzen: kein Zeitfenster, in dem ein
		// anderer Benutzer eine Datei hineinlegen koennte.
		if err := createDirWithSD(dir, sd); err != nil {
			return fmt.Errorf("anlegen: %w", err)
		}
	case err != nil:
		return err
	case !info.IsDir():
		// Eine Verknuepfung koennte die Rechtevergabe auf einen Systemordner umlenken.
		return &foreignDirError{}
	default:
		owner, err := ownerSID(dir)
		if err != nil {
			return err
		}
		if !core.StateDirOwnerAllowed(owner, owners) {
			return &foreignDirError{owner: owner}
		}
	}
	return setProtectedDACL(dir, sd)
}

// stateDirUsers liefert den in dieser Sitzung angemeldeten Benutzer und den Benutzer
// des elevierten Prozesses. Bestaetigt ein anderes Admin-Konto die UAC-Abfrage,
// unterscheiden sie sich; das nicht-elevierte jotti-relay laeuft als der angemeldete.
func stateDirUsers() (interactive, process string, err error) {
	process, err = processUserSID()
	if err != nil {
		return "", "", fmt.Errorf("ermitteln des Prozessbenutzers: %w", err)
	}
	interactive, err = sessionUserSID()
	if err != nil {
		return "", "", fmt.Errorf("angemeldeten Benutzer ermitteln: %w", err)
	}
	return interactive, process, nil
}

func processUserSID() (string, error) {
	token, err := syscall.OpenCurrentProcessToken()
	if err != nil {
		return "", err
	}
	defer func() { _ = token.Close() }()
	user, err := token.GetTokenUser()
	if err != nil {
		return "", err
	}
	return user.User.Sid.String()
}

func sessionUserSID() (string, error) {
	wtsapi32, err := loadSystemDLL("wtsapi32.dll")
	if err != nil {
		return "", err
	}
	defer func() { _ = wtsapi32.Release() }()
	query, err := wtsapi32.FindProc("WTSQuerySessionInformationW")
	if err != nil {
		return "", err
	}
	free, err := wtsapi32.FindProc("WTSFreeMemory")
	if err != nil {
		return "", err
	}

	name, err := wtsSessionString(query, free, wtsUserName)
	if err != nil {
		return "", err
	}
	if name == "" {
		return "", errors.New("in dieser Sitzung ist kein Benutzer angemeldet")
	}
	domain, err := wtsSessionString(query, free, wtsDomainName)
	if err != nil {
		return "", err
	}
	sid, _, _, err := syscall.LookupSID("", domain+`\`+name)
	if err != nil {
		return "", err
	}
	return sid.String()
}

func wtsSessionString(query, free *syscall.Proc, class uintptr) (string, error) {
	var buf *uint16
	var size uint32
	r, _, e := query.Call(0, wtsCurrentSession, class, uintptr(unsafe.Pointer(&buf)), uintptr(unsafe.Pointer(&size)))
	if r == 0 {
		return "", e
	}
	defer func() { _, _, _ = free.Call(uintptr(unsafe.Pointer(buf))) }()
	// size zaehlt Bytes einschliesslich der abschliessenden Null.
	return syscall.UTF16ToString(unsafe.Slice(buf, size/2)), nil
}

func loadSystemDLL(name string) (*syscall.DLL, error) {
	buf := make([]uint16, syscall.MAX_PATH)
	n, _, e := procGetSystemDirectoryW.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if n == 0 || n >= uintptr(len(buf)) {
		return nil, fmt.Errorf("ermitteln des System32-Pfads: %w", e)
	}
	return syscall.LoadDLL(filepath.Join(syscall.UTF16ToString(buf[:n]), name))
}

func ownerSID(path string) (string, error) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return "", err
	}
	var owner *syscall.SID
	var sd uintptr
	r, _, _ := procGetNamedSecurityInfoW.Call(uintptr(unsafe.Pointer(p)), seFileObject, ownerSecurityInformation,
		uintptr(unsafe.Pointer(&owner)), 0, 0, 0, uintptr(unsafe.Pointer(&sd)))
	if r != 0 {
		return "", fmt.Errorf("lesen des Besitzers: %w", syscall.Errno(r))
	}
	defer localFree(sd)
	return owner.String()
}

// accountName macht eine SID fuer die Meldung lesbar; ohne Aufloesung bleibt die SID.
func accountName(sidString string) string {
	sid, err := syscall.StringToSid(sidString)
	if err != nil {
		return sidString
	}
	account, domain, _, err := sid.LookupAccount("")
	if err != nil {
		return sidString
	}
	return domain + `\` + account + " (" + sidString + ")"
}

// securityDescriptor liefert einen per LocalAlloc belegten Deskriptor; der Aufrufer
// gibt ihn mit localFree frei.
func securityDescriptor(sddl string) (uintptr, error) {
	p, err := syscall.UTF16PtrFromString(sddl)
	if err != nil {
		return 0, err
	}
	var sd uintptr
	r, _, e := procConvertStringSDToSD.Call(uintptr(unsafe.Pointer(p)), sddlRevision1, uintptr(unsafe.Pointer(&sd)), 0)
	if r == 0 {
		return 0, fmt.Errorf("erzeugen der Sicherheitsbeschreibung: %w", e)
	}
	return sd, nil
}

func createDirWithSD(dir string, sd uintptr) error {
	p, err := syscall.UTF16PtrFromString(dir)
	if err != nil {
		return err
	}
	sa := syscall.SecurityAttributes{SecurityDescriptor: sd}
	sa.Length = uint32(unsafe.Sizeof(sa))
	return syscall.CreateDirectory(p, &sa)
}

// setProtectedDACL ersetzt die DACL und kappt die Vererbung von ProgramData; Windows
// gibt die neuen vererbbaren Eintraege an vorhandene Dateien und Unterordner weiter.
func setProtectedDACL(dir string, sd uintptr) error {
	var present, defaulted int32
	var dacl uintptr
	r, _, e := procGetSecurityDescriptorDacl.Call(sd, uintptr(unsafe.Pointer(&present)), uintptr(unsafe.Pointer(&dacl)), uintptr(unsafe.Pointer(&defaulted)))
	if r == 0 {
		return fmt.Errorf("lesen der DACL: %w", e)
	}
	if present == 0 || dacl == 0 {
		return errors.New("die Sicherheitsbeschreibung enthaelt keine DACL")
	}
	p, err := syscall.UTF16PtrFromString(dir)
	if err != nil {
		return err
	}
	r, _, _ = procSetNamedSecurityInfoW.Call(uintptr(unsafe.Pointer(p)), seFileObject,
		daclSecurityInformation|protectedDaclSecurityInformation, 0, 0, dacl, 0)
	if r != 0 {
		return fmt.Errorf("setzen der Zugriffsrechte: %w", syscall.Errno(r))
	}
	return nil
}

func localFree(h uintptr) {
	_, _ = syscall.LocalFree(syscall.Handle(h))
}
