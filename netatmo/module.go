package netatmo

import(
  "time"
  "strconv"
  "net/url"
  log "github.com/Sirupsen/logrus"
)

type Module struct{
  Id            string      `json:"_id"`
  Type          string
  Name          string      `json:"module_name"`
  Battery       float64     `json:"battery_percent"`
  LastStatus    Timestamp   `json:"last_status_store"`
  RfStrength    float64     `json:"rf_status"`
  Measures      []string    `json:"data_type"`
  Station       *Station
}

func (m *Module) StatsBetween(startTime time.Time, endTime time.Time, a *Api) []StatsSet{
  stats := []StatsSet{}

  sset := NewStatsSet("meta", "module", m.Station.Name, m.Name)
  sset.AddStat("battery", m.Station.LastStatus.Float, m.Battery)
  sset.AddStat("rf", m.Station.LastStatus.Float, m.RfStrength)
  stats = append(stats, sset)

  baseQuery := url.Values{}

  if !startTime.IsZero(){
    baseQuery.Set("date_begin", strconv.FormatInt(startTime.Unix(),10))
  }

  if !endTime.IsZero(){
    baseQuery.Set("date_end", strconv.FormatInt(endTime.Unix(),10))
  }

  var max int64 = 1000
  if startTime.IsZero() && endTime.IsZero(){
    max = 1
  }

  baseQuery.Set("limit", strconv.FormatInt(max, 10))

  baseQuery.Set("device_id", m.Station.Id)
  baseQuery.Set("module_id", m.Id)
  baseQuery.Set("scale", "max")
  baseQuery.Set("real_time", "true")
  baseQuery.Set("optimize", "false")

  stationDataSet := NewStatsSet("station", m.Station.Name, m.Name)

  for _,measure := range m.Measures{
    var rsp MeasureResponse

    q := baseQuery
    q.Set("type", measure)

    r := Request{
      Path:"getmeasure",
      Params:q,
    }

    err := a.DoCall(&r, &rsp)
    if err != nil{
      log.WithFields(log.Fields{
        "module": m.Name,
        "station": m.Station.Name,
        "error": err,
        "query": q.Encode(),
        "measure": measure,
      }).Warning("Error getting statistics for module")
      return stats
    }

    for time,measures := range rsp.Body{
      ts := Timestamp{}
      ts.FromString(time)
      stationDataSet.AddStat(measure, ts.Float, measures[0])
    }
  }

  stats = append(stats, stationDataSet)

  return stats
}
