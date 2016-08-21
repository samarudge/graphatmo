package netatmo

import(
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

func (m *Module) Stats(statsQuery Query, a *Api) []StatsSet{
  stats := []StatsSet{}

  sset := NewStatsSet("meta", "module", m.Station.Name, m.Name)
  sset.AddStat("battery", m.Station.LastStatus.Float, m.Battery)
  sset.AddStat("rf", m.Station.LastStatus.Float, m.RfStrength)
  stats = append(stats, sset)

  q := statsQuery.BaseQuery(m.Station.Id, m.Id, m.Measures)

  r := Request{
    Path:"getmeasure",
    Params:q,
  }

  var rsp MeasureResponse

  err := a.DoCall(&r, &rsp)
  if err != nil{
    log.WithFields(log.Fields{
      "module": m.Name,
      "station": m.Station.Name,
      "error": err,
      "query": q.Encode(),
    }).Warning("Error getting statistics for module")
    return stats
  }

  stationDataSet := NewStatsSet("station", m.Station.Name, m.Name)

  for time,measures := range rsp.Body{
    ts := Timestamp{}
    ts.FromString(time)

    for key,measure := range statsQuery.TargetMeasures(m.Measures){
      val := measures[key]
      stationDataSet.AddStat(measure, ts.Float, val)
    }
  }

  stats = append(stats, stationDataSet)

  return stats
}
