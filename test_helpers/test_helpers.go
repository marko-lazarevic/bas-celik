package testhelpers

import (
	"github.com/ebfe/scard"
	"github.com/stretchr/testify/mock"

	"github.com/ubavic/bas-celik/v2/document"
)

type ApolloMock struct {
	mock.Mock
}

func (m *ApolloMock) InitCard() error {
	args := m.Called()
	return args.Error(0)
}

func (m *ApolloMock) ReadCard() error {
	args := m.Called()
	return args.Error(0)
}

func (m *ApolloMock) GetDocument() (document.Document, error) {
	args := m.Called()
	doc, _ := args.Get(0).(document.Document)
	return doc, args.Error(1)
}

func (m *ApolloMock) Atr() []byte {
	args := m.Called()
	return args.Get(0).([]byte)
}

func (m *ApolloMock) ReadFile(name []byte) ([]byte, error) {
	args := m.Called(name)
	bytes, _ := args.Get(0).([]byte)
	return bytes, args.Error(1)
}

func (m *ApolloMock) Test() bool {
	args := m.Called()
	return args.Bool(0)
}

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