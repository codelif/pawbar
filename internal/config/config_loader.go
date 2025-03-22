package config

import (
	"os"

	"gopkg.in/yaml.v3"
)


func LoadBarConfig(configPath string) (BarConfig, error){
  var bc BarConfig
  data, err := os.ReadFile(configPath)
  if err != nil {
    return bc, err
  }

  err = yaml.Unmarshal(data, &bc)
  if err != nil {
    return bc, err
  }

  return bc, nil
}
