package updater

import(
  "time"
  log "github.com/Sirupsen/logrus"

  "github.com/yosmudge/graphatmo/netatmo"
)

func (u *Updater) Run(){
  var err error
  //Main app loop
  for {
    u.Stations, err = u.Api.FetchStations()
    if err != nil {
      log.WithFields(log.Fields{
        "error": err,
      }).Error("Error getting stations")
      continue
    }

    for _,s := range u.Stations{
      sq := netatmo.Query{}
      metrics := s.Stats(sq, &u.Api)
      u.Graphite.SendMetrics(metrics)
    }

    waitTime := time.Minute*3
    time.Sleep(waitTime)
  }
}
