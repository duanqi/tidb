// Copyright 2016 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package tidb

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/ngaut/log"
	"github.com/pingcap/tidb/ast"
)

var smallCount = 100
var bigCount = 10000

func prepareBenchSession() Session {
	store, err := NewStore("memory://bench")
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(log.LOG_LEVEL_ERROR)
	se, err := CreateSession(store)
	if err != nil {
		log.Fatal(err)
	}
	mustExecute(se, "use test")
	return se
}

func prepareBenchData(se Session, colType string, valueFormat string, valueCount int) {
	mustExecute(se, "drop table if exists t")
	mustExecute(se, fmt.Sprintf("create table t (pk int primary key auto_increment, col %s, index idx (col))", colType))
	mustExecute(se, "begin")
	for i := 0; i < valueCount; i++ {
		mustExecute(se, "insert t (col) values ("+fmt.Sprintf(valueFormat, i)+")")
	}
	mustExecute(se, "commit")
}

func prepareSortBenchData(se Session, colType string, valueFormat string, valueCount int) {
	mustExecute(se, "drop table if exists t")
	mustExecute(se, fmt.Sprintf("create table t (pk int primary key auto_increment, col %s)", colType))
	mustExecute(se, "begin")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < valueCount; i++ {
		mustExecute(se, "insert t (col) values ("+fmt.Sprintf(valueFormat, r.Intn(valueCount))+")")
	}
	mustExecute(se, "commit")
}

func prepareJoinBenchData(se Session, colType string, valueFormat string, valueCount int) {
	mustExecute(se, "drop table if exists t")
	mustExecute(se, fmt.Sprintf("create table t (pk int primary key auto_increment, col %s)", colType))
	mustExecute(se, "begin")
	for i := 0; i < valueCount; i++ {
		mustExecute(se, "insert t (col) values ("+fmt.Sprintf(valueFormat, i)+")")
	}
	mustExecute(se, "commit")
}

func readResult(rs ast.RecordSet, count int) {
	for count > 0 {
		x, err := rs.Next()
		if err != nil {
			log.Fatal(err)
		}
		if x == nil {
			log.Fatal(count)
		}
		count--
	}
	rs.Close()
}

func BenchmarkBasic(b *testing.B) {
	se := prepareBenchSession()
	for i := 0; i < b.N; i++ {
		rs, err := se.Execute("select 1")
		if err != nil {
			b.Fatal(err)
		}
		readResult(rs[0], 1)
	}
}

func BenchmarkTableScan(b *testing.B) {
	b.StopTimer()
	se := prepareBenchSession()
	prepareBenchData(se, "int", "%v", smallCount)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		rs, err := se.Execute("select * from t")
		if err != nil {
			b.Fatal(err)
		}
		readResult(rs[0], smallCount)
	}
}

func BenchmarkTableLookup(b *testing.B) {
	b.StopTimer()
	se := prepareBenchSession()
	prepareBenchData(se, "int", "%d", smallCount)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		rs, err := se.Execute("select * from t where pk = 64")
		if err != nil {
			b.Fatal(err)
		}
		readResult(rs[0], 1)
	}
}

func BenchmarkStringIndexScan(b *testing.B) {
	b.StopTimer()
	se := prepareBenchSession()
	prepareBenchData(se, "varchar(255)", "'hello %d'", smallCount)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		rs, err := se.Execute("select * from t where col > 'hello'")
		if err != nil {
			b.Fatal(err)
		}
		readResult(rs[0], smallCount)
	}
}

func BenchmarkStringIndexLookup(b *testing.B) {
	b.StopTimer()
	se := prepareBenchSession()
	prepareBenchData(se, "varchar(255)", "'hello %d'", smallCount)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		rs, err := se.Execute("select * from t where col = 'hello 64'")
		if err != nil {
			b.Fatal(err)
		}
		readResult(rs[0], 1)
	}
}

func BenchmarkIntegerIndexScan(b *testing.B) {
	b.StopTimer()
	se := prepareBenchSession()
	prepareBenchData(se, "int", "%v", smallCount)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		rs, err := se.Execute("select * from t where col >= 0")
		if err != nil {
			b.Fatal(err)
		}
		readResult(rs[0], smallCount)
	}
}

func BenchmarkIntegerIndexLookup(b *testing.B) {
	b.StopTimer()
	se := prepareBenchSession()
	prepareBenchData(se, "int", "%v", smallCount)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		rs, err := se.Execute("select * from t where col = 64")
		if err != nil {
			b.Fatal(err)
		}
		readResult(rs[0], 1)
	}
}

func BenchmarkDecimalIndexScan(b *testing.B) {
	b.StopTimer()
	se := prepareBenchSession()
	prepareBenchData(se, "decimal(32,6)", "%v.1234", smallCount)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		rs, err := se.Execute("select * from t where col >= 0")
		if err != nil {
			b.Fatal(err)
		}
		readResult(rs[0], smallCount)
	}
}

func BenchmarkDecimalIndexLookup(b *testing.B) {
	b.StopTimer()
	se := prepareBenchSession()
	prepareBenchData(se, "decimal(32,6)", "%v.1234", smallCount)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		rs, err := se.Execute("select * from t where col = 64.1234")
		if err != nil {
			b.Fatal(err)
		}
		readResult(rs[0], 1)
	}
}

func BenchmarkInsertWithIndex(b *testing.B) {
	b.StopTimer()
	se := prepareBenchSession()
	mustExecute(se, "drop table if exists t")
	mustExecute(se, "create table t (pk int primary key, col int, index idx (col))")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		mustExecute(se, fmt.Sprintf("insert t values (%d, %d)", i, i))
	}
}

func BenchmarkInsertNoIndex(b *testing.B) {
	b.StopTimer()
	se := prepareBenchSession()
	mustExecute(se, "drop table if exists t")
	mustExecute(se, "create table t (pk int primary key, col int)")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		mustExecute(se, fmt.Sprintf("insert t values (%d, %d)", i, i))
	}
}

func BenchmarkSort(b *testing.B) {
	b.StopTimer()
	se := prepareBenchSession()
	prepareSortBenchData(se, "int", "%v", bigCount)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		rs, err := se.Execute("select * from t order by col limit 50")
		if err != nil {
			b.Fatal(err)
		}
		readResult(rs[0], 50)
	}
}

func BenchmarkJoin(b *testing.B) {
	b.StopTimer()
	se := prepareBenchSession()
	prepareJoinBenchData(se, "int", "%v", smallCount)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		rs, err := se.Execute("select * from t a join t b on a.col = b.col")
		if err != nil {
			b.Fatal(err)
		}
		readResult(rs[0], smallCount)
	}
}

func BenchmarkJoinLimit(b *testing.B) {
	b.StopTimer()
	se := prepareBenchSession()
	prepareJoinBenchData(se, "int", "%v", smallCount)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		rs, err := se.Execute("select * from t a join t b on a.col = b.col limit 1")
		if err != nil {
			b.Fatal(err)
		}
		readResult(rs[0], 1)
	}
}
