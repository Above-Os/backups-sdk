package restic

import (
	"encoding/json"
	"reflect"
)

type ResticEnv struct {
	AWS_ACCESS_KEY_ID     string `env:"AWS_ACCESS_KEY_ID" json:"aws_access_key_id,omitempty"`
	AWS_SECRET_ACCESS_KEY string `env:"AWS_SECRET_ACCESS_KEY" json:"aws_secret_access_key,omitempty"`
	AWS_SESSION_TOKEN     string `env:"AWS_SESSION_TOKEN" json:"aws_session_token,omitempty"`
	RESTIC_REPOSITORY     string `env:"RESTIC_REPOSITORY" json:"restic_repository,omitempty"`
	RESTIC_PASSWORD       string `env:"RESTIC_PASSWORD" json:"-"`
}

func (r *ResticEnv) ToMap() map[string]string {
	m := make(map[string]string)
	v := reflect.ValueOf(*r)
	t := reflect.TypeOf(*r)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		tag := t.Field(i).Tag.Get("env")
		if tag != "" && field.String() != "" {
			m[tag] = field.String()
		}
	}
	return m
}

func (r *ResticEnv) ToString() string {
	res, _ := json.Marshal(r)
	return string(res)
}
