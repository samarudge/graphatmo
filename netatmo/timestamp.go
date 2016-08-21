package netatmo

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

func (t *Timestamp) FromString(s string) *Timestamp{
  t.String = s
  t.Unix,_ = strconv.Atoi(t.String)
  t.Timestamp = time.Unix(int64(t.Unix), 10).Format("2006-01-02 15:04:05")
  t.Time = time.Unix(int64(t.Unix), 0)
  t.Float = float64(t.Unix)
  return t
}

func (t *Timestamp) UnmarshalJSON(b []byte) error{
  t.FromString(string(b))
  return nil
}
