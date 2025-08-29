package code

import (
	"fmt"
	"path/filepath"
)

const (
	// CodeTemplateBeginning QuEST C++量子电路代码模版 - 开头
	CodeTemplateBeginning string = `/** @file
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
// jobDir: 参数文件路径 - params.json
// return: C++ 代码字符串
func MainCodeTemplateBeginning(numQubits int, jobDir string) string {
	return fmt.Sprintf(`// 重复模式-结束==========

int main() {
    // 加载参数
    ifstream f("%s");
    json compressed;
    f >> compressed;
    Params params;

    // 初始化QuEST环境
    initQuESTEnv();
    QuESTEnv env = getQuESTEnv();  // 获取环境信息

    // 初始化经典寄存器数组
    int creg[%d] = {0};

    // 创建 %d 量子比特系统
    Qureg qureg = createQureg(%d);
    initZeroState(qureg);

    // 应用量子门-开始==========
`, filepath.Join(jobDir, "params.json"), numQubits, numQubits, numQubits)
}

// MainCodeTemplateEnding QuEST C++量子电路代码模版 - main - 结尾
// numQubits: 量子比特数
// return: C++ 代码字符串
func MainCodeTemplateEnding(numQubits int) string {
	return fmt.Sprintf(`    // 应用量子门-结束==========

    // 报告状态
    reportStr("Final state:");
    // reportQureg(qureg);

	//cout << "Hello from rank " << env.rank << endl;
    // 计算并报告概率分布
    if (env.rank == 0) {
        cout << "\n量子态概率分布:\n";
		// cout << "[rank " << env.rank << "] 量子态概率分布:\n";
        // 遍历所有可能的状态
        for (long long int i = 0; i < %d; i++) {
            // 获取状态i的振幅
            qcomp amp = getQuregAmp(qureg, i);

            // 使用real()和imag()函数获取实部和虚部
            qreal realPart = real(amp);
            qreal imagPart = imag(amp);

            // 计算概率
            double prob = realPart * realPart + imagPart * imagPart;

			if(prob < 1e-3) {
                continue;
            }

            // 打印状态及其概率
            cout << "|";
            for (int q = %d; q >= 0; q--) {
                cout << ((i >> q) & 1);
            }
            cout << ">: " << prob << "\n";
        }
    }

    // 清理资源
    destroyQureg(qureg);
    finalizeQuESTEnv();
    return 0;
}`, 1<<numQubits, numQubits-1)
}
