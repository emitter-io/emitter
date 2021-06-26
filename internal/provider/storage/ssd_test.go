/**********************************************************************************
* Copyright (c) 2009-2019 Misakai Ltd.
* This program is free software: you can redistribute it and/or modify it under the
* terms of the GNU Affero General Public License as published by the  Free Software
* Foundation, either version 3 of the License, or(at your option) any later version.
*
* This program is distributed  in the hope that it  will be useful, but WITHOUT ANY
* WARRANTY;  without even  the implied warranty of MERCHANTABILITY or FITNESS FOR A
* PARTICULAR PURPOSE.  See the GNU Affero General Public License  for  more details.
*
* You should have  received a copy  of the  GNU Affero General Public License along
* with this program. If not, see<http://www.gnu.org/licenses/>.
************************************************************************************/
package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/emitter-io/emitter/internal/message"
	"github.com/emitter-io/stats"
	"github.com/kelindar/binary"
	"github.com/stretchr/testify/assert"
)

func getNTestMessages(count int) (frame message.Frame) {
	for i := 0; i < count; i++ {
		id := message.NewID(message.Ssid{0, uint32(i / 2), 2, uint32(i)})
		frame = append(frame, message.Message{
			ID:      id,
			Channel: []byte("test/channel/"),
			Payload: []byte(fmt.Sprintf("%v,%v,%v", 1, 2, i)),
			TTL:     100000,
		})
	}
	return
}

// Opens an NewSSD and runs a a test on it.
func runSSDTest(test func(store *SSD)) {

	// Prepare a store
	dir, _ := ioutil.TempDir("", "emitter")
	store := NewSSD(nil)
	store.Configure(map[string]interface{}{
		"dir": dir,
	})

	// Close once we're done and delete data
	defer os.RemoveAll(dir)
	defer store.Close()

	test(store)
}

func TestSSD_Store(t *testing.T) {
	runSSDTest(func(store *SSD) {
		err := store.storeFrame(getNTestMessages(10))
		assert.NoError(t, err)
	})
}

func TestSSD_Query(t *testing.T) {
	runSSDTest(func(store *SSD) {
		err := store.storeFrame(getNTestMessages(10))
		assert.NoError(t, err)

		zero := time.Unix(0, 0)
		f, err := store.Query([]uint32{0, 3, 2, 6}, zero, zero, 5)
		assert.NoError(t, err)
		assert.Len(t, f, 1)
	})
}

func TestSSD_QueryOrdered(t *testing.T) {
	runSSDTest(func(store *SSD) {
		testOrder(t, store)
	})
}

func TestSSD_QueryRetained(t *testing.T) {
	runSSDTest(func(store *SSD) {
		testRetained(t, store)
	})
}

func TestSSD_QueryRange(t *testing.T) {
	runSSDTest(func(store *SSD) {
		testRange(t, store)
	})
}

func TestSSD_QuerySurveyed(t *testing.T) {
	runSSDTest(func(s *SSD) {
		const wildcard = uint32(1815237614)
		msgs := getNTestMessages(10)
		s.storeFrame(msgs)
		zero := time.Unix(0, 0)
		tests := []struct {
			query    []uint32
			limit    int
			count    int
			gathered []byte
		}{
			{query: []uint32{0, 3, 2, 7}, limit: 10, count: 1},
			{query: []uint32{0, 1}, limit: 5, count: 2},
			{query: []uint32{0, 1}, limit: 5, count: 5, gathered: msgs.Encode()},
		}

		for _, tc := range tests {
			if tc.gathered == nil {
				s.survey = nil
			} else {
				s.survey = survey(func(string, []byte) (message.Awaiter, error) {
					return &mockAwaiter{f: func(_ time.Duration) [][]byte { return [][]byte{tc.gathered} }}, nil
				})
			}

			out, err := s.Query(tc.query, zero, zero, tc.limit)
			assert.NoError(t, err)
			count := 0
			for range out {
				count++
			}
			assert.Equal(t, tc.count, count)
		}
	})
}

func TestSSD_OnSurvey(t *testing.T) {
	runSSDTest(func(s *SSD) {
		s.storeFrame(getNTestMessages(10))
		zero := time.Unix(0, 0)
		tests := []struct {
			name        string
			query       lookupQuery
			expectOk    bool
			expectCount int
		}{
			{name: "dummy"},
			{name: "ssdstore"},
			{
				name:        "ssdstore",
				query:       newLookupQuery(message.Ssid{0, 1}, zero, zero, 1),
				expectOk:    true,
				expectCount: 1,
			},
			{
				name:        "ssdstore",
				query:       newLookupQuery(message.Ssid{0, 1}, zero, zero, 10),
				expectOk:    true,
				expectCount: 2,
			},
		}

		for _, tc := range tests {
			q, _ := binary.Marshal(tc.query)
			resp, ok := s.OnSurvey(tc.name, q)
			assert.Equal(t, tc.expectOk, ok)
			if tc.expectOk && ok {
				msgs, err := message.DecodeFrame(resp)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectCount, len(msgs))
			}
		}

		// Special, wrong payload case
		_, ok := s.OnSurvey("ssdstore", []byte{})
		assert.Equal(t, false, ok)
	})

}

// batch=1  	batch/s=89849	msg/s=89849
// batch=10  	batch/s=26015	msg/s=260149
// batch=100  	batch/s=4583	msg/s=458328
// batch=1000  	batch/s=430		msg/s=429714
// batch=10000  batch/s=36		msg/s=359491
func BenchmarkStore(b *testing.B) {
	runSSDTest(func(store *SSD) {
		benchmarkStoreSingle(b, store, 1)
		benchmarkStoreSingle(b, store, 10)
		benchmarkStoreSingle(b, store, 100)
		benchmarkStoreSingle(b, store, 1000)
		benchmarkStoreSingle(b, store, 10000)
	})
}

//batch=1  		batch/s=179990	msg/s=179990
//batch=10  	batch/s=51094	msg/s=510942
//batch=100  	batch/s=6606	msg/s=660574
//batch=1000  	batch/s=552		msg/s=551637
//batch=10000  	batch/s=50		msg/s=501079
func BenchmarkStoreParallel(b *testing.B) {
	runSSDTest(func(store *SSD) {
		benchmarkStoreParallel(b, store, 1, runtime.NumCPU())
		benchmarkStoreParallel(b, store, 10, runtime.NumCPU())
		benchmarkStoreParallel(b, store, 100, runtime.NumCPU())
		benchmarkStoreParallel(b, store, 1000, runtime.NumCPU())
		benchmarkStoreParallel(b, store, 10000, runtime.NumCPU())
	})
}

func benchmarkStoreParallel(b *testing.B, store *SSD, batchSize int, proceses int) {
	var wg sync.WaitGroup
	wg.Add(proceses)
	m := stats.NewMetric("elapsed")
	for i := 0; i < proceses; i++ {
		go func() {
			benchmarkStore(b, store, batchSize, m)
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Printf("batch=%v  \tbatch/s=%.0f\tmsg/s=%.0f \n", batchSize, m.Rate(), m.Rate()*float64(batchSize))
}

func benchmarkStoreSingle(b *testing.B, store *SSD, batchSize int) {
	m := stats.NewMetric("elapsed")
	benchmarkStore(b, store, batchSize, m)
	fmt.Printf("batch=%v  \tbatch/s=%.0f\tmsg/s=%.0f \n", batchSize, m.Rate(), m.Rate()*float64(batchSize))
}

func benchmarkStore(b *testing.B, store *SSD, batchSize int, m *stats.Metric) {
	batch := getNTestMessages(batchSize)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			default:
				store.storeFrame(batch)
				m.Update(int32(batchSize))
			}
		}
	}()

	time.Sleep(2 * time.Second)
}

// last=1		query/s=188057
// last=10  	query/s=68843
// last=100  	query/s=10547
// last=1000  	query/s=1121
func BenchmarkQuery(b *testing.B) {
	runSSDTest(func(store *SSD) {
		batchSize := 10
		for i := int64(0); i < 100000; i++ {
			store.storeFrame(getNTestMessages(batchSize))
		}

		benchmarkQuerySingle(b, store, 1)
		benchmarkQuerySingle(b, store, 10)
		benchmarkQuerySingle(b, store, 100)
		benchmarkQuerySingle(b, store, 1000)
	})
}

// last=1  		query/s=882693
// last=10  	query/s=482933
// last=100  	query/s=82463
// last=1000  	query/s=8711
func BenchmarkQueryParallel(b *testing.B) {
	runSSDTest(func(store *SSD) {
		batchSize := 10
		for i := int64(0); i < 100000; i++ {
			store.storeFrame(getNTestMessages(batchSize))
		}

		benchmarkQueryParallel(b, store, 1, runtime.NumCPU())
		benchmarkQueryParallel(b, store, 10, runtime.NumCPU())
		benchmarkQueryParallel(b, store, 100, runtime.NumCPU())
		benchmarkQueryParallel(b, store, 1000, runtime.NumCPU())
	})
}

func benchmarkQueryParallel(b *testing.B, store *SSD, last int, proceses int) {
	var wg sync.WaitGroup
	wg.Add(proceses)
	m := stats.NewMetric("elapsed")
	for i := 0; i < proceses; i++ {
		go func() {
			benchmarkQuery(b, store, last, m)
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Printf("last=%v  \tquery/s=%.0f \n", last, m.Rate())
}

func benchmarkQuerySingle(b *testing.B, store *SSD, last int) {
	m := stats.NewMetric("elapsed")
	benchmarkQuery(b, store, last, m)
	fmt.Printf("last=%v  \tquery/s=%.0f \n", last, m.Rate())
}

func benchmarkQuery(b *testing.B, store *SSD, last int, m *stats.Metric) {
	t0 := time.Unix(0, 0)
	t1 := time.Unix(time.Now().Unix(), 0)

	ssid := []uint32{0, 3, 2, 6}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			default:
				store.Query(ssid, t0, t1, last)
				m.Update(int32(last))
			}
		}
	}()

	time.Sleep(5 * time.Second)
}
