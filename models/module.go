package models

import(
  "../api"
  log "github.com/Sirupsen/logrus"
  "time"
  "net/url"
  "strings"
  "strconv"
  "fmt"
  "sort"
)

type Module struct {
  Name          string
  Id            string
  Station       *Station
  RfStrength    float64
  Battery       float64
  Measures      []string
  LastData      Timestamp
  DataInterval  int
}

func ModuleFromJson(moduleObj map[string]interface{}) Module{
  // Create a module from a JSON blob
  mod := Module{}

  mod.Name = moduleObj["module_name"].(string)
  mod.Id = moduleObj["_id"].(string)
  if val, ok := moduleObj["rf_status"]; ok {
    mod.RfStrength = val.(float64)
  }
  if val, ok := moduleObj["battery_vp"]; ok {
    mod.Battery = val.(float64)
  }

  mod.DataInterval = 300

  for _, measure := range moduleObj["data_type"].([]interface{}){
    mod.Measures = append(mod.Measures, measure.(string))
  }

  return mod
}

func (self *Module) ModName() string{
  return fmt.Sprintf("%s/%s", self.Station.Name, self.Name)
}

func (self *Module) NextData() time.Time{
  // Return expected next data for module
  // Netatmo API has 10 minute cache time, so update every 11 minutes
  return self.LastData.Time.Add(time.Duration(11)*time.Minute)
}

func (self *Module) Stats() []StatsSet{
  stats := []StatsSet{}

  if self.NextData().Before(time.Now()) {
    sset := NewStatsSet("meta", "module", self.Station.Name, self.Name)
    sset.AddStat("battery", self.Station.LastStatus, self.Battery)
    sset.AddStat("rf", self.Station.LastStatus, self.RfStrength)
    stats = append(stats, sset)

    dataFrom := time.Now().Add(time.Duration(3600*-1)*time.Second)

    // Now the stations actual metrics
    p := url.Values{}
    p.Set("device_id", self.Station.Id)
    p.Set("module_id", self.Id)
    p.Set("scale", "max")
    p.Set("type", strings.Join(self.Measures, ","))
    p.Set("real_time", "true")
    p.Set("optimize", "false")
    p.Set("date_begin", strconv.FormatInt(dataFrom.Unix(),10))
    p.Set("date_end", strconv.FormatInt(time.Now().Unix(),10))
    r := api.Request{
      Path:"getmeasure",
      Params:p,
    }
    err := self.Station.StationList.Api.DoCall(&r)
    if err != nil{
      log.Error(fmt.Sprintf("Error getting stats for %s/%s: %s", self.Station.Name, self.Name, err))
    } else {
      //Measurements
      data := r.Data["body"].
              (map[string]interface{})

      datapoints := self.TimeSeriesData(data)

      logFields := log.Fields{
        "DataInterval": self.DataInterval,
        "LastData": self.LastData.Timestamp,
        "LatestData": datapoints.LatestTimestamp().Timestamp,
        "Module": self.ModName(),
      }

      if self.LastData.Unix >= datapoints.LatestTimestamp().Unix {
        log.WithFields(logFields).Warning("Update due but no new data!")
      } else {
        sset := NewStatsSet("station", self.Station.Name, self.Name)
        self.TimestampStats(&sset, datapoints)

        stats = append(stats, sset)

        self.LastData = datapoints.LatestTimestamp()
        self.DataInterval = datapoints.DataInterval()
        log.WithFields(logFields).Info("Updated")
      }
    }
  }

  return stats
}

func (self *Module) TimeSeriesData(data map[string]interface{}) DataPointList{
  timestamps := []Timestamp{}
  datapoints := DataPointList{}

  for ts := range data{
    timestamps = append(timestamps, NewTimestamp(ts))
  }

  for _,ts := range timestamps{
    datapoint := DataPoint{}
    datapoint.Time = ts
    datapoint.Complete = true
    datapoint.Data = make(map[string]float64)

    rawData := data[ts.String].([]interface{})

    for i,measure := range self.Measures{
      periodData := rawData[i]
      if periodData == nil {
        datapoint.Complete = false
      } else {
        datapoint.Data[measure] = periodData.(float64)
      }
    }

    datapoints.DataPoints = append(datapoints.DataPoints, datapoint)
  }

  sort.Sort(ByTimestamp(datapoints.DataPoints))

  return datapoints
}

func (self *Module) TimestampStats(sset *StatsSet, datapoints DataPointList){
  for _,point := range datapoints.DataPoints{
    if point.Time.Unix > self.LastData.Unix {
      for name,value := range point.Data{
        sset.AddStat(name, point.Time.Float, value)
      }

      log.WithFields(log.Fields{
        "Module": self.ModName(),
        "Time": point.Time.String,
      }).Debug("Sending Data")
    }
  }
}
