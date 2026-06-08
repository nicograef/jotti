package settings

import (
	"fmt"
	"net/netip"
	"time"
)

type DirektverkaufModus string

const (
	DirektverkaufModusKeinBon     DirektverkaufModus = "kein_bon"
	DirektverkaufModusAbholbon    DirektverkaufModus = "abholbon"
	DirektverkaufModusAnStationen DirektverkaufModus = "an_stationen"
)

type BondruckEinstellungen struct {
	KassenbelegDruckerIP string
	DirektverkaufModus   DirektverkaufModus
	AbholbonDruckerIP    string
	UpdatedAt            time.Time
}

func validateOptionalIPv4(ip string, field string) error {
	if ip == "" {
		return nil
	}

	addr, err := netip.ParseAddr(ip)
	if err != nil || !addr.Is4() {
		return fmt.Errorf("invalid %s", field)
	}

	return nil
}

func (b BondruckEinstellungen) Validate() error {
	if err := validateOptionalIPv4(b.KassenbelegDruckerIP, "kassenbeleg_drucker_ip"); err != nil {
		return err
	}

	switch b.DirektverkaufModus {
	case DirektverkaufModusKeinBon, DirektverkaufModusAbholbon, DirektverkaufModusAnStationen:
		// valid
	default:
		return fmt.Errorf("invalid direktverkauf_modus")
	}

	if err := validateOptionalIPv4(b.AbholbonDruckerIP, "abholbon_drucker_ip"); err != nil {
		return err
	}

	if b.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at is required")
	}

	return nil
}

func NewBondruckEinstellungen(kassenbelegDruckerIP string, direktverkaufModus DirektverkaufModus, abholbonDruckerIP string) (BondruckEinstellungen, error) {
	b := BondruckEinstellungen{
		KassenbelegDruckerIP: kassenbelegDruckerIP,
		DirektverkaufModus:   direktverkaufModus,
		AbholbonDruckerIP:    abholbonDruckerIP,
		UpdatedAt:            time.Now(),
	}

	if err := b.Validate(); err != nil {
		return BondruckEinstellungen{}, err
	}

	return b, nil
}
