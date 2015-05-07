// Copyright 2015 realglobe, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pairwise

import (
	"github.com/realglobe-Inc/go-lib/erro"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	tag_id = "_id"

	tagAccount  = "account"
	tagSector   = "sector"
	tagPairwise = "pairwise"
)

// mongodb を使った同意情報の格納庫。
type mongoDb struct {
	pool *mgo.Session
	db   string
	coll string
}

// db: DB 名。
// coll: コレクション名。
func NewMongoDb(pool *mgo.Session, db, coll string) Db {
	return &mongoDb{
		pool: pool,
		db:   db,
		coll: coll,
	}
}

func (this *mongoDb) GetByPairwise(sect, pw string) (*Element, error) {
	conn := this.pool.New()
	defer conn.Close()

	var elem Element
	if err := conn.DB(this.db).C(this.coll).Find(bson.M{tagSector: sect, tagPairwise: pw}).Select(bson.M{tag_id: 0}).One(&elem); err != nil {
		if err == mgo.ErrNotFound {
			return nil, nil
		}
		return nil, erro.Wrap(err)
	}

	return &elem, nil
}

func (this *mongoDb) Save(elem *Element) error {
	conn := this.pool.New()
	defer conn.Close()

	if _, err := conn.DB(this.db).C(this.coll).Upsert(bson.M{tagAccount: elem.Account(), tagSector: elem.Sector()}, elem); err != nil {
		return erro.Wrap(err)
	}

	return nil
}
