package download

import (
	"testing"
)

func TestName(t *testing.T) {
	// 分割字符串拿到最后一位
	message := "#福利姬 网红美少女【茶杯恶犬】退圈未流出典"
	var fileNamePrefix string
	if len(message) > 8 {
		fileNamePrefix = SubStrDecodeRuneInString(getStrCn(message), 8)
	} else {
		fileNamePrefix = getStrCn(message)
	}
	t.Log(fileNamePrefix)
}
