package aux

import (
  "sync"
  "time"
  "os"
  "fmt"
)

type someState struct {
  mutex		*sync.Mutex
  good		time.Time
  bad		time.Time
  name		string
}

func NewSomeState(name string) *someState {
  return &someState{mutex: &sync.Mutex{}, name: name}
}

func (s someState) State(ok bool, err error) {
  s.mutex.Lock()
  defer s.mutex.Unlock()
  if ok {
    if s.good.Before(s.bad) {
      s.good = time.Now()
      fmt.Fprintln(os.Stderr, s.name, "is back")
    }
  } else {
    if s.good.After(s.bad) || s.good.Equal(s.bad) {
      s.bad = time.Now()
      fmt.Fprint(os.Stderr, s.name, "is down")
      if err != nil {
        fmt.Fprint(os.Stderr, ":", err)
      }
     fmt.Fprintln(os.Stderr)
    }
  }
}
