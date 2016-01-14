package config

import(
  "os"
  yaml "github.com/go-yaml/yaml"
  log "github.com/Sirupsen/logrus"
  "io/ioutil"
)

type Auth struct {
  ClientId        string  `yaml:"client_id"`
  ClientSecret    string  `yaml:"client_secret"`
}

type Config struct {
  Auth      Auth
  Graphite  string
}

func ParseConfig(configFile string) Config{
  configContent, err := ioutil.ReadFile(configFile)
  if err != nil {
    log.Fatal(err)
    os.Exit(1)
  }

  parsedConfig := Config{}
  err = yaml.Unmarshal(configContent, &parsedConfig)
  if err != nil {
    log.Fatalf("error: %v", err)
  }

  return parsedConfig
}
