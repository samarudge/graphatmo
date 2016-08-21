package updater

import(
  "os"
  log "github.com/Sirupsen/logrus"

  "github.com/YoSmudge/graphatmo/config"
  "github.com/YoSmudge/graphatmo/netatmo"
  "github.com/YoSmudge/graphatmo/graphite"
)

type Updater struct{
  Config      *config.Config
  Api         netatmo.Api
  NoSend      bool
  Graphite    *graphite.Graphite
  Stations    []netatmo.Station
}

func New(c *config.Config, noSend bool) *Updater{
  var err error
  u := &Updater{}
  u.Config = c
  u.Api, _ = netatmo.Create(*c)
  u.NoSend = noSend

  u.Graphite = &graphite.Graphite{}
  if u.NoSend {
    u.Graphite = graphite.CreateTest()
  } else {
    u.Graphite, err = graphite.Create(*u.Config)

    if err != nil{
      log.WithFields(log.Fields{
        "error": err,
      }).Error("Error connecting to Graphite")
      os.Exit(1)
    }
  }

  return u
}
