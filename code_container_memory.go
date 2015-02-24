package main

import (
	"github.com/realglobe-Inc/edo-toolkit/driver"
	"github.com/realglobe-Inc/go-lib/erro"
	"time"
)

type memoryCodeContainer codeContainerImpl

// スレッドセーフ。
func newMemoryCodeContainer(minIdLen int, procId string, savDur, ticExpDur, caStaleDur, caExpiDur time.Duration) codeContainer {
	return &memoryCodeContainer{
		driver.NewMemoryConcurrentVolatileKeyValueStore(caStaleDur, caExpiDur),
		newIdGenerator(minIdLen, procId),
		savDur,
		ticExpDur,
	}
}

func (this *memoryCodeContainer) get(codId string) (*code, error) {
	cod, err := ((*codeContainerImpl)(this)).get(codId)
	if err != nil {
		return nil, erro.Wrap(err)
	} else if cod == nil {
		return nil, nil
	}
	return cod.copy(), nil
}

func (this *memoryCodeContainer) put(cod *code) error {
	return ((*codeContainerImpl)(this)).put(cod.copy())
}

func (this *memoryCodeContainer) getAndSetEntry(codId string) (cod *code, tic string, err error) {
	cod, tic, err = ((*codeContainerImpl)(this)).getAndSetEntry(codId)
	if err != nil {
		return nil, "", erro.Wrap(err)
	} else if cod == nil {
		return nil, tic, nil
	}
	return cod.copy(), tic, nil
}

func (this *memoryCodeContainer) putIfEntered(cod *code, tic string) (ok bool, err error) {
	return ((*codeContainerImpl)(this)).putIfEntered(cod.copy(), tic)
}

func (this *memoryCodeContainer) close() error {
	return ((*codeContainerImpl)(this)).close()
}
