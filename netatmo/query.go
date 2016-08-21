package netatmo

import(
  "time"
  "sort"
  "strings"
  "strconv"
  "net/url"
)

type Query struct{
  Start         time.Time
  End           time.Time
  Limit         int64
  OnlyMeasures  []string
}

func (q *Query) TargetMeasures(measures []string) []string{
  var targetMeasures []string
  if len(q.OnlyMeasures) == 0{
    targetMeasures = measures
  } else {
    for _,m := range q.OnlyMeasures{
      var found bool
      for _,mm := range measures{
        if mm == m{
          found = true
          break
        }
      }

      if found{
        targetMeasures = append(targetMeasures, m)
      }
    }
  }
  sort.Strings(targetMeasures)
  return targetMeasures
}

func (q *Query) BaseQuery(deviceId string, moduleId string, measures []string) url.Values{
  p := url.Values{}

  if !q.Start.IsZero(){
    p.Set("date_begin", strconv.FormatInt(q.Start.Unix(),10))
  }

  if !q.End.IsZero(){
    p.Set("date_end", strconv.FormatInt(q.End.Unix(),10))
  }

  var max int64
  if q.Limit == 0{
    max = 1
  } else if q.Limit > 1024{
    max = 1024
  } else {
    max = q.Limit
  }

  p.Set("limit", strconv.FormatInt(max, 10))

  p.Set("device_id", deviceId)
  p.Set("module_id", moduleId)
  p.Set("scale", "max")
  p.Set("type", strings.Join(q.TargetMeasures(measures), ","))
  p.Set("real_time", "true")
  p.Set("optimize", "false")

  return p
}
