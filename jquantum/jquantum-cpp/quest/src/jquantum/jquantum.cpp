//
// Created by 祁剑 on 13/07/2025.
// Update: 祁剑 31/08/2025
//

#include "jquantum.h"

Params get_gate_params(const json& compressed, int start, int count) {
    const auto& patterns = compressed["patterns"];
    const auto& sequence = compressed["sequence"];
    const auto& flat = compressed["flat_params"];

    Params result;
    int idx = 0;        // 当前展开位置
    int flat_idx = 0;   // flat_params 下标

    for (const auto& block : sequence) {
        string ref = block["ref"].is_null() ? "" : block["ref"].get<string>();
        int times = block["count"];

        if (ref.empty()) {
            for (int t = 0; t < times; ++t) {
                if (idx >= start && idx < start + count) {
                    result.push_back({ flat[flat_idx] });  // 包一层 vector 表示参数组
                }
                ++flat_idx;
                ++idx;
                if (idx >= start + count)
                    return result;
            }
        } else {
            const auto& pattern = patterns[ref];
            int plen = (int) pattern.size();

            for (int t = 0; t < times; ++t) {
                for (int k = 0; k < plen; ++k) {
                    if (idx >= start && idx < start + count) {
                        result.push_back({ pattern[k] });  // 包一层 vector 表示参数组
                    }
                    ++idx;
                    if (idx >= start + count)
                        return result;
                }
            }
        }
    }

    return result;
}

Params get_repeats_params(const Params& params, const int repeats) {
    const int group_size = (int) params.size() / repeats;
    Params repeats_params = {};
    for (int r = 0; r < repeats; ++r) {
        for (int k = 0; k < group_size; ++k) {
            repeats_params.push_back(params[r * group_size + k]);
        }
    }
    return repeats_params;
}

vector<int> get_int_array(const json& j) {
    return j.get<vector<int>>();
}

// 正确的状态向量获取函数
vector<complex<double>> getStatevector(Qureg qureg, int numQubits) {
    vector<complex<double>> statevector;
    int numStates = 1 << numQubits;
    statevector.reserve(numStates);
    
    for (int i = 0; i < numStates; i++) {
        // 使用正确的 QuEST API 获取振幅
        qcomp amp = getQuregAmp(qureg, i);
        qreal realPart = real(amp);
        qreal imagPart = imag(amp);
        statevector.emplace_back(realPart, imagPart);
    }
    
    return statevector;
}

// 正确的计数计算函数
map<string, int> computeCounts(Qureg qureg, int numQubits, int shots) {
    map<string, int> counts;
    int numStates = 1 << numQubits;
    
    // 计算概率分布
    vector<double> probabilities(numStates);
    for (int i = 0; i < numStates; i++) {
        qcomp amp = getQuregAmp(qureg, i);
        qreal realPart = real(amp);
        qreal imagPart = imag(amp);
        probabilities[i] = realPart * realPart + imagPart * imagPart;
    }
    
    // 从概率分布中采样
    random_device rd;
    mt19937 gen(rd());
    discrete_distribution<> dist(probabilities.begin(), probabilities.end());
    
    for (int i = 0; i < shots; i++) {
        int outcome = dist(gen);
        string bitstring;
        for (int j = numQubits - 1; j >= 0; j--) {
            bitstring += ((outcome >> j) & 1) ? '1' : '0';
        }
        counts[bitstring]++;
    }
    
    return counts;
}

// 导出为 JSON 的函数
string exportToJson(Qureg qureg, int numQubits, string jobId, int shots) {
    stringstream ss;
    ss << setprecision(15);
    
    auto counts = computeCounts(qureg, numQubits, shots);
    auto statevec = getStatevector(qureg, numQubits);
    
    ss << "{";
    ss << "\"backend_name\": \"JQuantum\",";
    ss << "\"backend_version\": \"0.1.0\",";
    ss << "\"job_id\": \"" << jobId << "\",";
    ss << "\"success\": true,";
    ss << "\"results\": [{";
    ss << "\"shots\": " << shots << ",";
    ss << "\"success\": true,";
    ss << "\"meas_level\": 2,";
    ss << "\"data\": {";
    
    // 添加计数数据
    ss << "\"counts\": {";
    bool first = true;
    for (const auto& pair : counts) {
    if (!first) ss << ",";
        // 直接使用二进制字符串
        ss << "\"" << pair.first << "\": " << pair.second;
        first = false;
    }
    ss << "},";
    
    // 添加状态向量数据
    ss << "\"statevector\": [";
    first = true;
    for (const auto& amp : statevec) {
        if (!first) ss << ",";
        ss << "[" << amp.real() << "," << amp.imag() << "]";
        first = false;
    }
    ss << "]";
    
    ss << "},";
    ss << "\"header\": {\"num_qubits\": " << numQubits << "}";
    ss << "}]";
    ss << "}";
    
    return ss.str();
}
