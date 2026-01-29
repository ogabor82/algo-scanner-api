package db

import "errors"

var ErrMissingDatabaseURL = errors.New("DATABASE_URL is not set")
