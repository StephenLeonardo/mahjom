package state

import "errors"

var (
	ErrActionNotAllowed = errors.New("action is not allowed")
)

type IStateActionables interface {
	Draw() error
	Discard() error
	Claim() error
	CheckWin() error
	Win() error
	GetStateName() string
}

type defaultState struct{}

func (s *defaultState) Draw() error {
	return ErrActionNotAllowed
}

func (s *defaultState) Discard() error {
	return ErrActionNotAllowed
}

func (s *defaultState) Claim() error {
	return ErrActionNotAllowed
}

func (s *defaultState) CheckWin() error {
	return ErrActionNotAllowed
}

func (s *defaultState) Win() error {
	return ErrActionNotAllowed
}

func (s *defaultState) GetStateName() string {
	return "default state"
}
