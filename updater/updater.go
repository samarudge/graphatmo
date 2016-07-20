package updater

import(
  "github.com/yosmudge/graphatmo/api"
  "github.com/yosmudge/graphatmo/config"
)

type Updater struct{
  Config      *config.Config
  Api         api.Api
  NoSend      bool
}

func New(c *config.Config, noSend bool) *Updater{
  u := &Updater{}
  u.Config = c

  u.Api, _ = api.Create(*c)

  u.NoSend = noSend

  return u
}
