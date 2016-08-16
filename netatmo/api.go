package netatmo

import(
  "fmt"
  "os/user"
  "net/http"
  "golang.org/x/oauth2"

  "github.com/yosmudge/graphatmo/config"
)

type Api struct {
  Auth            config.Auth
  AuthFile        string
  BaseHost        string
  Endpoint        oauth2.Endpoint
  Config          oauth2.Config
  Client          *http.Client
}

func Create(config config.Config) (Api,error){
  a := Api{}
  a.BaseHost = "https://api.netatmo.net"
  a.Auth = config.Auth

  currentUser, _ := user.Current()
  a.AuthFile = fmt.Sprintf("%s/.graphatmo-auth.yml", currentUser.HomeDir)

  a.Endpoint = oauth2.Endpoint{
    AuthURL:  fmt.Sprintf("%s/oauth2/token", a.BaseHost),
    TokenURL: fmt.Sprintf("%s/oauth2/token", a.BaseHost),
  }

  a.Config = oauth2.Config{
    ClientID:     a.Auth.ClientId,
    ClientSecret: a.Auth.ClientSecret,
    Scopes:       []string{"read_station"},
    Endpoint:     a.Endpoint,
  }

  return a, nil
}

func (self *Api) PrepareClient() error{
  // Ensure client is loaded and token is refreshed
  token, err:= self.LoadToken()
  if err != nil {
    return err
  }

  if self.Client == nil {
    self.Client = self.Config.Client(oauth2.NoContext, &token)
  }

  return nil
}
