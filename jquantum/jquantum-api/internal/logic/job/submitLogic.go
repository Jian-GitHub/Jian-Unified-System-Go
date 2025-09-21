package job

import (
	"context"
	"encoding/json"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
	"jian-unified-system/jquantum/jquantum-api/internal/svc"
	"jian-unified-system/jquantum/jquantum-api/internal/types"
	"jian-unified-system/jquantum/jquantum-rpc/jquantum"
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

func (l *SubmitLogic) Submit(data []byte) (types.SubmitResp, error) {
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

	//// RabbitMQ
	//userID, err := l.ctx.Value("id").(json.Number).Int64()
	//if err != nil {
	//	return types.BaseResponse{
	//		Code:    -1,
	//		Message: "No user id",
	//	}, err
	//}
	//jobMsg := &jquantum.JobStructureMsg{
	//	UserID: userID,
	//	JobID:  jobID,
	//	CreateTime:   time.Now().ToString(),
	//}
	//msgData, err := json.Marshal(jobMsg)
	//if err != nil {
	//	return types.BaseResponse{
	//		Code:    -2,
	//		Message: "Job Message err",
	//	}, err
	//}
	//
	//err = l.svcCtx.Producer.Publish(msgData)
	//if err != nil {
	//	fmt.Println(err.Error())
	//}

	id, err := l.ctx.Value("id").(json.Number).Int64()
	if err != nil {
		return types.SubmitResp{
			BaseResponse: types.BaseResponse{
				Code:    -1,
				Message: "No user id",
			},
		}, errorx.Wrap(err, "No user id")
	}

	resp, err := l.svcCtx.JQuantumClient.Submit(l.ctx, &jquantum.SubmitReq{
		UserId: id,
		Thread: data,
	})
	if err != nil {
		return types.SubmitResp{
			BaseResponse: types.BaseResponse{
				Code:    -2,
				Message: "JQuantum Rpc err",
			},
		}, errorx.Wrap(err, "JQuantum Rpc err")
	}

	return types.SubmitResp{
		BaseResponse: types.BaseResponse{
			Code:    200,
			Message: "success",
		},
		JobID: resp.JobId,
	}, nil
}
