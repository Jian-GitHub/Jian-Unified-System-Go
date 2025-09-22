package util

import (
	"github.com/zeromicro/go-zero/core/logx"
	"time"
)

func RetryWithBackoff(name string, action func() error) error {
	backoff := time.Second        // 初始 1s
	maxBackoff := time.Minute * 5 // 上限 5m

	for {
		if err := action(); err != nil {
			logx.Errorf("%s 失败: %v", name, err)
			logx.Infof("等待 %v 后重试", backoff)
			time.Sleep(backoff)

			if backoff < maxBackoff {
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			}
			continue
		}
		return nil
	}
}
