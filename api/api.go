package api

import(
  "os"
  "fmt"
  "github.com/yosmudge/graphatmo/config"
  log "github.com/Sirupsen/logrus"
  "golang.org/x/oauth2"
  "github.com/segmentio/go-prompt"
  "strings"
  "encoding/json"
  "os/user"
  yaml "github.com/go-yaml/yaml"
  "io/ioutil"
  "net/http"
  "net/url"
)

type Api struct {
  Auth            config.Auth
  AuthFile        string
  BaseHost        string
  Endpoint        oauth2.Endpoint
  Config          oauth2.Config
  Client          *http.Client
}

type Request struct {
  Method          string
  Path            string
  Params          url.Values
  Response        http.Response
  Data            map[string]interface{}
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

func (self *Api) DoLogin() {
  // Do first time login for user/pass auth

  fmt.Printf(`Enter your email address and password to authenticate with Netatmo
These will not get stored, they will be exchanges via xauth for Oauth tokens which will be stored in %s
`, self.AuthFile)

  username := prompt.String("Netatmo Email")
  password := prompt.Password("Netatmo Password")

  fmt.Println("Attempting to authenticate with Netatmo...")
  token, err := self.Config.PasswordCredentialsToken(oauth2.NoContext, username, password)

  // Handling errors as strings, good job I like bash scripts
  if err != nil {
    errorStr := err.Error()
    if strings.Contains(errorStr, "Response") {
      responseLines := strings.Split(errorStr, "\n")[1]
      responseJSON := []byte(strings.TrimLeft(responseLines, "Response: "))
      var response map[string]interface{}

      err := json.Unmarshal(responseJSON, &response)
      if err != nil {
        log.Error("Could not parse JSON response from Netatmo during authentication")
      } else if v, ok := response["error"]; ok {
        if v.(string) == "invalid_grant" {
          log.Error("Invalid Grant: Maybe your email or password is incorrect. Try logging in to https://my.netatmo.com")
        } else {
          log.Error(responseLines)
        }
      } else {
        log.Error(responseLines)
      }
    } else if strings.HasPrefix(errorStr, "oauth2: cannot fetch token") {
      log.Error("Internal error from Netatmo API. Netatmo may be down or experiancing issues")
    } else {
      log.Error(err)
    }
    os.Exit(1)
  }

  fmt.Println("Authenticated with Netatmo successfully!")
  fmt.Printf("Saving auth details to %s\n", self.AuthFile)
  err = self.UpdateAuth(token)
  if err != nil {
    log.WithFields(log.Fields{"Error":err}).Error("Error saving auth data")
  }
}

func (self *Api) UpdateAuth(token *oauth2.Token) error {
  // Save auth data

  log.WithFields(log.Fields{"Target":self.AuthFile}).Debug("Saving authentication data")

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

func (self *Api) PrepareClient() error{
  // Ensure client is loaded and token is refreshed
  token, err:= self.LoadToken()
  if err != nil {
    return err
  }

  if self.Client == nil {
    log.Debug("Creating client")
    self.Client = self.Config.Client(oauth2.NoContext, &token)
  }

  return nil
}

func (self *Api) DoCall(request *Request) error {
  // Call netatmo API
  // Only implements "GET" at the moment because that's all we need
  if request.Method == "" {
    request.Method = "GET"
  }

  if request.Params == nil {
    request.Params = url.Values{}
  }

  err := self.PrepareClient()

  baseUrl, _ := url.Parse(self.BaseHost)
  baseUrl.Path = fmt.Sprintf("/api/%s", request.Path)
  baseUrl.RawQuery = request.Params.Encode()

  log.WithFields(log.Fields{
    "Method": request.Method,
    "URL": baseUrl.String(),
  }).Debug("Making call")

  if err != nil {
    return fmt.Errorf("Error with call %s %s: %s", request.Method, request.Path, err)
  }

  switch{
  case request.Method == "GET":
    resp, err := self.Client.Get(baseUrl.String())
    if err != nil {
      return err
    }
    log.WithFields(log.Fields{
      "StatusCode":resp.StatusCode,
      "ContentLength":resp.ContentLength,
    }).Debug("Netatmo Response")

    defer resp.Body.Close()
    responseRaw, _ := ioutil.ReadAll(resp.Body)

    if resp.StatusCode == 200 {
      if err := json.Unmarshal(responseRaw, &request.Data); err != nil {
        panic(err)
      }
    } else {
      return fmt.Errorf("Error with %s call: %s %s", request.Path, resp.StatusCode, responseRaw)
    }
  default:
    panic(fmt.Sprintf("Called DoCall with unknown method %s", request.Method))
  }

  return nil
}
