//
// Created by 祁剑 on 13/07/2025.
// Update: 祁剑 31/08/2025
//

//#ifndef TEST_JQUANTUM_H
//#define TEST_JQUANTUM_H
//
//#endif //TEST_JQUANTUM_H

#include <iostream>
#include "quest/include/quest.h"
#include "quest.h"
#include <vector>
#include "json.hpp"
#include <fstream>
#include <iomanip>
#include <random>
#include <cmath>
#include <sstream>

using namespace std;
using json = nlohmann::json;
using Params = vector<json>;

Params get_gate_params(const json& compressed, int start, int count = 1);

vector<int> get_int_array(const json& j);

Params get_repeats_params(const Params& params, int repeats = 1);


// 正确的状态向量获取函数
vector<complex<double>> getStatevector(Qureg qureg, int numQubits);

// 正确的计数计算函数
map<string, int> computeCounts(Qureg qureg, int numQubits, int shots = 1024);

// 导出为 JSON 的函数
string exportToJson(Qureg qureg, int numQubits, string jobId, int shots = 1024);