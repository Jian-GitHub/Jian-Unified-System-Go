package jobService

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
	"jian-unified-system/apollo/apollo-rpc/apollo"
	"jian-unified-system/jquantum/jquantum-rpc/internal/svc"
	"jian-unified-system/jquantum/jquantum-rpc/internal/util/code"
	ap "jian-unified-system/jus-core/data/mysql/apollo"
	"jian-unified-system/jus-core/types/mq/jquantum"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type Executor struct {
	UserID            int64
	JobID             string
	svc               *svc.ServiceContext
	exeDir            string
	np                int64
	codeFileName      string
	structureFileName string
	programFileName   string
	hostsFileName     string
}

func NewExecutor(svc *svc.ServiceContext) *Executor {
	return &Executor{
		svc:               svc,
		codeFileName:      "code.cpp",
		structureFileName: "structure.json",
		programFileName:   "program",
		hostsFileName:     "hosts",
	}
}

func (e *Executor) Process(body []byte) {
	var msg jquantum.JobStructureMsg
	err := json.Unmarshal(body, &msg)
	if err != nil {
		return
	}
	e.UserID = msg.UserID
	e.JobID = msg.JobID
	e.exeDir = filepath.Join(e.svc.Config.JQuantum.BaseUserDir, strconv.FormatInt(msg.UserID, 10), msg.JobID)

	// Send Email - Notify User Job is processing
	e.sendEmail("JQuantum Job is Processing", "Your Quantum Computing Job (Job ID: "+e.JobID+") is processing now.")
	e.updateJobState(1)
	err = e.GenerateCode()
	if err != nil {
		e.sendEmail("JQuantum Job is Finished (Failed)", "Your Quantum Computing Job (Job ID: "+e.JobID+") is finished now.\n Error info: \n"+err.Error())
		e.updateJobState(-1)
		return
	}
	err = e.Compile()
	if err != nil {
		e.sendEmail("JQuantum Job is Finished (Failed)", "Your Quantum Computing Job (Job ID: "+e.JobID+") is finished now.\n Error info: \n"+err.Error())
		return
	}
	_ = e.Run()
}

// readJSONFile 读取JSON文件内容
func (e *Executor) readJSONFile() ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(e.exeDir, e.structureFileName))
	if err != nil {
		return nil, fmt.Errorf("无法读取JSON文件 %s: %v", e.structureFileName, err)
	}
	return data, nil
}

// GenerateCode 根据 structure.json 生成 QuEST C++ 代码
func (e *Executor) GenerateCode() error {
	// 读取JSON文件
	jsonData, err := e.readJSONFile()
	if err != nil {
		logx.Error("错误: %v", err)
		return errorx.Wrap(err, "GenerateCode Error")
	}

	// 解析JSON数据
	var result jquantum.ResultJSON
	//err = json.Unmarshal([]byte(jsonData), &result)
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		//logx.Errorf("解析JSON失败: %v", err)
		return errorx.Wrap(err, "解析JSON失败")
	}

	// 设置默认值（如果某些字段缺失）
	if result.NumQubits <= 0 {
		//logx.Errorf("参数错误 量子比特数: %v", err)
		return errorx.Wrap(err, "参数错误 量子比特数")
	}
	if result.Sequence == nil {
		//logx.Errorf("结构错误 无门操作: %v", err)
		return errorx.Wrap(err, "结构错误 无门操作")
	}
	if result.Patterns == nil {
		result.Patterns = map[string]jquantum.PatternContent{}
	}

	// 判断集群可用状态, 内存是否足够
	resource, err := e.svc.KubernetesDeployService.CollectClusterResource()
	if err != nil {
		return errorx.Wrap(err, "集群错误 无法采集集群资源数据")
	}
	if resource.MaxQubits < result.NumQubits {
		return errorx.Wrap(err, fmt.Sprintf("资源不足错误 集群可用内存最大计算 %d 位, 当前任务计算 %d 位", resource.MaxQubits, result.NumQubits))
	}

	err = e.svc.KubernetesDeployService.GenerateHostsFile(resource, filepath.Join(e.svc.Config.JQuantum.BaseDir, e.hostsFileName))
	if err != nil {
		return errorx.Wrap(err, "MPI前置错误 无法生成 hosts 文件")
	}

	e.np = resource.TotalSlotsPow2

	// 创建转换器并生成代码
	converter := code.NewConverter(e.exeDir, e.JobID)
	questCode, err := converter.CircuitToQuestJSON(result)
	if err != nil {
		//logx.Errorf("生成代码失败: %v", err)
		return errorx.Wrap(err, "生成代码失败")
	}

	// 输出生成的代码到文件
	outputFile := filepath.Join(e.exeDir, e.codeFileName)
	err = os.WriteFile(outputFile, []byte(questCode), 0644)
	if err != nil {
		//logx.Errorf("写入输出文件失败: %v", err)
		return errorx.Wrap(err, "写入输出文件失败")
	}

	return nil
}

// Compile MPICXX 编译可执行文件
func (e *Executor) Compile() error {
	cmd := exec.Command("mpicxx",
		filepath.Join(e.exeDir, e.codeFileName),
		"-o", filepath.Join(e.exeDir, e.programFileName),
		"-I"+e.svc.Config.JQuantum.BaseDir,
		"-L"+e.svc.Config.JQuantum.BaseLibDir,
		"-lQuEST-fp2+mt+mpi",
		"-lm",
		"-lstdc++",
		"-ljquantum",
	)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	if err := cmd.Run(); err != nil {
		//stdoutStr := stdoutBuf.String()
		stderrStr := stderrBuf.String()
		fmt.Println("MPICXX File Compilation exited with error:", err)
		//fmt.Println(stdoutStr)
		fmt.Println(stderrStr)
		//fmt.Println("===================")
		e.updateJobState(-1)
		return errorx.Wrap(err, "MPICXX File Compilation exited with error")
	} else {
		//stdoutStr := stdoutBuf.String()
		//stderrStr := stderrBuf.String()
		fmt.Println("MPICXX File Compilation completed.")
		//fmt.Println(stdoutStr)
		//fmt.Println(stderrStr)
		//fmt.Println("===================")
	}
	return nil
}

func (e *Executor) Run() error {
	cmd := exec.Command("mpirun", "--allow-run-as-root",
		"-x", "LD_LIBRARY_PATH=/harmoniacore/jquantum/lib",
		"--hostfile", filepath.Join(e.svc.Config.JQuantum.BaseDir, e.hostsFileName),
		"-np", strconv.FormatInt(e.np, 10),
		filepath.Join(e.exeDir, e.programFileName),
	)
	// 继承当前环境变量
	cmd.Env = os.Environ()
	// 添加 LD_LIBRARY_PATH
	cmd.Env = append(cmd.Env, "LD_LIBRARY_PATH="+e.svc.Config.JQuantum.BaseLibDir)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	if err := cmd.Run(); err != nil {
		stdoutStr := stdoutBuf.String()
		stderrStr := stderrBuf.String()
		fmt.Println("MPIRUN job exited with error:", err)
		fmt.Println(strings.TrimSpace(stdoutStr))
		fmt.Println(strings.TrimSpace(stderrStr))
		fmt.Println("===================")
		e.sendEmail("JQuantum Job is Finished (Failed)", "Your Quantum Computing Job (Job ID: "+e.JobID+") is finished now.\n Error info: \n"+strings.TrimSpace(stderrStr))
		e.updateJobState(-2)
		return errorx.Wrap(err, "MPIRUN job exited with error")
	} else {
		stdoutStr := stdoutBuf.String()
		stderrStr := stderrBuf.String()
		fmt.Println("MPIRUN job completed.")
		fmt.Println(strings.TrimSpace(stdoutStr))
		fmt.Println(strings.TrimSpace(stderrStr))
		//fmt.Println("===================")
		e.sendEmail("JQuantum Job is Finished (Success)", "Your Quantum Computing Job (Job ID: "+e.JobID+") is finished now.")
		e.updateJobState(2)
	}
	return nil
}

func (e *Executor) updateJobState(state int) {
	go func() {
		err := e.svc.JobModel.UpdateState(context.Background(), e.JobID, state)
		if err != nil {
			return
		}
	}()
}

func (e *Executor) sendEmail(subject, body string) {
	// Send Email
	go func() {
		fmt.Println("Email Starting")
		resp, err := e.svc.ApolloAccount.UserInfo(context.Background(), &apollo.UserInfoReq{
			UserId: e.UserID,
		})
		if err != nil {
			fmt.Println("Apollo User Account Error: ", err.Error())
			return
		}
		if resp == nil {
			fmt.Println("No notification info email")
			return
		}

		var user ap.UserInfo
		err = json.Unmarshal(resp.UserBytes, &user)

		name := user.FamilyName + user.MiddleName + user.GivenName
		if name == "" {
			name = "User"
		}

		err = e.svc.EmailService.Send(
			user.NotificationEmail.String,
			subject,
			"<div>Hello "+name+"!</div><div>"+body+"</div>",
		)
		if err != nil {
			fmt.Println(err.Error())
			return
		} else {
			fmt.Println("Email: ok")
		}
	}()
}
