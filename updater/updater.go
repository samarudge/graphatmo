package updater

import(
  "os"
  log "github.com/Sirupsen/logrus"

  "github.com/yosmudge/graphatmo/config"
  "github.com/yosmudge/graphatmo/netatmo"
  "github.com/yosmudge/graphatmo/graphite"
)

type Updater struct{
  Config      *config.Config
  Api         netatmo.Api
  NoSend      bool
  Graphite    *graphite.Graphite
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
