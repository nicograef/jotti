package settings

import (
	"time"

	"github.com/google/uuid"
)

type Kassenidentitaet struct {
	Seriennummer uuid.UUID
	AngelegtAm   time.Time
}
