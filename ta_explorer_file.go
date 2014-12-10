package main

import (
	"encoding/json"
	"github.com/realglobe-Inc/edo/driver"
	"github.com/realglobe-Inc/go-lib-rg/erro"
	"time"
)

func jsonKeyGen(before string) string {
	return before + ".json"
}

func taExplorerTreeMarshal(value interface{}) (data []byte, err error) {
	data, err = json.Marshal(value.(*taExplorerTree).toContainer())
	if err != nil {
		return nil, erro.Wrap(err)
	}
	return data, nil
}

// data を JSON として、map[string]string にデコードしてから taExplorerTree をつくる。
func taExplorerTreeUnmarshal(data []byte) (interface{}, error) {
	var uriToUuid map[string]string
	if err := json.Unmarshal(data, &uriToUuid); err != nil {
		return nil, erro.Wrap(err)
	}

	tree := newTaExplorerTree()
	tree.fromContainer(uriToUuid)
	return tree, nil
}

// スレッドセーフ。
func NewFileTaExplorer(path string, expiDur time.Duration) TaExplorer {
	return newTaExplorer(driver.NewFileKeyValueStore(path, jsonKeyGen, taExplorerTreeMarshal, taExplorerTreeUnmarshal, expiDur))
}
