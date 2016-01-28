package models

import(
  "time"
  "strconv"
)

type Timestamp struct{
  Unix        int
  String      string
  Timestamp   string
  Float       float64
  Time        time.Time
}

func NewTimestamp(timestring string) Timestamp{
  t := Timestamp{}
  t.String = timestring
  t.Unix,_ = strconv.Atoi(timestring)
  t.Timestamp = time.Unix(int64(t.Unix), 10).Format("2006-01-02 15:04:05")
  t.Time = time.Unix(int64(t.Unix), 0)
  t.Float = float64(t.Unix)
  return t
}
