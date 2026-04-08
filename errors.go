package main

import "errors"

var (
	ErrInvalidPosition = errors.New("position must be between 0 and 8")
	ErrPositionTaken   = errors.New("position is already taken")
	ErrNotYourTurn     = errors.New("it is not your turn")
	ErrMatchFull       = errors.New("match is already full")
	ErrMatchNotStarted = errors.New("match has not started yet")
)
