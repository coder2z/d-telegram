package download

import (
	"context"
	"fmt"
	"github.com/coder2z/d-telegram/config"
	"github.com/coder2z/d-telegram/xlog"
	"github.com/google/uuid"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
	"github.com/panjf2000/ants/v2"
	"github.com/spf13/cast"
	"go.uber.org/zap"
	"os"
	"path"
	"strings"
	"unicode/utf8"
)

var (
	ch chan *tg.Message
)

func Push(item *tg.Message) {
	xlog.GetLogger().Info("download.Push", zap.Any("item", item))
	ch <- item
}

func Init(client *tg.Client, cfg *config.Config) {
	ch = make(chan *tg.Message, cfg.DownloadPool/10)
	p, _ := ants.NewPoolWithFunc(cfg.DownloadPool, h{
		cfg:    cfg,
		client: client,
	}.handle())
	go func() {
		defer p.Release()
		for {
			select {
			case item := <-ch:
				_ = p.Invoke(item)
			}
		}
	}()
}

type h struct {
	cfg    *config.Config
	client *tg.Client
}

func (_self h) handle() func(item interface{}) {
	return func(item interface{}) {
		ctx := context.Background()
		message, ok := item.(*tg.Message)
		if !ok {
			return
		}
		mediaDocument, ok := message.Media.(*tg.MessageMediaDocument)
		if !ok {
			return
		}
		document, ok := mediaDocument.Document.(*tg.Document)
		if !ok {
			return
		}

		// 取message的前几位作为文件名
		var (
			fileName, fileNamePrefix string
		)
		if len(message.Message) > 8 {
			fileNamePrefix = SubStrDecodeRuneInString(getStrCn(message.Message), 8)
		} else {
			fileNamePrefix = getStrCn(message.Message)
		}

		var id int64

		switch v := message.PeerID.(type) {
		case *tg.PeerChannel:
			id = v.ChannelID
		case *tg.PeerChat:
			id = v.ChatID
		case *tg.PeerUser:
			id = v.UserID
		}

		filePath := path.Join(_self.cfg.DownloadDir, cast.ToString(id))

		if len(document.Attributes) >= 2 {
			filename, _ := document.Attributes[1].(*tg.DocumentAttributeFilename)
			fileName = path.Join(
				filePath,
				fmt.Sprintf("%s_%s", fileNamePrefix, filename.FileName))
		}
		if len(fileName) <= 0 {
			mTypeL := strings.Split(document.MimeType, "/")
			fileName = path.Join(filePath,
				fmt.Sprintf("%s_%s", fileNamePrefix, uuid.New().String()+"."+mTypeL[len(mTypeL)-1]))
		}
		// 判断文件夹是否存在 不存在就创建
		if !isExist(filePath + "/") {
			if err := os.MkdirAll(filePath+"/", os.ModePerm); err != nil {
				xlog.GetLogger().Error("mkdir err", zap.Error(err))
				return
			}
		}

		// 文件下载用携程去干
		newDownloader := downloader.NewDownloader()
		_, err := newDownloader.Download(_self.client, &tg.InputDocumentFileLocation{
			ID:            document.ID,
			AccessHash:    document.AccessHash,
			FileReference: document.FileReference,
		}).ToPath(ctx, fileName)
		if err != nil {
			xlog.GetLogger().Info("download.Download ERROR", zap.Error(err))
			return
		}
		xlog.GetLogger().Info("download.Download success", zap.Any("fileName", fileName))
	}
}

// GetStrCn 提取中文
func getStrCn(str string) (cnStr string) {
	r := []rune(str)
	var strSlice []string
	for i := 0; i < len(r); i++ {
		if r[i] <= 40869 && r[i] >= 19968 {
			cnStr = cnStr + string(r[i])
			strSlice = append(strSlice, cnStr)
		}
	}
	return
}

func SubStrDecodeRuneInString(s string, length int) string {
	var size, n int
	for i := 0; i < length && n < len(s); i++ {
		_, size = utf8.DecodeRuneInString(s[n:])
		n += size
	}
	return s[:n]
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return true
}
