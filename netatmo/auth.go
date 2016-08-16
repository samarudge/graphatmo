package netatmo

import(
  "os"
  "fmt"
  "io/ioutil"
  "golang.org/x/oauth2"
  yaml "github.com/go-yaml/yaml"
  log "github.com/Sirupsen/logrus"
)

func (self *Api) SaveToken(token *oauth2.Token) error {
  // Save auth data

  log.WithFields(log.Fields{"authFile":self.AuthFile}).Debug("Saving auth token")

  yml, err := yaml.Marshal(&token)
  if err != nil {
    return err
  }

  err = ioutil.WriteFile(self.AuthFile, yml, 0600)
  if err != nil {
    return err
  }

  return nil
}

func (self *Api) LoadToken() (oauth2.Token, error){
  var token oauth2.Token
  if _, err := os.Stat(self.AuthFile); os.IsNotExist(err) {
    return token, fmt.Errorf("The authentication file doesn't exist! Maybe you need to run `graphatmo login`?")
  }

  authContent, err := ioutil.ReadFile(self.AuthFile)
  if err != nil {
    return token, err
  }

  token = oauth2.Token{}
  err = yaml.Unmarshal(authContent, &token)
  if err != nil {
    return token, err
  }

  return token, nil
}
