package models

import(
  "../api"
  log "github.com/Sirupsen/logrus"
  "time"
)

type StationList struct {
  Api           api.Api
  Stations      []*Station
  LastUpdate    time.Time
}

type Station struct {
  Name          string
  Id            string
  WifiStatus    float64
  LastStatus    float64
  Modules       []Module
  StationList   *StationList
}

func (self *StationList) FetchStations() error{
  // Fetch stations and modules from Netatmo
  if self.LastUpdate.Before(time.Now().Add(time.Duration(time.Minute*-30))) {
    r := api.Request{Path:"getstationsdata"}
    err := self.Api.DoCall(&r)
    if err != nil{
      return err
    }

    // I don't know if this kind of multi-line thing is ok but
    //   it makes my brain hurt less when dealing with JSON
    stationResponse := r.Data["body"].
                (map[string]interface{})["devices"].
                ([]interface{})

    for _, stationJson := range stationResponse {
      stationObj := stationJson.(map[string]interface{})
      stnId := stationObj["_id"].(string)

      stn := &Station{}
      for i := range self.Stations{
        s := self.Stations[i]
        if s.Id == stnId {
          stn = s
        }
      }

      if stn.Id == "" {
        stn := Station{}
        stn.Name = stationObj["station_name"].(string)
        stn.Id = stationObj["_id"].(string)
        self.Stations = append(self.Stations, &stn)
        stn.StationList = self

        // Station is also a module
        mod := ModuleFromJson(stationObj)
        mod.Station = &stn
        stn.Modules = append(stn.Modules, mod)

        moduleResponse := stationObj["modules"].([]interface{})
        for _, moduleJson := range moduleResponse {
          moduleObj := moduleJson.(map[string]interface{})

          mod := ModuleFromJson(moduleObj)
          mod.Station = &stn
          stn.Modules = append(stn.Modules, mod)
        }

        log.WithFields(log.Fields{
          "StationName": stn.Name,
          "ModuleCount": len(stn.Modules),
        }).Debug("Found new station")
      }

      stn.WifiStatus = stationObj["wifi_status"].(float64)
      stn.LastStatus = stationObj["last_status_store"].(float64)
    }

    self.LastUpdate = time.Now()
  }

  return nil
}

func (self *StationList) NextData() time.Time{
  // Return how long until the next datapoint
  nextPoint := time.Now().Add(time.Duration(5)*time.Minute)

  for _, stn := range self.Stations{
    for _, mod := range stn.Modules{
      nextDataForModule := mod.NextData()
      if nextDataForModule.Before(nextPoint) {
        nextPoint = nextDataForModule
      }
    }
  }

  if nextPoint.Before(time.Now()){
    nextPoint = time.Now().Add(time.Duration(1)*time.Minute)
  }

  return nextPoint
}

func (self *Station) Stats() []StatsSet{
  // Stats for the station
  stats := []StatsSet{}

  // Base station data
  sset := NewStatsSet("meta", "station", self.Name)
  sset.AddStat("wifi", self.LastStatus, self.WifiStatus)
  stats = append(stats, sset)

  // Modules
  for i := range self.Modules{
    module := &self.Modules[i]

    stats = append(stats, module.Stats()...)
  }

  return stats
}
