package model

import (
	"time"

	"github.com/google/uuid"
)

const UserTableName = "tbl_user"

type User struct {
	Id       uuid.UUID `db:"id"`
	Username string    `db:"username"`
	Password string    `db:"password"`
	State    State     `db:"state"`
	CreateTs time.Time `db:"create_ts"`
	UpdateTs time.Time `db:"update_ts"`
	Version  int       `db:"version"`
}
