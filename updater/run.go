package updater

import(
  "os"
  "time"
  log "github.com/Sirupsen/logrus"

  "github.com/yosmudge/graphatmo/graphite"
  "github.com/yosmudge/graphatmo/models"
)

func (u *Updater) Run(){
  var err error
  g := graphite.Graphite{}
  if u.NoSend {
    g = graphite.CreateTest()
  } else {
    g, err = graphite.Create(*u.Config)

    if err != nil{
      log.WithFields(log.Fields{
        "error": err,
      }).Error("Error connecting to Graphite")
      os.Exit(1)
    }
  }

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

    g.SendMetrics(metrics)

    waitTime := stns.NextData().Sub(time.Now())
    log.WithFields(log.Fields{
      "NextUpdate": stns.NextData().Format("2006-01-02 15:04:05"),
    }).Info("Waiting for next update")
    time.Sleep(waitTime)
  }
}
