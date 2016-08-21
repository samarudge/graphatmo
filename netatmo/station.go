package netatmo

import(
  "fmt"
  "encoding/json"
)

type getStationsDataResponse struct{
  Body struct{
    Devices  []Station
  }
}

type Station struct{
  *Module
  Id            string      `json:"_id"`
  Name          string      `json:"station_name"`
  Modules       []Module
  WifiStatus    float64     `json:"wifi_status"`
  LastStatus    Timestamp   `json:"last_status_store"`
}

func (a *Api) FetchStations() ([]Station, error){
  var stations []Station
  stationData := getStationsDataResponse{}
  r := Request{
    Path:"getstationsdata",
  }
  err := a.DoCall(&r, &stationData)
  if err != nil{
    return stations, err
  }
  stations = stationData.Body.Devices

  for si,_ := range stations{
    s := &stations[si]
    // I'm sure there's a better way to do this
    // create station module
    var err error
    var sj []byte
    sj, err = json.Marshal(s)
    stationModule := Module{}

    if err == nil{
      err = json.Unmarshal(sj, &stationModule)
    }

    if err != nil {
      return stations, fmt.Errorf("Error remarshalling station to module: %s", err)
    }

    s.Modules = append(s.Modules, stationModule)

    for mi,_ := range s.Modules{
      m := &s.Modules[mi]
      m.Station = s
    }
  }

  return stations, err
}

func (s *Station) Stats(statsQuery Query, a *Api) []StatsSet{
  // Stats for the station
  stats := []StatsSet{}

  // Base station data
  sset := NewStatsSet("meta", "station", s.Name)
  sset.AddStat("wifi", s.LastStatus.Float, s.WifiStatus)
  stats = append(stats, sset)

  // Modules
  for i := range s.Modules{
    module := s.Modules[i]
    stats = append(stats, module.Stats(statsQuery, a)...)
  }

  return stats
}
