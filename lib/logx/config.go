package logx

type Config struct {
	ServiceName         string `json:",optional"`
	Mode                string `json:",options=regular|console|volume,default=regular"`
	Path                string `json:",default=logs"`
	Compress            bool   `json:",optional"`
	KeepDays            int    `json:",optional"`
	StackCooldownMillis int    `json:",default=100"`
}
