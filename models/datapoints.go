package models


type DataPoint struct {
  Time          Timestamp
  Data          map[string]float64
  Complete      bool
}

type DataPointList struct {
  DataPoints    []DataPoint
}

func (self *DataPointList) LatestTimestamp() Timestamp{
  return self.DataPoints[len(self.DataPoints)-1].Time
}

func (self *DataPointList) DataInterval() int{
  return self.LatestTimestamp().Unix-self.DataPoints[len(self.DataPoints)-2].Time.Unix
}

type ByTimestamp []DataPoint

func(self ByTimestamp) Len() int { return len(self) }
func(self ByTimestamp) Swap(i, j int) { self[i], self[j] = self[j], self[i] }
func(self ByTimestamp) Less(i, j int) bool { return self[i].Time.Unix < self[j].Time.Unix }
