package godiff

import (
	"encoding/json"
)

func (d arlist) JSONize() map[string]interface{} {
	res := map[string]interface{}{}

	for k, v := range d {
		res[k.String()] = v
	}

	return res
}

func (d arlist) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.JSONize())
}

func (d modlist) JSONize() map[string]interface{} {
	res := map[string]interface{}{}

	for k, v := range d {
		res[k.String()] = v
	}

	return res
}

func (d modlist) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.JSONize())
}
