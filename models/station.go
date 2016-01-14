package models

import(
  "../api"
  log "github.com/Sirupsen/logrus"
  "fmt"
  "net/url"
  "strings"
  "time"
  "strconv"
  "sort"
)

type StationList struct {
  Api           api.Api
  Stations      []Station
}

type Station struct {
  Name          string
  Id            string
  WifiStatus    float64
  LastStatus    float64
  Modules       []Module
  StationList   *StationList
}

func FetchStations(a api.Api) (StationList, error){
  // Fetch stations and modules from Netatmo
  stns := StationList{Api:a}

  r := api.Request{Path:"getstationsdata"}
  err := a.DoCall(&r)
  if err != nil{
    return stns, err
  }

  // I don't know if this kind of multi-line thing is ok but
  //   it makes my brain hurt less when dealing with JSON
  stationResponse := r.Data["body"].
              (map[string]interface{})["devices"].
              ([]interface{})
  for _, stationJson := range stationResponse {
    stationObj := stationJson.(map[string]interface{})
    stn := Station{}

    stn.Name = stationObj["station_name"].(string)
    stn.Id = stationObj["_id"].(string)
    stn.WifiStatus = stationObj["wifi_status"].(float64)
    stn.LastStatus = stationObj["last_status_store"].(float64)

    // Station is also a module
    stn.Modules = append(stn.Modules, ModuleFromJson(stationObj))

    moduleResponse := stationObj["modules"].([]interface{})
    for _, moduleJson := range moduleResponse {
      moduleObj := moduleJson.(map[string]interface{})

      stn.Modules = append(stn.Modules, ModuleFromJson(moduleObj))
    }

    stn.StationList = &stns

    log.Debug(fmt.Sprintf("Found station %s with %d modules", stn.Name, len(stn.Modules)))
    stns.Stations = append(stns.Stations, stn)
  }

  return stns, nil
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

    if module.NextData().Before(time.Now()) {
      sset := NewStatsSet("meta", "module", self.Name, module.Name)
      sset.AddStat("battery", self.LastStatus, module.Battery)
      sset.AddStat("rf", self.LastStatus, module.RfStrength)
      stats = append(stats, sset)

      dataFrom := time.Now().Add(time.Duration(3600*-1)*time.Second)

      // Now the stations actual metrics
      p := url.Values{}
      p.Set("device_id", self.Id)
      p.Set("module_id", module.Id)
      p.Set("scale", "max")
      p.Set("type", strings.Join(module.Measures, ","))
      p.Set("real_time", "true")
      p.Set("optimize", "false")
      p.Set("date_begin", strconv.FormatInt(dataFrom.Unix(),10))
      p.Set("date_end", strconv.FormatInt(time.Now().Unix(),10))
      r := api.Request{
        Path:"getmeasure",
        Params:p,
      }
      err := self.StationList.Api.DoCall(&r)
      if err != nil{
        log.Error(fmt.Sprintf("Error getting stats for %s/%s: %s", self.Name, module.Name, err))
      } else {
        //Measurements
        data := r.Data["body"].
                (map[string]interface{})

        timestamps := []int{}

        for ts := range data{
          t,_ := strconv.Atoi(ts)
          timestamps = append(timestamps, int(t))
        }
        sort.Ints(timestamps)

        latestTimestamp := timestamps[len(timestamps)-1]

        modName := fmt.Sprintf("%s/%s", self.Name, module.Name)

        logFields := log.Fields{
          "DataInterval": module.DataInterval,
          "LastData": time.Unix(int64(module.LastData), 10).Format("2006-01-02 15:04:05"),
          "LatestData": time.Unix(int64(latestTimestamp), 10).Format("2006-01-02 15:04:05"),
          "Module": modName,
        }

        if module.LastData >= latestTimestamp {
          log.WithFields(logFields).Warning("Update due but no new data!")
        } else {
          sset := NewStatsSet("station", self.Name, module.Name)
          for _,ts := range timestamps{
            if (ts > module.LastData && module.LastData > 0) || ts == latestTimestamp {
              periodData := data[strconv.FormatInt(int64(ts),10)].([]interface{})
              for i, measure := range module.Measures{
                sset.AddStat(measure, float64(ts), periodData[i].(float64))
              }

              log.WithFields(log.Fields{
                "Module": modName,
                "Time": time.Unix(int64(ts), 10).Format("2006-01-02 15:04:05"),
              }).Debug("Sending Data")
            }
          }

          stats = append(stats, sset)

          module.LastData = latestTimestamp
          module.DataInterval = latestTimestamp-timestamps[len(timestamps)-2]
          log.WithFields(logFields).Info("Updated")
        }
      }
    }
  }

  return stats
}
