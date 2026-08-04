package state

import "errors"

var (
	ErrActionNotAllowed = errors.New("action is not allowed")
)

type IStateActionables interface {
	GetStateName() string

	Resolve() error
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
