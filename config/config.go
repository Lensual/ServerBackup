package config

type Config struct {
	Global TaskTemplate            `yaml:"global"`
	Task   map[string]TaskTemplate `yaml:"task"`
}

type TaskTemplate struct {
	Backup   Backup   `yaml:"backup"`
	Compress Compress `yaml:"compress"`
	Storage  Storage  `yaml:"storage"`
	Cron     string   `yaml:"cron"`
}

// func (t *TaskTemplate) Merge(from *TaskTemplate) {
// 	typeOfT := reflect.TypeOf(*t)
// 	typeOfF := reflect.TypeOf(*from)
// 	for i := 0; i < typeOfF.NumField(); i++ {
// 		fField := typeOfF.Field(i)
// 		tField, ok := typeOfT.FieldByName(fField.Name)
// 		if ok {
// 			fValue := reflect.ValueOf(fField)
// 			tValue := reflect.ValueOf(tField)
// 			tValue.Set(fValue)
// 		}
// 	}
// }

type Backup struct {
	Driver string                 `yaml:"driver"`
	Option map[string]interface{} `yaml:"option"`
}

type Compress struct {
	Driver string                 `yaml:"driver"`
	Option map[string]interface{} `yaml:"option"`
}

type Storage struct {
	Driver string                 `yaml:"driver"`
	Option map[string]interface{} `yaml:"option"`
}
