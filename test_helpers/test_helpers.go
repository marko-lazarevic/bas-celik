package testhelpers

import (
	"github.com/ebfe/scard"
	"github.com/stretchr/testify/mock"
)


type CardMock struct {
	mock.Mock
}

func (m *CardMock) BeginTransaction() error {
	args := m.Called()
	return args.Error(0)
}

func (m *CardMock) EndTransaction(d scard.Disposition) error {
	args := m.Called(d)
	return args.Error(0)
}

func (m *CardMock) Status() (*scard.CardStatus, error) {
	args := m.Called()
	status, _ := args.Get(0).(*scard.CardStatus)
	return status, args.Error(1)
}

func (m *CardMock) Transmit(apdu []byte) ([]byte, error) {
	args := m.Called(apdu)
	bytes, _ := args.Get(0).([]byte)
	return bytes, args.Error(1)
}