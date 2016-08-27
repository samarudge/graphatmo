package main

import(
  "os"
  "fmt"
  "os/user"
  "path/filepath"
  "github.com/voxelbrain/goptions"
  log "github.com/Sirupsen/logrus"

  "github.com/YoSmudge/graphatmo/config"
  "github.com/YoSmudge/graphatmo/netatmo"
  "github.com/YoSmudge/graphatmo/updater"
)

type options struct {
  Verbose   bool            `goptions:"-v, --verbose, description='Log verbosely'"`
  Help      goptions.Help   `goptions:"-h, --help, description='Show help'"`
  Config    string          `goptions:"-c, --config, description='Config Yaml file to use'"`
  NoSend    bool            `goptions:"-n, --no-send, description='Fetch data from Netatmo but don\\'t send to Graphite'"`
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
  log.WithFields(log.Fields{
    "configFile": fullConfigPath,
  }).Debug("Loading config file")

  if _, err := os.Stat(fullConfigPath); os.IsNotExist(err) {
    log.WithFields(log.Fields{
      "configFile": fullConfigPath,
    }).Error("Config file does not exist")
    os.Exit(1)
  }

  config := config.ParseConfig(fullConfigPath)

  if parsedOptions.Verb == "login" {
    a, _ := netatmo.Create(config)
    a.DoLogin()
  } else if parsedOptions.Verb == "" || parsedOptions.Verb == "run" {
    u := updater.New(&config, parsedOptions.NoSend)
    u.Run()
  }
}
