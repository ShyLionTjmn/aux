package aux

import (
  "strings"
  "strconv"
  "sync"
  "time"
  "runtime"
  "fmt"
  "github.com/gomodule/redigo/redis"
)

func SplitByNum(s string) []interface{} {
  ret := make([]interface{}, 0)
  cur := int(0)
  for len(s[cur:]) > 0 {
    numpos := strings.IndexAny(s[cur:], "0123456789")
    if numpos == 0 {
      nnpos := 0
      for len(s[cur+nnpos:]) > 0 && strings.Index("0123456789", s[cur+nnpos:cur+nnpos+1]) >= 0 {
        nnpos++
      }
      ival, _ := strconv.ParseInt(s[cur:cur+nnpos], 10, 64)
      ret = append(ret, ival)
      cur += nnpos
    } else if numpos > 0 {
      ret = append(ret, s[cur:cur+numpos])
      cur += numpos
    } else {
      ret = append(ret, s[cur:])
      break
    }
  }
  return ret
}

type ByNum []string

func (a ByNum) Len() int		{ return len(a) }
func (a ByNum) Swap(i, j int)		{ a[i], a[j] = a[j], a[i] }
func (a ByNum) Less(i, j int) bool {
  aa := SplitByNum(a[i])
  ba := SplitByNum(a[j])

  alen := len(aa)
  blen := len(ba)

  mlen := alen
  if blen > alen { mlen=blen }

  for idx := 0; idx < mlen; idx++ {
    if idx >= alen {
      return true
    } else if idx >= blen {
      return false
    }

    switch aa[idx].(type) {
    case int64:

      switch ba[idx].(type) {
      case int64:
        if aa[idx].(int64) != ba[idx].(int64) {
          return aa[idx].(int64) < ba[idx].(int64)
        }
      case string:
        return true
      }

    case string:

      switch ba[idx].(type) {
      case int64:
        return false
      case string:
        if aa[idx].(string) != ba[idx].(string) {
          return strings.Compare(aa[idx].(string), ba[idx].(string)) < 0
        }
      }
    }
  }
  return true
}

type StrByNum []string

func (a StrByNum) Len() int		{ return len(a) }
func (a StrByNum) Swap(i, j int)		{ a[i], a[j] = a[j], a[i] }
func (a StrByNum) Less(i, j int) bool {
  ai, _ := strconv.Atoi(a[i])
  aj, _ := strconv.Atoi(a[j])
  return ai < aj
}

type M map[string]interface{}

func (m M) e(k string) bool {
  _, ret := m[k]
  return ret
}

func (m M) Evu(k ... string) bool {
  if len(k) == 0 { return false }

  if !m.e(k[0]) { return false }
  switch m[k[0]].(type) {
  case M:
    return m[k[0]].(M).Evu(k[1:]...)
  case uint64:
    return len(k) == 1
  default:
    return false
  }
}

func (m M) Evi(k ... string) bool {
  if len(k) == 0 { return false }

  if !m.e(k[0]) { return false }
  switch m[k[0]].(type) {
  case M:
    return m[k[0]].(M).Evi(k[1:]...)
  case int64:
    return len(k) == 1
  default:
    return false
  }
}

func (m M) Evs(k ... string) bool {
  if len(k) == 0 { return false }

  if !m.e(k[0]) { return false }
  switch m[k[0]].(type) {
  case M:
    return m[k[0]].(M).Evs(k[1:]...)
  case string:
    return len(k) == 1
  default:
    return false
  }
}

func (m M) EvM(k ... string) bool {
  if len(k) == 0 { return false }
  if !m.e(k[0]) {
    return false
  }
  switch m[k[0]].(type) {
  case M:
    if len(k) == 1 {
      return true
    } else {
      return m[k[0]].(M).EvM(k[1:]...)
    }
  default:
    return false
  }
}

func (m M) EvA(k ... string) bool {
  if len(k) == 0 { return false }
  if !m.e(k[0]) {
    return false
  }
  switch m[k[0]].(type) {
  case M:
    if len(k) == 1 {
      return false
    } else {
      return m[k[0]].(M).EvA(k[1:]...)
    }
  default:
    return len(k) == 1
  }
}

func (m M) MkM(k ... string) M {
  if len(k) == 0 { return nil }
  if !m.e(k[0]) {
    m[k[0]] = make(M)
  } else if !m.EvM(k[0]) {
    return nil //key exists and NOT hash
  }
  if len(k) == 1 { return m[k[0]].(M) }

  return m[k[0]].(M).MkM(k[1:]...)
}

const INT64_MIN int64= -9223372036854775808
const INT64_ERR =INT64_MIN
const INT64_MAX int64= 9223372036854775807
const INT64_MAXu uint64= 9223372036854775807

const UINT64_MAX uint64= 18446744073709551615
const UINT64_ERR = UINT64_MAX

const STRING_ERROR = "M.vs.error"

func (m M) Vi(k ... string) int64 {
  if len(k) == 0 { return INT64_ERR }
  if !m.e(k[0]) { return INT64_ERR }

  switch m[k[0]].(type) {
  case M:
    if len(k) == 1 { return INT64_ERR }
    return m[k[0]].(M).Vi(k[1:]...)
  case int64:
    if len(k) != 1 { return INT64_ERR }
    return m[k[0]].(int64)
  case uint64:
    if len(k) != 1 || m[k[0]].(uint64) > INT64_MAXu { return INT64_ERR }
    return int64(m[k[0]].(uint64))
  case string:
    if len(k) != 1 { return INT64_ERR }
    ret, err := strconv.ParseInt(m[k[0]].(string), 10, 64)
    if err != nil { return INT64_ERR }
    return ret
  default:
    return INT64_ERR
  }
}

func (m M) Vu(k ... string) uint64 {
  if len(k) == 0 { return UINT64_ERR }
  if !m.e(k[0]) { return UINT64_ERR }

  switch m[k[0]].(type) {
  case M:
    if len(k) == 1 { return UINT64_ERR }
    return m[k[0]].(M).Vu(k[1:]...)
  case uint64:
    if len(k) != 1 { return UINT64_ERR }
    return m[k[0]].(uint64)
  case int64:
    if len(k) != 1 || m[k[0]].(int64) < 0 { return UINT64_ERR }
    return uint64(m[k[0]].(int64))
  case string:
    if len(k) != 1 { return UINT64_ERR }
    ret, err := strconv.ParseUint(m[k[0]].(string), 10, 64)
    if err != nil { return UINT64_ERR }
    return ret
  default:
    return UINT64_ERR
  }
}

func (m M) Vs(k ... string) string {
  if len(k) == 0 { return STRING_ERROR }
  if !m.e(k[0]) { return STRING_ERROR }

  switch m[k[0]].(type) {
  case M:
    if len(k) == 1 { return STRING_ERROR }
    return m[k[0]].(M).Vs(k[1:]...)
  case uint64:
    if len(k) != 1 { return STRING_ERROR }
    return strconv.FormatUint(m[k[0]].(uint64), 10)
  case int64:
    if len(k) != 1 { return STRING_ERROR }
    return strconv.FormatInt(m[k[0]].(int64), 10)
  case string:
    if len(k) != 1 { return STRING_ERROR }
    return m[k[0]].(string)
  default:
    return STRING_ERROR
  }
}

func (m M) VM(k ... string) M {
  if len(k) == 0 { return m }
  if !m.e(k[0]) { return nil }

  switch m[k[0]].(type) {
  case M:
    if len(k) == 1 { return m[k[0]].(M) }
    return m[k[0]].(M).VM(k[1:]...)
  default:
    return nil
  }
}

func (m M) VA(k ... string) interface{} {
  if len(k) == 0 { return nil }
  if !m.e(k[0]) { return nil }

  switch m[k[0]].(type) {
  case M:
    if len(k) == 1 { return m[k[0]] }
    return m[k[0]].(M).VA(k[1:]...)
  default:
    if len(k) == 1 { return m[k[0]] }
    return nil
  }
}

func IsHexNumber(s string) bool {
  if len(s) == 0 { return false }
  for c := 0; c < len(s); c++ {
    if strings.Index("0123456789abcdefABCDEF", s[c:c+1]) < 0 {
      return false
    }
  }
  return true
}

func IsNumber(s string) bool {
  if len(s) == 0 { return false }
  for c := 0; c < len(s); c++ {
    if strings.Index("0123456789", s[c:c+1]) < 0 {
      return false
    }
  }
  return true
}

func WaitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
  c := make(chan struct{})
  go func() {
    defer close(c)
    wg.Wait()
  }()

  select {
    case <-c:
      return false // completed normally
    case <-time.After(timeout):
      return true // timed out
  }
}

func RedisCheck(r redis.Conn, red_sock_type, red_sock, red_db string) (redis.Conn, error) {
  var err error
  var red = r

  if r != nil && r.Err() == nil {
    _, err = r.Do("SELECT", red_db)
    if err != nil {
      r.Close()
      red = nil
    } else {
      return red, nil
    }
  }

  err = nil

  if red == nil {
    red, err = redis.Dial(red_sock_type, red_sock)
  }

  if err != nil { return nil, err }

  _, err = red.Do("SELECT", red_db)
  if err != nil {
    red.Close()
    red = nil
  }

  return red, err
}

func GetMemUsage() string {
  var m runtime.MemStats
  runtime.ReadMemStats(&m)
  // For info on each, see: https://golang.org/pkg/runtime/#MemStats
  return fmt.Sprintf("Alloc = %v KiB\tTotalAlloc = %v KiB\tSys = %v KiB\tNumGC = %v", BToKb(m.Alloc), BToKb(m.TotalAlloc), BToKb(m.Sys), m.NumGC)
}

func BToKb(b uint64) uint64 {
  return b / 1024
}
