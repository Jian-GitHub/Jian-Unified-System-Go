from typing import List, Dict, Union, Tuple, Any
from qiskit import QuantumCircuit


# 增强版门归一化（包含量子电路特有操作）
def normalize_gate(gate: str) -> str:
    # 等价门分组
    equivalence_groups = {
        # ('mcx', 'cx', 'ccx'): 'mcx',
        # ('crx', 'cry', 'crz'): 'cr',
        # ('rx', 'ry', 'rz'): 'r',
        # ('mcx', 'mcy', 'mcz'): 'mc',
        # ('x', 'y', 'z'): 'single_axis',  # 单量子比特门
        # ('h', 's', 'sdg', 't', 'tdg', 'id'): 'single_gate',  # 其他单量子比特门
        # ('cx', 'cy', 'cz', 'ch'): 'two_qubit',  # 两量子比特门
        # ('swap', 'iswap'): 'swap_type',
        # ('ccx', 'cswap'): 'toffoli_type',
        # ('reset', 'measure'): 'meta_operation'  # 元操作
    }

    # 特殊处理多控制门模式
    # if gate.startswith('mc'):
    #     return 'mc'

    # 检查等价组
    for group, normalized in equivalence_groups.items():
        if gate in group:
            return normalized

    return gate


class QuantumPatternEncoder:
    def __init__(self):
        from . import params_generator
        self.pattern_defs = {}
        self.params = []
        self.params_generator = params_generator.ParamsGenerator()
        self.pattern_counter = 1
        self.pattern_cache = {}
        self.value_threshold = 4  # 模式最小价值阈值
        self.min_pattern_length = 3  # 最小模式长度
        self.support_gates = ['h', 'x', 'y', 'z', 's', 'sdg', 't', 'tdg', 'sx', 'sxdg', 'rx', 'ry', 'rz', 'p', 'u', 'cx', 'cz', 'cy', 'ch', 'swap', 'iswap', 'cu3', 'crx', 'cry', 'crz', 'cu1', 'rxx', 'ryy',
                         'rzz', 'rzx', 'ccx', 'cswap', 'cswap', 'mcx', 'mcy', 'mcz', 'mcp']

    def _pattern_value(self, length: int, count: int) -> int:
        """计算模式价值（考虑长度和重复次数）"""
        return length * count

    def _create_pattern_key(self, segment: List[Union[str, Dict]]) -> Tuple:
        """创建模式唯一标识键（考虑量子电路特性）"""
        return tuple(
            item if isinstance(item, str) else (item['ref'], item['count'])
            for item in segment
        )

    def _run_length_encode(self, gates: List[str]) -> List[Dict]:
        """量子电路专用游程编码"""
        if not gates:
            return []

        normalized = [normalize_gate(g) for g in gates]
        # normalized = gates
        result = []
        current = gates[0]
        current_norm = normalized[0]
        count = 1

        for i in range(1, len(gates)):
            # 特殊处理测量和重置操作
            # if gates[i] in ['measure', 'reset']:
            #     if count > 0:
            #         if count > 1:
            #             result.append({'ref': current, 'count': count})
            #         else:
            #             result.append(current)
            #         count = 0
            #     result.append(gates[i])
            #     continue

            if normalized[i] == current_norm:
                count += 1
            else:
                if count > 1:
                    result.append({'ref': current, 'count': count})
                else:
                    result.append(current)
                current = gates[i]
                current_norm = normalized[i]
                count = 1

        if count > 0:
            if count > 1:
                result.append({'ref': current, 'count': count})
            else:
                result.append(current)

        return result

    def _detect_subpatterns(
            self,
            sequence: List[Union[str, Dict]],
            depth: int = 0
    ) -> List[Union[str, Dict]]:
        """量子电路专用模式检测"""
        if len(sequence) < self.min_pattern_length:
            return sequence

        n = len(sequence)
        i = 0
        result = []

        while i < n:
            # 跳过测量和重置操作
            # if isinstance(sequence[i], str) and sequence[i] == 'measure':
            #     i += 1
            #     continue

            best_value = self.value_threshold
            best_pattern = None
            best_length = 0
            best_repeats = 0

            # 限制搜索范围以提高性能
            max_length = min(50, (n - i) // 2)  # 最大模式长度限制

            # 寻找以i开始的最佳模式
            for L in range(self.min_pattern_length, max_length + 1):
                j = i + L
                repeats = 1

                # 计算连续重复次数
                while j <= n - L:
                    segment1 = sequence[i:i + L]
                    segment2 = sequence[j:j + L]

                    if self._pattern_equal(segment1, segment2):
                        repeats += 1
                        j += L
                    else:
                        break

                # 计算模式价值
                pattern_value = self._pattern_value(L, repeats)

                # 更新最佳模式
                if repeats > 1 and pattern_value > best_value:
                    best_value = pattern_value
                    best_pattern = sequence[i:i + L]
                    best_length = L
                    best_repeats = repeats

            # 如果找到高价值模式
            if best_pattern:
                # 递归处理子模式
                compressed_pattern = self._detect_subpatterns(best_pattern, depth + 1)

                # 创建模式键并检查缓存
                pattern_key = self._create_pattern_key(compressed_pattern)

                if pattern_key in self.pattern_cache:
                    pattern_ref = self.pattern_cache[pattern_key]
                else:
                    pattern_ref = f"pattern_{self.pattern_counter}"
                    self.pattern_counter += 1
                    # self.pattern_defs[pattern_ref] = compressed_pattern
                    self.pattern_defs[pattern_ref] = {
                        "content": compressed_pattern,
                        "count": len(compressed_pattern),
                        "total": self._count_total_gates(compressed_pattern)
                    }

                    self.pattern_cache[pattern_key] = pattern_ref

                # 添加模式引用
                result.append({
                    'ref': pattern_ref,
                    'count': best_repeats
                })
                i += best_length * best_repeats
            else:
                # 没有模式，添加当前元素
                result.append(sequence[i])
                i += 1

        return result

    def _count_total_gates(self, pattern: List[Union[str, Dict]]) -> int:
        """递归统计模式中实际包含的门操作总数"""
        total = 0
        for item in pattern:
            if isinstance(item, str):
                total += 1
            elif isinstance(item, dict):
                ref = item['ref']
                count = item['count']
                if ref in self.pattern_defs:
                    # 如果是引用已有 pattern
                    ref_content = self.pattern_defs[ref]["content"]
                    sub_total = self._count_total_gates(ref_content)
                    total += sub_total * count
                else:
                    # 普通压缩结构，如 {"ref": "x", "count": 4}
                    total += count
        return total

    def _pattern_equal(
            self,
            seq1: List[Union[str, Dict]],
            seq2: List[Union[str, Dict]]
    ) -> bool:
        """量子电路专用模式比较"""
        if len(seq1) != len(seq2):
            return False

        for item1, item2 in zip(seq1, seq2):
            if type(item1) != type(item2):
                return False

            if isinstance(item1, str):
                # 特殊处理测量和重置操作
                if item1 in ['measure', 'reset'] or item2 in ['measure', 'reset']:
                    return item1 == item2

                if normalize_gate(item1) != normalize_gate(item2):
                    return False
            else:  # dict (模式引用)
                if item1['ref'] != item2['ref'] or item1['count'] != item2['count']:
                    return False

        return True

    def encode_circuit(self, qc: QuantumCircuit, shots: int = 1024) -> Dict[str, Any]:
        """编码量子电路"""
        # 重置状态
        self.pattern_defs = {}
        self.pattern_cache = {}
        self.pattern_counter = 1

        # 映射量子比特
        qubit_map = {qubit: idx for idx, qubit in enumerate(qc.qubits)}

        # 提取门序列
        gate_sequence = []
        num_gates = 0
        for instruction in qc:
            gate_name = instruction.operation.name.lower()
            if gate_name not in self.support_gates:
                continue
            gate_sequence.append(gate_name)
            num_gates += 1

            # 获取映射后的量子位索引
            qubits = [qubit_map[q] for q in instruction.qubits]
            # 根据量子门类型分发到对应的处理函数
            num_qubits = len(qubits)
            if num_qubits == 1:
                self.params_generator.handle_single_qubit_gate(instruction, qubits)
            elif num_qubits == 2:
                self.params_generator.handle_two_qubit_gate(instruction, qubits)
            elif num_qubits >= 3:
                self.params_generator.handle_multi_qubit_gate(instruction, qubits)

        self._set_params(self.params_generator.get_params())

        # 处理连续重复
        rle_sequence = self._run_length_encode(gate_sequence)

        # 检测子序列模式
        final_sequence = self._detect_subpatterns(rle_sequence)

        return {
            "shots": shots,
            "num_gates": num_gates,
            "num_qubits": qc.num_qubits,
            "patterns": self.pattern_defs,
            "sequence": final_sequence
        }

    def _set_params(self, params):
        self.params = params

    def get_params(self):
        return self.params
