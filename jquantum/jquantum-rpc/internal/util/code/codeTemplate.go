package code

import (
	"fmt"
)

const (
	// TemplateBeginning QuEST C++量子电路代码模版 - 开头
	TemplateBeginning string = `/** @file
* 由 Qiskit 生成的 QuEST 模拟代码
* author: JQuantum
* version: v0.1
*/

#include "quest/include/quest.h"
#include <cmath>  // 包含数学函数
#include "quest/include/jquantum.h"
#include <fstream>

// 重复模式-开始==========`
)

// MainCodeTemplateBeginning QuEST C++量子电路代码模版 - main - 开头
// numQubits: 量子比特数
// shots: 采样数
// jobDir: 参数文件路径 - params.json
// return: C++ 代码字符串
func MainCodeTemplateBeginning(numQubits, shots int, jobDir string) string {
	return fmt.Sprintf(`// 重复模式-结束==========

int main() {
    // 加载参数
	string job_dir = "%s";
	int num_qubits = %d;
	int shots = %d;
    ifstream f(job_dir + "/params.json");
    json compressed;
    f >> compressed;
    Params params;

    // 初始化QuEST环境
    initQuESTEnv();
    QuESTEnv env = getQuESTEnv();  // 获取环境信息

    // 初始化经典寄存器数组
    // int creg[num_qubits] = {0};

    // 创建 %d 量子比特系统
    Qureg qureg = createQureg(num_qubits);
    initZeroState(qureg);

    // 应用量子门-开始==========
`, jobDir, numQubits, shots, numQubits)
}

// MainCodeTemplateEnding QuEST C++量子电路代码模版 - main - 结尾
// return: C++ 代码字符串
func MainCodeTemplateEnding() string {
	return fmt.Sprintf(`    // 应用量子门-结束==========

    // 报告状态
    // reportStr("Final state:");
    // reportQureg(qureg);

    // 计算并报告概率分布
    if (env.rank == 0) {
		string jsonResult = exportToJson(qureg, num_qubits, shots);
		
		// 将结果保存到文件
        ofstream outFile(job_dir + "/result.json");
        outFile << jsonResult;
        outFile.close();
        
        cout << "success" << endl;
    }

    // 清理资源
    destroyQureg(qureg);
    finalizeQuESTEnv();
    return 0;
}`)
}
