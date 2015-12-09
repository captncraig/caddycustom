package stats

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/rcrowley/go-metrics"
)

type Metric struct {
	Metric string
	Tags   M
	Value  int64
}

type Statistic interface {
	GetMetrics() []*Metric
}

type base struct {
	metric string
	tags   M
}
type counter struct {
	base
	i int64
}

func (c *counter) Increment(i int64) {
	if i == 0 {
		return
	}
	atomic.AddInt64(&c.i, i)
}

func (c *counter) Value() int64 {
	return atomic.LoadInt64(&c.i)
}

func (c *counter) GetMetrics() []*Metric {
	return []*Metric{&Metric{
		Metric: c.metric,
		Tags:   c.tags,
		Value:  c.Value(),
	}}
}

type sample struct {
	base
	metrics.Sample
}

func (s *sample) GetMetrics() []*Metric {
	return []*Metric{
		&Metric{
			Metric: s.metric + "_avg",
			Tags:   s.tags,
			Value:  int64(s.Mean()),
		},
		&Metric{
			Metric: s.metric + "_95",
			Tags:   s.tags,
			Value:  int64(s.Percentile(.95)),
		},
		&Metric{
			Metric: s.metric + "_max",
			Tags:   s.tags,
			Value:  s.Max(),
		},
	}
}

func init() {
	// collect basic system stats
	Statistics.RegisterFunc(func() []*Metric {
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		return []*Metric{
			&Metric{
				Metric: "caddy_goroutines",
				Tags:   M{},
				Value:  int64(runtime.NumGoroutine()),
			},
			&Metric{
				Metric: "caddy_alloc",
				Tags:   M{},
				Value:  int64(ms.Alloc),
			},
			&Metric{
				Metric: "caddy_heap_inuse",
				Tags:   M{},
				Value:  int64(ms.HeapAlloc),
			},
			&Metric{
				Metric: "caddy_heap_objects",
				Tags:   M{},
				Value:  int64(ms.HeapObjects),
			},
			&Metric{
				Metric: "caddy_stack_inuse",
				Tags:   M{},
				Value:  int64(ms.StackInuse),
			},
			&Metric{
				Metric: "caddy_gc_pause",
				Tags:   M{},
				Value:  int64(ms.PauseTotalNs),
			},
			&Metric{
				Metric: "caddy_numgc",
				Tags:   M{},
				Value:  int64(ms.NumGC),
			},
		}
	})
}

//Statistics is the Global collection of every statistic we know about
var Statistics = &StatList{
	counters:   map[string]*counter{},
	samples:    map[string]*sample{},
	everything: []Statistic{},
}

type StatList struct {
	mu         sync.RWMutex
	counters   map[string]*counter
	samples    map[string]*sample
	everything []Statistic
}

type M map[string]string

func (s *StatList) Register(st Statistic) {
	s.mu.Lock()
	s.everything = append(s.everything, st)
	s.mu.Unlock()
}

func (s *StatList) RegisterFunc(f StatisticFunc) {
	s.Register(f)
}

type StatisticFunc func() []*Metric

func (f StatisticFunc) GetMetrics() []*Metric {
	return f()
}

func (s *StatList) Add(metric string, tags M, increment int64) {
	key := metricKey(metric, tags)
	c, ok := s.counters[key]
	if !ok {
		s.mu.Lock()
		if c, ok = s.counters[key]; !ok {
			c = &counter{
				base: base{
					metric: metric,
					tags:   tags,
				},
			}
			s.counters[key] = c
			s.everything = append(s.everything, c)
		}
		s.mu.Unlock()
	}
	c.Increment(increment)
}

func (s *StatList) Sample(metric string, tags M, value int64) {
	key := metricKey(metric, tags)
	samp, ok := s.samples[key]
	if !ok {
		s.mu.Lock()
		if samp, ok = s.samples[key]; !ok {
			samp = &sample{
				base: base{
					metric: metric,
					tags:   tags,
				},
				Sample: metrics.NewExpDecaySample(1028, 0.015),
			}
			s.samples[key] = samp
			s.everything = append(s.everything, samp)
		}
		s.mu.Unlock()
	}
	samp.Update(value)
}

func metricKey(metric string, tags M) string {
	var keys []string
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	b := &bytes.Buffer{}
	for i, k := range keys {
		if i > 0 {
			fmt.Fprint(b, ",")
		}
		fmt.Fprintf(b, "%s=%s", k, tags[k])
	}
	return fmt.Sprintf("%s{%s}", metric, b.String())
}

func (s *StatList) GetAll() []*Metric {
	all := []*Metric{}
	s.mu.RLock()
	for _, stat := range s.everything {
		m := stat.GetMetrics()
		all = append(all, m...)
	}
	s.mu.RUnlock()
	return all
}

func (s *StatList) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.GetAll())
}
