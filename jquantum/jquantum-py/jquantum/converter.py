import json
from . import compress
from qiskit import QuantumCircuit
from . import pattern_code

class Converter:
    def __init__(self):
        self.qbit_index = 0
        self.params = []
        self.encoder = compress.QuantumPatternEncoder()

    # ========================
    # 辅助函数
    # ========================
    def get_targets_array(self, targets):
        """生成目标量子位数组的C++代码"""
        if len(targets) == 0:
            return ""
        target_str = ', '.join(map(str, targets))
        # return f"    int targets_{targets[0]}[] = {{{target_str}}};\n"
        return target_str
    from typing import Any, Dict, List, Union

    Element = Union[str, Dict[str, Any]]
    PatternDict = Dict[str, Dict[str, Any]]
    support_gates = ['h', 'x', 'y', 'z', 's', 'sdg', 't', 'tdg', 'sx', 'sxdg', 'rx', 'ry', 'rz', 'p', 'u', 'cx',
                          'cz', 'cy', 'ch', 'swap', 'iswap', 'cu3', 'crx', 'cry', 'crz', 'cu1', 'rxx', 'ryy',
                          'rzz', 'rzx', 'ccx', 'cswap', 'cswap', 'mcp']


    def resolve_element_structured(self, element: Element, patterns: PatternDict) -> Any:
        """结构化解析一个元素（保留嵌套结构）"""
        if isinstance(element, str):
            return element
        elif isinstance(element, dict):
            ref = element["ref"]
            count = element.get("count", 1)

            if ref in patterns:
                content = patterns[ref]["content"]
                return [
                    [self.resolve_element_structured(item, patterns) for item in content]
                    for _ in range(count)
                ]
            else:
                # 普通引用（不是 pattern），直接展开为列表
                return [ref for _ in range(count)]
        else:
            raise ValueError(f"未知元素类型: {element}")

    def generate_patterns_code(self, patterns: Dict[str, Dict[str, Any]]) -> str | None:
        patterns_head = ""
        patterns_content = "\n"
        patterns_code = ""
        param_index = 0

        if patterns is None or len(patterns) == 0:
            return patterns_code

        for name, pattern in patterns.items():
            # print(f"📦 Pattern: {name}")
            # print("共:", pattern.get("count", []))
            patterns_head += f"void {name}(Qureg& qureg, const Params& pattern_params);\n"
            patterns_content += f"void {name}(Qureg& qureg, const Params& pattern_params) {{\n"
            patterns_content += f"\tParams params;\n"
            content = pattern.get("content", [])

            for idx, item in enumerate(content):
                if isinstance(item, str):
                    # print(f"  {idx}. Literal: {item}")
                    patterns_content += f"\n\tparams = {{}};\n"
                    patterns_content += f"\n\tparams.push_back(pattern_params[{param_index}]);\n"
                    patterns_content += pattern_code.name_to_code(item, 1)
                    param_index += 1
                elif isinstance(item, dict):
                    ref = item.get("ref", "<missing>")
                    count = item.get("count", 1)
                    # print(f"  {idx}. Ref: {ref} (count={count})")
                    patterns_content += "\n\tparams = {};\n"
                    patterns_content += f"\tfor (int i = {param_index}; i < {param_index + count}; i++) {{\n"
                    patterns_content += f"\t\tparams.push_back(pattern_params[i]);\n"
                    patterns_content += f"\t}};\n"
                    patterns_content += f"    for(int i = 0; i < {count}; i++) {{\n"
                    patterns_content += pattern_code.name_to_code(ref, 2, 'i')
                    patterns_content += f"    }}\n"
                    param_index += count
                else:
                    print(f"  {idx}. ! Unexpected item: {item}")

            patterns_content += "}\n"
            print()
            patterns_code += patterns_head + patterns_content
            # print(patterns_code)
            return patterns_code

    def expand_all_patterns_structured(self, patterns: PatternDict) -> Dict[str, List[Any]]:
        """批量结构化展开所有 pattern，保持 key 映射"""
        result = {}
        for name, pattern_data in patterns.items():
            result.update({name: self.expand_pattern_structured(pattern_data, patterns)})
        return result

    def expand_pattern_structured(self, pattern_data: Dict[str, Any], patterns: PatternDict) -> List[Any]:
        """展开一个 pattern 的内容（保留嵌套结构）"""
        patterns_expanded = []
        content = pattern_data["content"]
        patterns_expanded.append([self.resolve_element_structured(item, patterns) for item in content])
        return patterns_expanded

    def generate_pattern_content_params(self, element: Element, patterns: PatternDict) -> Any:
        """结构化解析一个元素（保留嵌套结构）"""
        if isinstance(element, str):
            return element
        elif isinstance(element, dict):
            ref = element["ref"]
            count = element.get("count", 1)

            if ref in patterns:
                content = patterns[ref]["content"]
                return [
                    [self.generate_pattern_content_params(item, patterns) for item in content]
                    for _ in range(count)
                ]
            else:
                # 普通引用（不是 pattern），直接展开为列表
                return [ref for _ in range(count)]
        else:
            raise ValueError(f"未知元素类型: {element}")

    def generate_pattern_params(self, pattern_data: Dict[str, Any], patterns: PatternDict) -> List[Any]:
        """展开一个 pattern 的内容（保留嵌套结构）"""
        patterns_expanded = []
        content = pattern_data["content"]
        patterns_expanded.append([self.generate_pattern_content_params(item, patterns) for item in content])
        return patterns_expanded

    def generate_sequence_code(self, sequence_data: List[Dict[str, Any]], patterns: PatternDict) -> str:
        """展开 sequence 列表（保留嵌套结构）"""
        sequence_code = ""
        instruction_index = 0
        for item in sequence_data:
            if isinstance(item, str):
                # print(f"跳过字符串项: {item}")
                # if item in patterns:
                #     sequence_code += f"\tparams = get_gate_params(compressed, {instruction_index});\n"

                sequence_code += f"\tparams = get_gate_params(compressed, {instruction_index});\n"
                sequence_code += pattern_code.name_to_code(item, 1) + "\n"
                instruction_index += 1
                # sequence_code += "    " + item + "();\n"
                continue

            ref = item["ref"]
            count = item.get("count", 1)

            # print(ref)
            # print(count)

            if ref in patterns:
                qbit_num = patterns[ref]["total"]
                patterns_total_gates_num = qbit_num * count
                # sequence_code += f"\tparams = get_gate_params(compressed, {instruction_index}, {qbit_num});\n"
                # sequence_code += f"\tparams = get_repeats_params(get_gate_params(compressed, {instruction_index}, {patterns_total_gates_num}), {patterns_total_gates_num});\n"
                sequence_code += f"    for(int i = {instruction_index}; i < {instruction_index + patterns_total_gates_num}; i+={qbit_num}) {{\n"
                sequence_code += f"        params = get_gate_params(compressed, i, {qbit_num});\n"
                sequence_code += "        " + ref + "(qureg, params);\n"
                sequence_code += "    }\n\n"
                instruction_index += patterns_total_gates_num

                # print(f"for(int i = 0; i < {count}; i++) {{")
                # print("    apply_" + ref + "();")
                # print("}")

                # pattern = patterns[ref]
                # expanded = expand_pattern_structured(pattern, patterns)
                # for _ in range(count):
                #     result.extend(expanded)
            else:
                # 原子引用
                # result.append([ref] * count)

                if count > 1:
                    # sequence_code += f"\tparams = get_gate_params(compressed, {instruction_index}, {count});\n"
                    # get_repeats_params(get_gate_params(compressed, 0, 10), 10);
                    sequence_code += f"\tparams = get_repeats_params(get_gate_params(compressed, {instruction_index}, {count}), {count});\n"
                    sequence_code += f"    for(int i = 0; i < {count}; i++) {{\n"
                    sequence_code += pattern_code.name_to_code(ref, 2, 'i')
                    # sequence_code += "        " + ref + "();\n"
                    sequence_code += "    }\n\n"
                    instruction_index += count
                else:
                    sequence_code += f"\tparams = get_gate_params(compressed, {instruction_index});\n"
                    sequence_code += pattern_code.name_to_code(ref, 1)
                    # sequence_code += "        " + ref + "();\n"
                    instruction_index += 1

                # print(f"for(int i = 0; i < {count}; i++) {{")
                # print("    " + ref + "();")
                # print("}")

        return sequence_code

    def expand_sequence(self, sequence_data: List[Dict[str, Any]], patterns: PatternDict) -> List:
        """展开 sequence 列表（保留嵌套结构）"""
        result = []
        # print(sequence_data)
        for item in sequence_data:
            if isinstance(item, str):
                result.append(item)
                continue

            ref = item["ref"]
            count = item.get("count", 1)

            # print(ref)
            # print(count)

            if ref in patterns:
                pattern = patterns[ref]
                expanded = self.expand_pattern_structured(pattern, patterns)
                # pprint.pprint("===")
                # pprint.pprint(expanded)
                # pprint.pprint("===")
                # for _ in range(count):
                #     result.extend(expanded)
                result.extend([[expanded] * count])
            else:
                # 原子引用
                print('item: ', item)
                if count > 1:
                    result.append([ref] * count)
                else:
                    result.append([ref])

                # result.append([ref] * count)

        return result

    def prepare_params_sequence(self, sequence_data: List[Dict[str, Any]], patterns: PatternDict, qc: QuantumCircuit) -> List:
        """展开 sequence 列表（保留嵌套结构）"""
        params = []
        result = []
        instruction_index = 0
        sequence_code = ""
        # 映射量子比特
        qubit_map = {qubit: idx for idx, qubit in enumerate(qc.qubits)}

        # print(sequence_data)
        for index, item in enumerate(sequence_data):
            # print(index)
            # print(qc[index])
            # print(type(qc[index]))

            # 如果是原子引用
            if isinstance(item, str):
                if item == "measure":
                    # print("跳过measure")
                    instruction_index += 1
                    continue

                result.append(item)
                # sequence_code += "    " + item + "();\n"
                # print('instruction_index', instruction_index)
                # print('qc ', qc[instruction_index])
                sequence_code = self.convert_instruction(qc[instruction_index], qubit_map, sequence_code)
                instruction_index += 1
                continue

            # 重复模式
            ref = item["ref"]
            count = item.get("count", 1)

            # print(ref)
            # print(count)

            if ref in patterns:
                pattern = patterns[ref]
                expanded = self.generate_pattern_params(pattern, patterns)
                # pprint.pprint("===")
                # pprint.pprint(expanded)
                # pprint.pprint("===")
                # for _ in range(count):
                #     result.extend(expanded)
                result.extend([[expanded] * count])
                instruction_index += (pattern['total']) * count
            else:
                # 简单重复 - {'ref': x, 'count': x}
                if ref == "measure":
                    # print("跳过measure")
                    instruction_index += count
                    continue
                # 原子引用
                # print('item: ', item)
                if count > 1:
                    result.append([ref] * count)
                else:
                    result.append([ref])
                instruction_index += count
                # result.append([ref] * count)

        # print("aaaaaaaaaa")
        # print(sequence_code)
        # print("sequence_code")
        return result

    # def ensure_list(self, x):
    #     """返回list"""
    #     if isinstance(x, (list, tuple)):
    #         return list(x)
    #     else:
    #         return [x]

    # ========================
    # 单量子位门操作
    # ========================
    def handle_single_qubit_gate(self, instruction, qubits, code):
        """处理单量子位门操作"""
        op_name = instruction.operation.name.lower()
        target = qubits[0]

        if op_name == 'h':
            code += f"    // H门: 量子位 {target}\n"
            code += f"    applyHadamard(qureg, {target});\n"

        elif op_name == 'x':
            code += f"    // X门: 量子位 {target}\n"
            code += f"    applyPauliX(qureg, {target});\n"

        elif op_name == 'y':
            code += f"    // Y门: 量子位 {target}\n"
            code += f"    applyPauliY(qureg, {target});\n"

        elif op_name == 'z':
            code += f"    // Z门: 量子位 {target}\n"
            code += f"    applyPauliZ(qureg, {target});\n"

        elif op_name == 's':
            code += f"    // S门: 量子位 {target}\n"
            code += f"    applyS(qureg, {target});\n"

        elif op_name == 'sdg':
            code += f"    // S†门: 量子位 {target}\n"
            code += f"    DiagMatr1 s_dagger = getDiagMatr1({{1, -1_i}});\n"
            code += f"    applyDiagMatr1(qureg, {target}, s_dagger);\n"

        elif op_name == 't':
            code += f"    // T门: 量子位 {target}\n"
            code += f"    applyT(qureg, {target});\n"

        elif op_name == 'tdg':
            code += f"    // T†门: 量子位 {target}\n"
            code += f"    qcomp t_dagger_val = 1/std::sqrt(2) - 1_i/std::sqrt(2);\n"
            code += f"    DiagMatr1 t_dagger = getDiagMatr1({{1, t_dagger_val}});\n"
            code += f"    applyDiagMatr1(qureg, {target}, t_dagger);\n"

        elif op_name == 'id':
            code += f"    // 恒等门: 量子位 {target}\n"

        elif op_name == 'sx':
            code += f"    // √X门: 量子位 {target}\n"
            code += f"    CompMatr1 sx_matr = getCompMatr1({{\n"
            code += f"        {{0.5+0.5_i, 0.5-0.5_i}},\n"
            code += f"        {{0.5-0.5_i, 0.5+0.5_i}}\n"
            code += f"    }});\n"
            code += f"    applyCompMatr1(qureg, {target}, sx_matr);\n"

        elif op_name == 'sxdg':
            code += f"    // √X†门: 量子位 {target}\n"
            code += f"    CompMatr1 sxdg_matr = getCompMatr1({{\n"
            code += f"        {{0.5-0.5_i, 0.5+0.5_i}},\n"
            code += f"        {{0.5+0.5_i, 0.5-0.5_i}}\n"
            code += f"    }});\n"
            code += f"    applyCompMatr1(qureg, {target}, sxdg_matr);\n"

        elif op_name in ['rx', 'ry', 'rz']:
            angle = float(instruction.operation.params[0])
            axis = op_name[1].upper()  # 提取旋转轴 (X, Y, Z)
            code += f"    // R{axis}门({angle}): 量子位 {target}\n"
            code += f"    applyRotate{axis}(qureg, {target}, {angle});\n"

        elif op_name == 'p':
            angle = float(instruction.operation.params[0])
            code += f"    // 相位门({angle}): 量子位 {target}\n"
            code += f"    DiagMatr1 phase_matr = getDiagMatr1({{1, exp(1_i*{angle})}});\n"
            code += f"    applyDiagMatr1(qureg, {target}, phase_matr);\n"

        elif op_name == 'u':
            theta, phi, lam = [float(p) for p in instruction.operation.params]
            code += f"    // U门({theta}, {phi}, {lam}): 量子位 {target}\n"
            code += f"    qcomp u11 = cos({theta}/2);\n"
            code += f"    qcomp u12 = -exp(1_i*{lam}) * sin({theta}/2);\n"
            code += f"    qcomp u21 = exp(1_i*{phi}) * sin({theta}/2);\n"
            code += f"    qcomp u22 = exp(1_i*({phi}+{lam})) * cos({theta}/2);\n"
            code += "    CompMatr1 u_matr = getCompMatr1({{u11, u12}, {u21, u22}});\n"
            code += f"    applyCompMatr1(qureg, {target}, u_matr);\n"

        return code

    # ========================
    # 双量子位门操作
    # ========================
    def handle_two_qubit_gate(self, instruction, qubits, code):
        """处理双量子位门操作"""
        op_name = instruction.operation.name.lower()
        q0, q1 = qubits

        if op_name == 'cx':
            code += f"    // CNOT门: 控制位 {q0}, 目标位 {q1}\n"
            # code += get_targets_array([q1])
            # code += f"    applyControlledMultiQubitNot(qureg, {q0}, targets_{q1}, 1);\n"

            code += f"    applyControlledMultiQubitNot(qureg, {q0}, {', '.join(map(str, [q1]))}, 1);\n"

        elif op_name == 'cz':
            code += f"    // CZ门: 控制位 {q0}, 目标位 {q1}\n"
            code += f"    applyControlledPauliZ(qureg, {q0}, {q1});\n"

        elif op_name == 'cy':
            code += f"    // CY门: 控制位 {q0}, 目标位 {q1}\n"
            code += f"    applyControlledPauliY(qureg, {q0}, {q1});\n"

        elif op_name == 'ch':
            code += f"    // CH门: 控制位 {q0}, 目标位 {q1}\n"
            code += f"    applyControlledHadamard(qureg, {q0}, {q1});\n"

        elif op_name == 'swap':
            code += f"    // SWAP门: {q0} <-> {q1}\n"
            code += f"    applySwap(qureg, {q0}, {q1});\n"

        elif op_name == 'iswap':
            code += f"    // iSWAP门: {q0}, {q1}\n"
            # code += f"    CompMatr2 iswap_matr = getCompMatr2({{{{1, 0, 0, 0}},{{0, 0, 1_i, 0}},{{0, 1_i, 0, 0}},{{0, 0, 0, 1}}}})"
            code += f"    applyCompMatr2(qureg, {q0}, {q1}, getCompMatr2({{{{1, 0, 0, 0}},{{0, 0, 1_i, 0}},{{0, 1_i, 0, 0}},{{0, 0, 0, 1}}}}));\n"

        elif op_name in ['crx', 'cry', 'crz']:
            angle = float(instruction.operation.params[0])
            axis = op_name[2].upper()  # 提取旋转轴 (X, Y, Z)
            code += f"    // 控制R{axis}门({angle}): 控制位 {q0}, 目标位 {q1}\n"
            code += f"    applyControlledRotate{axis}(qureg, {q0}, {q1}, {angle});\n"

        elif op_name == 'cu1':
            angle = float(instruction.operation.params[0])
            code += f"    // 控制U1门({angle}): 控制位 {q0}, 目标位 {q1}\n"
            # code += f"    DiagMatr1 cu1_matr = getDiagMatr1({{1, exp(1_i*{angle})}});\n"
            code += f"    applyControlledDiagMatr1(qureg, {q0}, {q1}, getDiagMatr1({{1, exp(1_i*{angle})}}));\n"

        elif op_name == 'cu3':
            theta, phi, lam = [float(p) for p in instruction.operation.params]
            code += f"    // 控制U3门({theta},{phi},{lam}): 控制位 {q0}, 目标位 {q1}\n"
            # code += f"    qcomp u11 = cos({theta}/2);\n"
            # code += f"    qcomp u12 = -exp(1_i*{lam}) * sin({theta}/2);\n"
            # code += f"    qcomp u21 = exp(1_i*{phi}) * sin({theta}/2);\n"
            # code += f"    qcomp u22 = exp(1_i*({phi}+{lam})) * cos({theta}/2);\n"
            # code += "    CompMatr1 cu3_matr = getCompMatr1({{u11, u12}, {u21, u22}});\n"
            # code += f"    getCompMatr1({{cos({theta}/2), -exp(1_i*{lam}) * sin({theta}/2)}}, {{exp(1_i*{phi}) * sin({theta}/2), exp(1_i*({phi}+{lam})) * cos({theta}/2)}});\n"
            code += f"    applyControlledCompMatr1(qureg, {q0}, {q1}, getCompMatr1({{cos({theta}/2), -exp(1_i*{lam}) * sin({theta}/2)}}, {{exp(1_i*{phi}) * sin({theta}/2), exp(1_i*({phi}+{lam})) * cos({theta}/2)}}));\n"

        elif op_name == 'rxx':
            angle = float(instruction.operation.params[0])
            code += f"    // RXX门({angle}): {q0}, {q1}\n"
            code += f"    applyRotateX(qureg, {q0}, {angle});\n"
            code += f"    applyRotateX(qureg, {q1}, {angle});\n"
            code += f"    applyControlledPhaseGadget(qureg, {q0}, &{q1}, 1, -{angle});\n"

        elif op_name == 'ryy':
            angle = float(instruction.operation.params[0])
            code += f"    // RYY门({angle}): {q0}, {q1}\n"
            code += f"    applyRotateY(qureg, {q0}, {angle});\n"
            code += f"    applyRotateY(qureg, {q1}, {angle});\n"
            code += f"    applyControlledPhaseGadget(qureg, {q0}, &{q1}, 1, -{angle});\n"

        elif op_name == 'rzz':
            angle = float(instruction.operation.params[0])
            code += f"    // RZZ门({angle}): {q0}, {q1}\n"
            code += f"    applyControlledPhaseGadget(qureg, {q0}, &{q1}, 1, {angle});\n"

        elif op_name == 'rzx':
            angle = float(instruction.operation.params[0])
            code += f"    // RZX门({angle}): {q0}, {q1}\n"
            code += f"    applyHadamard(qureg, {q1});\n"
            code += f"    applyControlledPhaseGadget(qureg, {q0}, &{q1}, 1, {angle});\n"
            code += f"    applyHadamard(qureg, {q1});\n"

        return code

    # ========================
    # 多量子位门操作
    # ========================
    def handle_multi_qubit_gate(self, instruction, qubits, code):
        """处理多量子位门操作"""
        op_name = instruction.operation.name.lower()
        num_qubits = len(qubits)

        if op_name == 'ccx':
            controls = qubits[:2]
            target = qubits[2]
            code += f"    // Toffoli门: 控制位 {controls[0]} 和 {controls[1]}, 目标位 {target}\n"
            # code += f"    int toffoli_ctrls[] = {{{', '.join(map(str, controls))}}};\n"
            # code += get_targets_array([target])
            # code += f"    applyMultiControlledMultiQubitNot(qureg, toffoli_ctrls, 2, targets_{target}, 1);\n"
            code += f"    applyMultiControlledMultiQubitNot(qureg, {{{', '.join(map(str, controls))}}}, 2, {', '.join(map(str, [target]))}, 1);\n"

        elif op_name == 'cswap':
            code += f"    // Fredkin门: 控制位 {qubits[0]}, 目标位 {qubits[1]} 和 {qubits[2]}\n"
            code += f"    applyControlledSwap(qureg, {qubits[0]}, {qubits[1]}, {qubits[2]});\n"

        elif op_name in ['mcx', 'mcy', 'mcz']:
            num_controls = num_qubits - 1
            controls = qubits[:num_controls]
            target = qubits[-1]

            # 根据门类型选择对应的操作
            # gate_type = op_name[1:].upper()  # 提取门类型 (CX, CY, CZ)

            if op_name == 'mcx':
                code += f"    // 多控制X门: {num_controls} 控制位, 目标位 {target}\n"
                # code += f"    int mcx_ctrls[] = {{{', '.join(map(str, controls))}}};\n"
                # code += get_targets_array([target])
                # code += f"    applyMultiControlledMultiQubitNot(qureg, mcx_ctrls, {num_controls}, targets_{target}, 1);\n"
                code += f"    applyMultiControlledMultiQubitNot(qureg, {{{', '.join(map(str, controls))}}}, {num_controls}, {', '.join(map(str, [target]))}, 1);\n"

            elif op_name == 'mcy':
                code += f"    // 多控制Y门: {num_controls} 控制位, 目标位 {target}\n"
                # code += f"    int mcy_ctrls[] = {{{', '.join(map(str, controls))}}};\n"
                # code += get_targets_array([target])
                # code += f"    applyMultiControlledPauliY(qureg, mcy_ctrls, {num_controls}, targets_{target}, 1);\n"
                code += f"    applyMultiControlledPauliY(qureg, {{{', '.join(map(str, controls))}}}, {num_controls}, {', '.join(map(str, [target]))}, 1);\n"

            elif op_name == 'mcz':
                code += f"    // 多控制Z门: {num_controls} 控制位, 目标位 {target}\n"
                # code += f"    int mcz_ctrls[] = {{{', '.join(map(str, controls))}}};\n"
                # code += get_targets_array([target])
                # code += f"    applyMultiControlledPauliZ(qureg, mcz_ctrls, {num_controls}, targets_{target}, 1);\n"
                code += f"    applyMultiControlledPauliZ(qureg, {{{', '.join(map(str, controls))}}}, {num_controls}, {', '.join(map(str, [target]))}, 1);\n"

        elif op_name == 'mcp':
            angle = float(instruction.operation.params[0])
            num_controls = num_qubits - 1
            controls = qubits[:num_controls]
            target = qubits[-1]
            code += f"    // 多控制相位门({angle}): {num_controls} 控制位, 目标位 {target}\n"
            # code += f"    int mcp_ctrls[] = {{{', '.join(map(str, controls))}}};\n"
            # code += f"    DiagMatr1 phase_matr = getDiagMatr1({{1, exp(1_i*{angle})}});\n"
            # code += f"    applyMultiControlledDiagMatr1(qureg, mcp_ctrls, {num_controls}, {target}, phase_matr);\n"
            code += f"    applyMultiControlledDiagMatr1(qureg, {{{', '.join(map(str, controls))}}}, {num_controls}, {target}, getDiagMatr1({{1, exp(1_i*{angle})}}));\n"

        return code

    # ========================
    # 特殊操作 (测量、重置等)
    # ========================
    def handle_special_operations(self, instruction, qubits, code):
        """处理测量、重置等特殊操作"""
        op_name = instruction.operation.name.lower()
        target = qubits[0]

        '''
            switch (ab()) {
            case 1:
                printf("");
            case 2:
                break;
            default: ;
        }
        '''
        if op_name == 'reset':
            code += f"    // 重置量子位 {target}\n"
            code += f"    switch(qreal prob0 = calcProbOfQubitOutcome(qureg, {target}, 0)) {{\n"
            code += f"        case prob0 < 1e-12\n"
            code += f"            // 概率为零，强制重置为|1>\n"
            code += f"            applyPauliX(qureg, {target});\n"
            code += f"            break;\n"
            code += f"        case prob0 < 1.0\n"
            code += f"            // 叠加态，投影测量后重置\n"
            code += f"            if (applyQubitMeasurement(qureg, {target}) == 1) {{\n"
            code += f"                applyPauliX(qureg, {target});\n"
            code += f"            }}\n"
            code += f"            break;\n"
            code += f"    }}\n"

            # code += f"    qreal prob0 = calcProbOfQubitOutcome(qureg, {target}, 0);\n"
            # code += f"    if (prob0 < 1e-12) {{\n"
            # code += f"        // 概率为零，强制重置为|1>\n"
            # code += f"        applyPauliX(qureg, {target});\n"
            # code += f"    }} else if (prob0 < 1.0) {{\n"
            # code += f"        // 叠加态，投影测量后重置\n"
            # code += f"        int outcome = applyQubitMeasurement(qureg, {target});\n"
            # code += f"        if (outcome == 1) {{\n"
            # code += f"            applyPauliX(qureg, {target});\n"
            # code += f"        }}\n"
            # code += f"    }}\n"

        return code

    # ========================
    # 主转换函数
    # ========================
    def convert_instruction(self, instruction, qubit_map, code):
        """转换量子门指令为C++代码"""

        if instruction.operation.name.lower() not in self.support_gates:
            return ""
        # 获取映射后的量子位索引
        qubits = [qubit_map[q] for q in instruction.qubits]

        # 根据量子门类型分发到对应的处理函数
        num_qubits = len(qubits)

        if num_qubits == 1:
            code = self.handle_single_qubit_gate(instruction, qubits, code)
        elif num_qubits == 2:
            code = self.handle_two_qubit_gate(instruction, qubits, code)
        elif num_qubits >= 3:
            code = self.handle_multi_qubit_gate(instruction, qubits, code)

        # 处理特殊操作（测量、重置等）
        # code = handle_special_operations(instruction, qubits, code)

        return code

    def circuit_to_quest(self, qc):
        """将Qiskit量子电路转换为QuEST C++模拟代码"""

        # 获取量子比特数量
        num_qubits = qc.num_qubits
        num_clbits = len(qc.clbits)

        # 创建代码模板
        code = f"""
    /** @file
     * 由 Qiskit 生成的 QuEST 模拟代码
     */

    #include "quest/include/quest.h"
    #include <iostream>
    #include <vector>
    #include <string>
    #include <cmath>  // 包含数学函数
    #include "quest/include/json.hpp"

    using json = nlohmann::json;
    using std::vector;
    using std::string;

    int main() {{
        // 初始化QuEST环境
        initQuESTEnv();
        QuESTEnv env = getQuESTEnv();  // 获取环境信息

        // 初始化经典寄存器数组
        int creg[{num_clbits}] = {{0}};

        // 创建 {num_qubits} 量子比特系统
        Qureg qureg = createQureg({num_qubits});
        initZeroState(qureg);

        // 应用量子门
    """

        # 映射量子比特
        qubit_map = {qubit: idx for idx, qubit in enumerate(qc.qubits)}

        for instruction in qc:
            code = self.convert_instruction(instruction, qubit_map, code)

        code += f"""
        // 报告状态
        reportStr("Final state:");
        reportQureg(qureg);

        // 计算并报告概率分布
        if (env.rank == 0) {{
            std::cout << "\\n量子态概率分布:\\n";
            // 遍历所有可能的状态
            for (long long int i = 0; i < {2 ** num_qubits}; i++) {{
                // 获取状态i的振幅
                qcomp amp = getQuregAmp(qureg, i);

                // 使用real()和imag()函数获取实部和虚部
                qreal realPart = real(amp);
                qreal imagPart = imag(amp);

                // 计算概率
                double prob = realPart * realPart + imagPart * imagPart;

                // 打印状态及其概率
                std::cout << "|";
                for (int q = {num_qubits - 1}; q >= 0; q--) {{
                    std::cout << ((i >> q) & 1);
                }}
                std::cout << ">: " << prob << "\\n";
            }}
        }}

        // 清理资源
        destroyQureg(qureg);
        finalizeQuESTEnv();
        return 0;
    }}
    """
        return code

    def generate_json_files(self, qc: QuantumCircuit, shots: int = 1024) ->(str, str):
        """将Qiskit量子电路转换为 电路JSON 和 参数JSON """
        from . import params_compressor
        code_structure = json.dumps(self.encoder.encode_circuit(qc, shots), ensure_ascii=False)
        self.params = self.encoder.get_params()
        compressor = params_compressor.ParameterCompressor()
        compressor.compress(self.params)
        compressed_params = compressor.to_json()

        return code_structure, compressed_params

    def circuit_to_quest_json(self, qc):
        """将Qiskit量子电路转换为QuEST C++模拟代码"""

        # 获取量子比特数量
        num_qubits = qc.num_qubits
        num_clbits = len(qc.clbits)

        # 使用模式编码器
        # encoder = compress.QuantumPatternEncoder()
        result = self.encoder.encode_circuit(qc)

        # print("\n模式检测结果:")
        # print(json.dumps(result, indent=2))
        patterns = result["patterns"]

        # 创建代码模板
        code = f"""
/** @file
* 由 Qiskit 生成的 QuEST 模拟代码
*/

#include "quest/include/quest.h"
#include <cmath>  // 包含数学函数
#include "quest/src/jquantum/jquantum.h"
#include <fstream>

// 重复模式-开始==========
"""

        # 重复模式
        code += self.generate_patterns_code(patterns)

        # 主函数
        code += f"""
// 重复模式-结束==========

int main() {{
    // 加载参数
    ifstream f("./compressed.json");
    json compressed;
    f >> compressed;
    Params params;

    // 初始化QuEST环境
    initQuESTEnv();
    QuESTEnv env = getQuESTEnv();  // 获取环境信息

    // 初始化经典寄存器数组
    int creg[{num_clbits}] = {{0}};

    // 创建 {num_qubits} 量子比特系统
    Qureg qureg = createQureg({num_qubits});
    initZeroState(qureg);

    // 应用量子门-开始==========
"""

        # 映射量子比特
        # qubit_map = {qubit: idx for idx, qubit in enumerate(qc.qubits)}

        # for instruction in qc:
        #     code = self.convert_instruction(instruction, qubit_map, code)

        self.params = self.encoder.get_params()
        # pprint.pprint(len(self.params))
        # pprint.pprint(self.params)
        sequence = result["sequence"]
        sequence_code = self.generate_sequence_code(sequence, patterns)
        code += sequence_code



        code += f"""
    // 应用量子门-结束==========

    // 报告状态
    reportStr("Final state:");
    reportQureg(qureg);

    // 计算并报告概率分布
    if (env.rank == 0) {{
        cout << "\\n量子态概率分布:\\n";
        // 遍历所有可能的状态
        for (long long int i = 0; i < {2 ** num_qubits}; i++) {{
            // 获取状态i的振幅
            qcomp amp = getQuregAmp(qureg, i);

            // 使用real()和imag()函数获取实部和虚部
            qreal realPart = real(amp);
            qreal imagPart = imag(amp);

            // 计算概率
            double prob = realPart * realPart + imagPart * imagPart;

            // 打印状态及其概率
            cout << "|";
            for (int q = {num_qubits - 1}; q >= 0; q--) {{
                cout << ((i >> q) & 1);
            }}
            cout << ">: " << prob << "\\n";
        }}
    }}

    // 清理资源
    destroyQureg(qureg);
    finalizeQuESTEnv();
    return 0;
}}
"""
        return code
