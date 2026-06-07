package settings

import (
	"fmt"
	"net/netip"
	"time"
)

type BondruckEinstellungen struct {
	KassenbelegDruckerIP string
	UpdatedAt            time.Time
}

func (b BondruckEinstellungen) Validate() error {
	if b.KassenbelegDruckerIP != "" {
		addr, err := netip.ParseAddr(b.KassenbelegDruckerIP)
		if err != nil || !addr.Is4() {
			return fmt.Errorf("invalid kassenbeleg_drucker_ip")
		}
	}
	if b.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at is required")
	}
	return nil
}

func NewBondruckEinstellungen(kassenbelegDruckerIP string) (BondruckEinstellungen, error) {
	b := BondruckEinstellungen{
		KassenbelegDruckerIP: kassenbelegDruckerIP,
		UpdatedAt:            time.Now(),
	}
	if err := b.Validate(); err != nil {
		return BondruckEinstellungen{}, err
	}
	return b, nil
}
