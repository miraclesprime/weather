package storage

import (
    "sync"
    "time"

    "github.com/miraclesprime/weather/internal/weather"
)

type CityEntry struct {
    Aggregated *weather.NormalizedWeather
    History    []*weather.NormalizedWeather
    LastSuccessfulFetch time.Time
}

type Store struct {
    mu sync.RWMutex
    data map[string]*CityEntry
}

func New() *Store {
    return &Store{data: make(map[string]*CityEntry)}
}

func (s *Store) Save(city string, aggregated *weather.NormalizedWeather) {
    s.mu.Lock()
    defer s.mu.Unlock()
    e, ok := s.data[city]
    if !ok {
        e = &CityEntry{}
        s.data[city] = e
    }
    e.Aggregated = aggregated
    e.History = append(e.History, aggregated)
    e.LastSuccessfulFetch = time.Now().UTC()
}

func (s *Store) Get(city string) (*CityEntry, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    e, ok := s.data[city]
    return e, ok
}

func (s *Store) AllLastFetches() map[string]time.Time {
    s.mu.RLock()
    defer s.mu.RUnlock()
    m := make(map[string]time.Time, len(s.data))
    for k, v := range s.data {
        m[k] = v.LastSuccessfulFetch
    }
    return m
}
