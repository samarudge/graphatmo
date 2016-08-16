package updater

import(
  "time"
  log "github.com/Sirupsen/logrus"

  "github.com/yosmudge/graphatmo/models"
)

func (u *Updater) Run(){
  stns := models.StationList{Api:u.Api}

  //Main app loop
  for {
    err := stns.FetchStations()
    if err != nil {
      log.WithFields(log.Fields{
        "error": err,
      }).Error("Error getting stations")
    }

    metrics := []models.StatsSet{}

    for i := range stns.Stations{
      station := stns.Stations[i]
      stationStats := station.Stats()
      metrics = append(metrics, stationStats...)
    }

    u.Graphite.SendMetrics(metrics)

    waitTime := stns.NextData().Sub(time.Now())
    log.WithFields(log.Fields{
      "NextUpdate": stns.NextData().Format("2006-01-02 15:04:05"),
    }).Info("Waiting for next update")
    time.Sleep(waitTime)
  }
}
