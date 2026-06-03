package settings

import (
	"time"

	"github.com/google/uuid"
)

type SystemConfig struct {
	Seriennummer uuid.UUID
	AngelegtAm   time.Time
}
