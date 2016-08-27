package netatmo

import(
  "os"
  "fmt"
  "strings"
  "encoding/json"
  "golang.org/x/oauth2"
  log "github.com/Sirupsen/logrus"
  "github.com/segmentio/go-prompt"
)

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
      log.Error("Internal error from Netatmo API. Netatmo may be down or experiencing issues")
    } else {
      log.Error(err)
    }
    os.Exit(1)
  }

  fmt.Println("Authenticated with Netatmo successfully!")
  fmt.Printf("Saving auth details to %s\n", self.AuthFile)
  err = self.SaveToken(token)
  if err != nil {
    log.WithFields(log.Fields{"Error":err}).Error("Error saving auth data")
  }
}
