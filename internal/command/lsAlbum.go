package command

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"go-micloud/internal/api"
	"go-micloud/pkg/utils"
	"time"
)

func (r *Command) ListAlbum() *cli.Command {
	return &cli.Command{
		Name:  "lsAlbum",
		Usage: "列出所有相册",
		Action: func(ctx *cli.Context) error {
			ablums, err := r.Request.GetAblums()
			if err != nil {
				return err
			}
			format(ablums)
			return nil
		},
	}
}

func format(ablums []*api.Album) {
	fmt.Printf("total %d\n", len(ablums))
	fmt.Printf("文件数 |    最后更新时间     | 相册名\n")
	fmt.Printf("---------------------------\n")
	for _, v := range ablums {
		t := time.Unix(v.LastUpdateTime/1000, 0)
		fmt.Printf("%-6d | %-6s | %s\n", v.MediaCount, t.Format(utils.YmdHis), v.Name)
	}
}
