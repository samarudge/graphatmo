package updater

import(
  "github.com/yosmudge/graphatmo/netatmo"
  "github.com/yosmudge/graphatmo/config"
)

type Updater struct{
  Config      *config.Config
  Api         netatmo.Api
  NoSend      bool
}

func New(c *config.Config, noSend bool) *Updater{
  u := &Updater{}
  u.Config = c

  u.Api, _ = netatmo.Create(*c)

  u.NoSend = noSend

  return u
}
