package models

import(
  "time"
)

type Module struct {
  Name          string
  Id            string
  RfStrength    float64
  Battery       float64
  Measures      []string
  LastData      int
  DataInterval  int
}

func ModuleFromJson(moduleObj map[string]interface{}) Module{
  // Create a module from a JSON blob
  mod := Module{}

  mod.Name = moduleObj["module_name"].(string)
  mod.Id = moduleObj["_id"].(string)
  if val, ok := moduleObj["rf_status"]; ok {
    mod.RfStrength = val.(float64)
  }
  if val, ok := moduleObj["battery_vp"]; ok {
    mod.Battery = val.(float64)
  }

  mod.DataInterval = 300

  for _, measure := range moduleObj["data_type"].([]interface{}){
    mod.Measures = append(mod.Measures, measure.(string))
  }

  return mod
}

func (mod *Module) NextData() time.Time{
  // Return expected next data for module
  // Netatmo API has 10 minute cache time, so update every 11 minutes
  return time.Unix(int64(mod.LastData), 0).Add(time.Duration(11)*time.Minute)
}
