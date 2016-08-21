package netatmo

import(
  "fmt"
  "net/url"
  "net/http"
  "io/ioutil"
  "encoding/json"
  log "github.com/Sirupsen/logrus"
)

type Request struct {
  Method          string
  Path            string
  Params          url.Values
  Response        http.Response
  Data            map[string]interface{}
}

func (self *Api) DoCall(request *Request, response interface{}) error {
  <-self.Limiter
  // Call netatmo API
  // Only implements "GET" at the moment because that's all we need
  if request.Method == "" {
    request.Method = "GET"
  }

  if request.Params == nil {
    request.Params = url.Values{}
  }

  baseUrl, _ := url.Parse(self.BaseHost)
  baseUrl.Path = fmt.Sprintf("/api/%s", request.Path)
  baseUrl.RawQuery = request.Params.Encode()
  baseLogFields := log.Fields{
    "method": request.Method,
    "path": baseUrl.String(),
  }

  log.WithFields(baseLogFields).Debug("Making call")

  err := self.PrepareClient()
  if err != nil {
    return fmt.Errorf("Error with call %s %s: %s", request.Method, request.Path, err)
  }

  switch{
  case request.Method == "GET":
    resp, err := self.Client.Get(baseUrl.String())
    if err != nil {
      baseLogFields["err"] = err
      log.WithFields(baseLogFields).Error("Error making call")
      return err
    }

    defer resp.Body.Close()
    responseRaw, _ := ioutil.ReadAll(resp.Body)

    if resp.StatusCode == 200 {
      if err := json.Unmarshal(responseRaw, &request.Data); err != nil {
        baseLogFields["err"] = err
        log.WithFields(baseLogFields).Error("Error making call")
        return fmt.Errorf("Could not unmarshal JSON: %s", err)
      }
      if err := json.Unmarshal(responseRaw, response); err != nil {
        baseLogFields["err"] = err
        log.WithFields(baseLogFields).Error("Error making call")
        return fmt.Errorf("Could not unmarshal JSON: %s", err)
      }
    } else {
      baseLogFields["err"] = err
      log.WithFields(baseLogFields).Error("Error making call")
      return fmt.Errorf("Error with %s call: %d %s", request.Path, resp.StatusCode, responseRaw)
    }
  default:
    panic(fmt.Sprintf("Called DoCall with unknown method %s", request.Method))
  }

  return nil
}
