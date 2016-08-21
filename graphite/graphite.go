package graphite

import(
  grp "github.com/marpaia/graphite-golang"
  log "github.com/Sirupsen/logrus"
  "github.com/yosmudge/graphatmo/netatmo"
  "github.com/yosmudge/graphatmo/config"
  "strings"
  "strconv"
  "time"
  "fmt"
)

type Graphite struct{
  Connection  *grp.Graphite
}

func splitHost(hostString string) (string, int){
  hostParts := strings.Split(hostString, ":")
  hostPort, err := strconv.Atoi(hostParts[1])
  if err != nil{
    panic(fmt.Sprintf("Unable to parse Graphite host string", hostString))
  }

  return hostParts[0], hostPort
}

func CreateTest() *Graphite{
  g := Graphite{}
  g.Connection = grp.NewGraphiteNop("",0)
  return &g
}

func Create(config config.Config) (*Graphite, error){
  g := Graphite{}

  hostname, port := splitHost(config.Graphite)

  conn, err := grp.NewGraphite(hostname, port)
  if err != nil{
    return &g, err
  }

  g.Connection = conn
  err = g.Connection.Disconnect()
  if err != nil{
    return &g, err
  }

  return &g, nil
}

func (self *Graphite) SendMetrics(metrics []netatmo.StatsSet) error{
  err := self.Connection.Connect()
  if err != nil{
    return err
  }

  for _, metricSet := range metrics{
    for _, stat := range metricSet.Data{
      fullMetricName := strings.Join([]string{"graphatmo",metricSet.Key,stat.Key}, ".")

      metricLogData := log.Fields{
          "Name": fullMetricName,
          "Value": stat.Value,
          "Timestamp": time.Unix(stat.Timestamp, 0).Format("2006-01-02 15:04:05"),
      }
      log.WithFields(metricLogData).Debug("Sending metric")

      m := grp.NewMetric(fullMetricName, stat.Value, stat.Timestamp)

      err = self.Connection.SendMetrics([]grp.Metric{m})
      if err != nil{
        log.WithFields(metricLogData).Error("Error sending metric")
      }
    }
  }

  if ! self.Connection.IsNop() {
    err = self.Connection.Disconnect()
    if err != nil{
      return err
    }
  }

  return nil
}
