package main

import(
  "github.com/voxelbrain/goptions"
  log "github.com/Sirupsen/logrus"
  "fmt"
  "path/filepath"
  "os/user"
  "os"
  "time"
  "./config"
  "./api"
  "./models"
  "./graphite"
)

type options struct {
  Verbose   bool            `goptions:"-v, --verbose, description='Log verbosely'"`
  Help      goptions.Help   `goptions:"-h, --help, description='Show help'"`
  Config    string          `goptions:"-c, --config, description='Config Yaml file to use'"`
  Verb      goptions.Verbs
  Login     struct{}        `goptions:"login"`
  Run       struct{
    Backfill  bool          `goptions:"--backfill, description='Backfill information in Graphite'"`
  }                         `goptions:"run"`
}

func main() {
  parsedOptions := options{}
  currentUser, _ := user.Current()
  parsedOptions.Config = fmt.Sprintf("%s/.graphatmo.yml", currentUser.HomeDir)

  goptions.ParseAndFail(&parsedOptions)

  if parsedOptions.Verbose{
    log.SetLevel(log.DebugLevel)
  } else {
    log.SetLevel(log.InfoLevel)
  }

  log.SetFormatter(&log.TextFormatter{FullTimestamp:true})

  log.Debug("Logging verbosely!")

  fullConfigPath, _ := filepath.Abs(parsedOptions.Config)
  log.Debug(fmt.Sprintf("Loading config file %s", fullConfigPath))

  if _, err := os.Stat(fullConfigPath); os.IsNotExist(err) {
    log.Error(fmt.Sprintf("Config file %s does not exist!",fullConfigPath))
    os.Exit(1)
  }

  config := config.ParseConfig(fullConfigPath)

  a, err := api.Create(config)
  if err != nil{
    log.Error(fmt.Sprintf("Error connecting to Netatmo: %v", err))
    os.Exit(1)
  }

  if parsedOptions.Verb == "login" {
    a.DoLogin()
  } else if parsedOptions.Verb == "" || parsedOptions.Verb == "run" {
    g, err := graphite.Create(config)
    if err != nil{
      log.Error(fmt.Sprintf("Error connecting to Graphite: %v", err))
      os.Exit(1)
    }

    stationList, err := models.FetchStations(a)

    //Main app loop
    for {
      if err != nil {
        log.Error(fmt.Sprintf("Error getting stations: %v", err))
        os.Exit(1)
      }

      metrics := []models.StatsSet{}

      for i := range stationList.Stations{
        station := &stationList.Stations[i]
        stationStats := station.Stats()
        metrics = append(metrics, stationStats...)
      }

      g.SendMetrics(metrics)

      waitTime := stationList.NextData().Sub(time.Now())
      log.WithFields(log.Fields{
        "NextUpdate": stationList.NextData().Format("2006-01-02 15:04:05"),
      }).Info("Waiting for next update")
      time.Sleep(waitTime)
    }
  }
}
