package api

type File struct {
	Sha1       string
	ModifyTime uint
	Size       int64
	CreateTime uint
	Name       string
	Id         string
	Type       string
	Revision   string
	Parent     *File
	Child      []*File
	IsActive   bool
}

type Msg struct {
	Result    string
	Retryable bool
	Code      int
	Data      struct {
		HasMore bool
		List    []*File
	}
}

type MsgV2 struct {
	Result    string
	Retryable bool
	Code      int
	Data      File
}

type UploadJson struct {
	Content UploadContent `json:"content"`
}

type UploadContent struct {
	Name     string      `json:"name"`
	ParentId string      `json:"parentId"`
	Storage  interface{} `json:"storage"`
}
type UploadStorage struct {
	Size     int64       `json:"size"`
	Sha1     string      `json:"sha1"`
	Kss      interface{} `json:"kss"`
	UploadId string      `json:"uploadId"`
	Exists   bool        `json:"exists"`
}
type UploadExistedStorage struct {
	UploadId string `json:"uploadId"`
	Exists   bool   `json:"exists"`
}

type UploadKss struct {
	BlockInfos []BlockInfo `json:"block_infos"`
}

type Kss struct {
	Stat            string              `json:"stat"`
	NodeUrls        interface{}         `json:"node_urls"`
	SecureKey       string              `json:"secure_key"`
	ContentCacheKey string              `json:"contentCacheKey"`
	FileMeta        string              `json:"file_meta"`
	CommitMetas     []map[string]string `json:"commit_metas"`
}

type BlockInfo struct {
	Blob struct{} `json:"blob"`
	Sha1 string   `json:"sha1"`
	Md5  string   `json:"md5"`
	Size int64    `json:"size"`
}

type FileDb struct {
	Sha1       string
	Md5        string
	ModifyTime int64
}
