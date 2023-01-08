package cache2gowithclient

import "errors"

var (
	ErrKeyNotFound      = errors.New("Key not found in cache")
	ErrSetValFailed     = errors.New("expected cmd: set [key] [value] [duration]")
	ErrGetValFailed     = errors.New("expected cmd: get [key]")
	ErrCheckExistFailed = errors.New("expected cmd: check [key]")
)
