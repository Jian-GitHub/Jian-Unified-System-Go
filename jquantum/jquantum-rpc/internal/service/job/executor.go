package jobService

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
	"jian-unified-system/jquantum/jquantum-rpc/internal/util/code"
	"jian-unified-system/jus-core/types/mq/jquantum"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type Executor struct {
	//UserID int64
	//JobID  string
	baseDir string
	dir     string
}

func NewExecutor(baseDir string) *Executor {
	return &Executor{
		//UserID: userID,
		//JobID:  jobID,
		//dir:    filepath.Join(baseDir, strconv.FormatInt(userID, 10), jobID),
		baseDir: baseDir,
	}
}

func (e *Executor) Process(body []byte) {
	var msg jquantum.JobStructureMsg
	err := json.Unmarshal(body, &msg)
	if err != nil {
		return
	}
	e.dir = filepath.Join(e.baseDir, strconv.FormatInt(msg.UserID, 10), msg.JobID)

	//executor := joblogic.NewExecutor(msg.UserID, msg.JobID, c.config.JQuantum.BaseDir)
	e.GenerateCode()
	e.Compile()
	e.Run()
}

// readJSONFile 读取JSON文件内容
func (e *Executor) readJSONFile() ([]byte, error) {
	fileName := "structure.json"
	filePath := filepath.Join(e.dir, fileName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法读取JSON文件 %s: %v", fileName, err)
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

	//jsonData := "{\"num_qubits\": 18, \"patterns\": {\"pattern_1\": {\"content\": [{\"count\": 18, \"ref\": \"h\"},\n                                        {\"count\": 13, \"ref\": \"x\"},\n                                        \"h\",\n                                        \"mcx\",\n                                        \"h\",\n                                        {\"count\": 26, \"ref\": \"x\"},\n                                        \"h\",\n                                        \"mcx\",\n                                        \"h\",\n                                        {\"count\": 13, \"ref\": \"x\"},\n                                        {\"count\": 18, \"ref\": \"h\"},\n                                        {\"count\": 18, \"ref\": \"x\"},\n                                        \"h\",\n                                        \"mcx\",\n                                        \"h\",\n                                        {\"count\": 18, \"ref\": \"x\"}],\n                            \"count\": 16,\n                            \"total\": 133}},\n \"sequence\": [{\"count\": 284, \"ref\": \"pattern_1\"}, {\"count\": 18, \"ref\": \"h\"}]}"

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

	// 输出解析的信息
	fmt.Printf("解析成功: %d 量子比特, %d 模式, %d 序列项\n",
		result.NumQubits, len(result.Patterns), len(result.Sequence))

	// 创建转换器并生成代码
	converter := code.NewConverter(e.dir)
	questCode, err := converter.CircuitToQuestJSON(result)
	if err != nil {
		//logx.Errorf("生成代码失败: %v", err)
		return errorx.Wrap(err, "生成代码失败")
	}

	// 输出生成的代码到文件
	outputFile := filepath.Join(e.dir, "code.cpp")
	err = os.WriteFile(outputFile, []byte(questCode), 0644)
	if err != nil {
		//logx.Errorf("写入输出文件失败: %v", err)
		return errorx.Wrap(err, "写入输出文件失败")
	}

	fmt.Printf("代码已生成到: %s\n", outputFile)
	fmt.Printf("量子比特数: %d\n", result.NumQubits)
	if len(result.Patterns) == 0 {
		fmt.Println("提示: 电路较小，未检测到重复模式")
	}
	return nil
}

// Compile MPICXX 编译可执行文件
func (e *Executor) Compile() error {
	cmd := exec.Command("mpicxx",
		filepath.Join(e.dir, "code.cpp"),
		"-o", filepath.Join(e.dir, "program"),
		"-I/harmoniacore/jquantum",
		"-L/harmoniacore/jquantum/lib",
		"-lQuEST-fp2+mt+mpi",
		"-lm",
		"-lstdc++",
		"-ljquantum",
	)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	if err := cmd.Run(); err != nil {
		stdoutStr := stdoutBuf.String()
		stderrStr := stderrBuf.String()
		fmt.Println("MPICXX File Compilation exited with error:", err)
		fmt.Println(stdoutStr)
		fmt.Println(stderrStr)
		fmt.Println("===================")
		return errorx.Wrap(err, "MPICXX File Compilation exited with error")
	} else {
		stdoutStr := stdoutBuf.String()
		stderrStr := stderrBuf.String()
		fmt.Println("MPICXX File Compilation completed.")
		fmt.Println(stdoutStr)
		fmt.Println(stderrStr)
		fmt.Println("===================")
	}
	return nil
}

func (e *Executor) Run() error {
	cmd := exec.Command("mpirun", "--allow-run-as-root",
		"--hostfile", "/harmoniacore/jquantum/hosts.txt",
		"-np", "4",
		filepath.Join(e.dir, "program"),
	)
	// 继承当前环境变量
	cmd.Env = os.Environ()
	// 添加 LD_LIBRARY_PATH
	cmd.Env = append(cmd.Env, "LD_LIBRARY_PATH=/harmoniacore/jquantum/lib")

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
		return errorx.Wrap(err, "MPIRUN job exited with error")
	} else {
		stdoutStr := stdoutBuf.String()
		//stderrStr := stderrBuf.String()
		fmt.Println("MPIRUN job completed.")
		fmt.Println(strings.TrimSpace(stdoutStr))
		//fmt.Println(strings.TrimSpace(stderrStr))
		fmt.Println("===================")
	}
	return nil
}
