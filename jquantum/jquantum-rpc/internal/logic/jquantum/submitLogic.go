package jquantumlogic

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/errorx"
	"io"
	"jian-unified-system/jus-core/time"
	jq "jian-unified-system/jus-core/types/mq/jquantum"
	"os"
	"path/filepath"
	"strconv"

	"jian-unified-system/jquantum/jquantum-rpc/internal/svc"
	"jian-unified-system/jquantum/jquantum-rpc/jquantum"
	jquantumModel "jian-unified-system/jus-core/data/mysql/jquantum"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubmitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitLogic {
	return &SubmitLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Submit 提交任务
func (l *SubmitLogic) Submit(in *jquantum.SubmitReq) (*jquantum.SubmitResp, error) {
	// todo: add your logic here and delete this line

	data := in.Thread
	readerAt := bytes.NewReader(data)

	jobID := uuid.NewString()

	err := l.saveFiles(readerAt, int64(len(data)), in.UserId, jobID)
	if err != nil {
		return nil, err
	}

	_, err = l.svcCtx.JobModel.Insert(l.ctx, &jquantumModel.Job{
		Id:     jobID,
		UserId: in.UserId,
		Status: 0,
	})
	if err != nil {
		return nil, err
	}

	err = l.writeMsg(in.UserId, jobID)
	if err != nil {
		return nil, err
	}

	return &jquantum.SubmitResp{
		JobId: jobID,
	}, nil
}

func (l *SubmitLogic) saveFiles(data io.ReaderAt, size int64, userID int64, jobID string) error {
	dir := filepath.Join(l.svcCtx.Config.JQuantum.BaseDir, strconv.FormatInt(userID, 10), jobID)

	// 创建目标目录
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return errorx.Wrap(err, "Failed to create target dir")
	}

	// 读取 zip 文件
	zipReader, err := zip.NewReader(data, size)
	if err != nil {
		return errorx.Wrap(err, "Invalid zip file")
	}

	// 解压每个文件到指定路径
	for _, zipFile := range zipReader.File {
		zipFilePath := filepath.Join(dir, zipFile.Name)

		if zipFile.FileInfo().IsDir() {
			_ = os.MkdirAll(zipFilePath, os.ModePerm)
			continue
		}

		// 确保目录存在
		if err := os.MkdirAll(filepath.Dir(zipFilePath), os.ModePerm); err != nil {
			return errorx.Wrap(err, "Failed to create dir for file")
		}

		// 创建目标文件
		dstFile, err := os.OpenFile(zipFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zipFile.Mode())
		if err != nil {
			return errorx.Wrap(err, "Failed to create file (\" + zipFile.Name + \")")
		}

		// 解压内容
		srcFile, err := zipFile.Open()
		if err != nil {
			_ = dstFile.Close()
			return errorx.Wrap(err, "Failed to open zip entry: (\" + zipFile.Name + \")")
		}

		_, err = io.Copy(dstFile, srcFile)
		_ = srcFile.Close()
		_ = dstFile.Close()

		if err != nil {
			return errorx.Wrap(err, "Failed to write file (\" + zipFile.Name + \")")
		}
	}
	return nil
}

func (l *SubmitLogic) writeMsg(userID int64, jobID string) error {
	// RabbitMQ
	jobMsg := &jq.JobStructureMsg{
		UserID: userID,
		JobID:  jobID,
		Time:   time.Now().ToString(),
	}
	msgData, err := json.Marshal(jobMsg)
	if err != nil {
		return errorx.Wrap(err, "Job Message err")
	}

	err = l.svcCtx.Producer.Publish(msgData)
	if err != nil {
		fmt.Println(err.Error())
	}
	return nil
}
