package structure

type VideoOwner struct {
	Mid  int64  `json:"mid"`
	Name string `json:"name"`
	Face string `json:"face"`
}

type VideoPage struct {
	Page     int    `json:"page"`
	Part     string `json:"part"`
	Duration int    `json:"duration"`
}

type Video struct {
	Bvid    string      `json:"bvid"`
	Aid     int64       `json:"aid"`
	Title   string      `json:"title"`
	Pic     string      `json:"pic"`
	Owner   VideoOwner  `json:"owner"`
	Desc    string      `json:"desc"`
	PubDate int64       `json:"pubDate"`
	Pages   []VideoPage `json:"pages"`
}
