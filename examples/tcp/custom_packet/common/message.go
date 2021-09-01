package common

type Json01Req struct {
	Key1 string `json:"key_1"`
	Key2 int    `json:"key_2"`
	Key3 bool   `json:"key_3"`
}

type Json01Resp struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}
