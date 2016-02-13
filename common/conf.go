package common
import (
	"io/ioutil"
	"encoding/json"
)

func ReadConfig(filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		logger.Fatal(err.Error())
	}
	err = json.Unmarshal(data, &C)
	if err != nil {
		logger.Fatal(err.Error())
	}
}
