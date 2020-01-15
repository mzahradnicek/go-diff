package godiff

import (
	"encoding/json"
)

func (d arlist) MarshalJSON() ([]byte, error) {
	res := map[string]interface{}{}

	for k, v := range d {
		res[k.String()] = v
	}

	return json.Marshal(res)
}

func (d modlist) MarshalJSON() ([]byte, error) {
	res := map[string]update{}

	for k, v := range d {
		res[k.String()] = v
	}

	return json.Marshal(res)
}
