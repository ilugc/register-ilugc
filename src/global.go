package register

import (
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"slices"
)

type Global struct {
	logger *log.Logger
}
var G *Global

type Participant struct {
	RegisteredTimeMicro int64 `json:"timemicro"`
	RegisteredTime string `json:"time"`
	Name string `json:"name"`
	Email string `json:"email"`
	Mobile string `json:"mobile"`
	Org string `json:"org"`
	Place string `json:"place"`
	QrCode []byte `json:"qrcode"`
}

type ConfigDetails struct {
	Domain string `json:"domain"`
	Hostport string `json:"hostport"`
	Static string `json:"static"`
	OpenedTime string `json:"openedtime"`
	OpenedTimeMicro int64 `json:"openedtimemicro"`
	DefaultMax int64 `json:"defaultmax"`
	StopRegistration bool `json:"stopregistration"`
	AdminUsername string `json:"adminusername"`
	AdminPasswordHash []byte `json:"adminpasswordhash"`
}

func StructToMap(v any, ignorelist []string) map[string]string {
	vmap := make(map[string]string)
	valueof := reflect.Indirect(reflect.ValueOf(v))
	typeof := valueof.Type()
	for _, visiblefield := range reflect.VisibleFields(typeof) {
		if slices.Contains(ignorelist, visiblefield.Name) == true {
			continue
		}

		valuef := valueof.FieldByIndex(visiblefield.Index)
		switch valuef.Kind() {
		case reflect.Int64: vmap[visiblefield.Name] = strconv.FormatInt(valuef.Int(), 10)
		case reflect.Uint64: vmap[visiblefield.Name] = strconv.FormatUint(valuef.Uint(), 10)
		case reflect.Bool: vmap[visiblefield.Name] = strconv.FormatBool(valuef.Bool())
		case reflect.String: vmap[visiblefield.Name] = valuef.String()
		}
	}
	return vmap
}

func StructSetFromMap(v any, m map[string]any, ignorelist []string) {
	valueof := reflect.Indirect(reflect.ValueOf(v))
	for k0, v0 := range m {
		if slices.Contains(ignorelist, k0) == true {
			continue
		}

		valuef := valueof.FieldByName(k0)
		if valuef.IsValid() == false {
			G.logger.Println("Invalid value for key ", k0)
			continue
		}
		switch valuef.Kind() {
		case reflect.Int64: valuef.SetInt(int64(v0.(float64)))
		case reflect.Uint64: valuef.SetUint(uint64(v0.(float64)))
		case reflect.Bool: valuef.SetBool(v0.(bool))
		case reflect.String: valuef.SetString(v0.(string))
		}
	}
}

func GetStaticPath() (string, error) {
	realpath, err := filepath.EvalSymlinks("/proc/self/exe")
	if err != nil {
		G.logger.Println(err)
		return "", err
	}
	staticpath := filepath.Clean(filepath.Dir(realpath) + "/../../static")
	_, err = os.Stat(staticpath)
	if err != nil {
		G.logger.Println(err)
		return "", err
	}
	return staticpath, nil
}

func init() {
	if G != nil {
		return
	}
	G = &Global{}
	G.logger = log.Default()
	G.logger.SetFlags(log.Ldate |log.Ltime | log.Lmicroseconds | log.Lshortfile)
}
