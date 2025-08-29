package job

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"jian-unified-system/jquantum/jquantum-api/internal/svc"
	"jian-unified-system/jquantum/jquantum-api/internal/types"
	"jian-unified-system/jus-core/time"
	"jian-unified-system/jus-core/types/mq/jquantum"
)

type SubmitLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubmitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitLogic {
	return &SubmitLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitLogic) Submit(jobID string) (types.BaseResponse, error) {
	// todo: add your logic here and delete this line
	// Kafka
	//data := "JQ"
	//err := l.svcCtx.KafkaWriter.WriteMessages(l.ctx, kafka.Message{
	//	Key:   []byte("亓"),
	//	Value: []byte(data),
	//	Headers: []kafka.Header{
	//		{
	//			Key:   "祁",
	//			Value: []byte("剑"),
	//		},
	//	},
	//})
	//if err != nil {
	//	fmt.Println(err.Error())
	//	return err
	//}

	// RabbitMQ
	userID, err := l.ctx.Value("id").(json.Number).Int64()
	if err != nil {
		return types.BaseResponse{
			Code:    -1,
			Message: "No user id",
		}, err
	}
	jobMsg := &jquantum.JobStructureMsg{
		UserID: userID,
		JobID:  jobID,
		Time:   time.Now().ToString(),
	}
	msgData, err := json.Marshal(jobMsg)
	if err != nil {
		return types.BaseResponse{
			Code:    -2,
			Message: "Job Message err",
		}, err
	}

	err = l.svcCtx.Producer.Publish(msgData)
	if err != nil {
		fmt.Println(err.Error())
	}

	return types.BaseResponse{
		Code:    200,
		Message: "success",
	}, nil
}
