package models

import(
  "regexp"
  "strings"
  "fmt"
)

type StatsSet struct{
  Key     string
  Data    []Stat
}

type Stat struct{
  Key         string
  Timestamp   int64
  Value       string
}

func NewStatsSet(keys ...string) StatsSet{
  set := StatsSet{}
  set.Key = set.SanitizeName(keys)
  return set
}

func (self *StatsSet) SanitizeName(keys []string) string{
  // Make a string Graphite safe
  sanitizedKeys := []string{}
  re := regexp.MustCompile("[^a-zA-Z0-9]")

  for _,key := range keys{
    sanitizedKey := re.ReplaceAllString(key, "")
    sanitizedKeys = append(sanitizedKeys, sanitizedKey)
  }

  return strings.Join(sanitizedKeys, ".")
}

func (self *StatsSet) AddStat(key string, time float64, value float64){
  // Add a stat value
  s := Stat{}
  s.Key = self.SanitizeName([]string{key})
  s.Timestamp = int64(time)
  s.Value = fmt.Sprintf("%.1f", value)
  self.Data = append(self.Data, s)
}
