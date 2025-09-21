def _indent(indent: int) -> str:
    return "\t" * indent

def name_to_code(op_name, indent, repeats=None):
    """处理单量子位门操作"""
    code = ""
    param_prefix = "params[0]"
    if repeats is not None:
        param_prefix = f"params[{repeats}]"

    # ========================
    # 单量子位门操作
    # ========================
    if op_name == 'h':
        code += _indent(indent) + f"// H门\n"
        code += _indent(indent) + f"applyHadamard(qureg, {param_prefix}[0]);\n"

    elif op_name == 'x':
        code += _indent(indent) + f"// X门\n"
        code += _indent(indent) + f"applyPauliX(qureg, {param_prefix}[0]);\n"

    elif op_name == 'y':
        code += _indent(indent) + f"// Y门\n"
        code += _indent(indent) + f"applyPauliY(qureg, {param_prefix}[0]);\n"

    elif op_name == 'z':
        code += _indent(indent) + f"// Z门\n"
        code += _indent(indent) + f"applyPauliZ(qureg, {param_prefix}[0]);\n"

    elif op_name == 's':
        code += _indent(indent) + f"// S门\n"
        code += _indent(indent) + f"applyS(qureg, {param_prefix}[0]);\n"

    elif op_name == 'sdg':
        code += _indent(indent) + f"// S†门\n"
        code += _indent(indent) + f"applyDiagMatr1(qureg, {param_prefix}[0], getDiagMatr1({{1, -1_i}}));\n"

    elif op_name == 't':
        code += _indent(indent) + f"// T门\n"
        code += _indent(indent) + f"applyT(qureg, {param_prefix}[0]);\n"

    elif op_name == 'tdg':
        code += _indent(indent) + f"// T†门\n"
        code += _indent(indent) + f"applyDiagMatr1(qureg, {param_prefix}[0], getDiagMatr1({{1, 1/sqrt(2) - 1_i/sqrt(2)}}));\n"

    elif op_name == 'id':
        code += _indent(indent) + f"// 恒等门\n"

    elif op_name == 'sx':
        code += _indent(indent) + f"// √X门\n"
        code += _indent(indent) + f"applyCompMatr1(qureg, {param_prefix}[0], CompMatr1 sx_matr = getCompMatr1({{ {{0.5+0.5_i, 0.5-0.5_i}}, {{0.5-0.5_i, 0.5+0.5_i}} }}));\n"

    elif op_name == 'sxdg':
        code += _indent(indent) + f"// √X†门\n"
        code += _indent(indent) + f"applyCompMatr1(qureg, {param_prefix}[0], getCompMatr1({{ {{0.5-0.5_i, 0.5+0.5_i}}, {{0.5+0.5_i, 0.5-0.5_i}} }}));\n"

    elif op_name in ['rx', 'ry', 'rz']:
        axis = op_name[1].upper()  # 提取旋转轴 (X, Y, Z)
        code += _indent(indent) + f"// R{axis}门\n"
        code += _indent(indent) + f"applyRotate{axis}(qureg, {param_prefix}[0], {param_prefix}[1]);\n"

    elif op_name == 'p':
        # angle = float(instruction.operation.params[0])
        code += _indent(indent) + f"// P相位门\n"
        code += _indent(indent) + f"applyDiagMatr1(qureg, {param_prefix}[0], getDiagMatr1({{1, exp(1_i*(double){param_prefix}[1])}}));\n"

    elif op_name == 'u':
        code += _indent(indent) + f"// U门\n"
        code += "applyCompMatr1(qureg, {param_prefix}[0], getCompMatr1({{cos(((double){param_prefix}[1])/2), -exp(1_i*((double){param_prefix}[3])) * sin(((double){param_prefix}[1])/2)}, {exp(1_i*(double){param_prefix}[2]) * sin(((double){param_prefix}[1])/2), exp(1_i*((double){param_prefix}[2]+(double){param_prefix}[3])) * cos(((double){param_prefix}[1])/2)}}));\n"

    # ========================
    # 双量子位门操作
    # ========================
    if op_name == 'cx':
        code += _indent(indent) + f"// CNOT门\n"
        code += _indent(indent) + f"applyControlledMultiQubitNot(qureg, {param_prefix}[0], ((vector<int>) {param_prefix}[1]).data(), 1);\n"

    elif op_name == 'cz':
        code += _indent(indent) + f"// CZ门\n"
        code += _indent(indent) + f"applyControlledPauliZ(qureg, {param_prefix}[0], {param_prefix}[1]);\n"

    elif op_name == 'cy':
        code += _indent(indent) + f"// CY门: 控制位\n"
        code += _indent(indent) + f"applyControlledPauliY(qureg, {param_prefix}[0], {param_prefix}[1]);\n"

    elif op_name == 'ch':
        code += _indent(indent) + f"// CH门: 控制位\n"
        code += _indent(indent) + f"applyControlledHadamard(qureg, {param_prefix}[0], {param_prefix}[1]);\n"

    elif op_name == 'swap':
        code += _indent(indent) + f"// SWAP门\n"
        code += _indent(indent) + f"applySwap(qureg, {param_prefix}[0], {param_prefix}[1]);\n"

    elif op_name == 'iswap':
        code += _indent(indent) + f"// iSWAP门\n"
        code += _indent(indent) + f"applyCompMatr2(qureg, {param_prefix}[0], {param_prefix}[1], getCompMatr2({{{{1, 0, 0, 0}},{{0, 0, 1_i, 0}},{{0, 1_i, 0, 0}},{{0, 0, 0, 1}}}}));\n"

    elif op_name in ['crx', 'cry', 'crz']:
        axis = op_name[2].upper()  # 提取旋转轴 (X, Y, Z)
        code += _indent(indent) + f"// 控制R{axis}门\n"
        code += _indent(indent) + f"applyControlledRotate{axis}(qureg, {param_prefix}[0], {param_prefix}[1], {param_prefix}[2]);\n"

    elif op_name == 'cu1':
        code += _indent(indent) + f"// 控制U1门\n"
        code += _indent(indent) + f"applyControlledDiagMatr1(qureg, {param_prefix}[0], {param_prefix}[1], getDiagMatr1({{1, exp(1_i*(double){param_prefix}[2])}}));\n"

    elif op_name == 'cu3':
        code += _indent(indent) + f"// 控制U3门\n"
        code += _indent(indent) + f"applyControlledCompMatr1(qureg, {param_prefix}[0], {param_prefix}[1], getCompMatr1({{cos(((double){param_prefix}[2])/2), -exp(1_i*(double){param_prefix}[4]) * sin(((double){param_prefix}[2])/2)}}, {{exp(1_i*(double){param_prefix}[3]) * sin(((double){param_prefix}[2])/2), exp(1_i*((double){param_prefix}[3]+(double){param_prefix}[4])) * cos(((double){param_prefix}[2])/2)}}));\n"

    elif op_name == 'rxx':
        code += _indent(indent) + f"// RXX门\n"
        code += _indent(indent) + f"applyRotateX(qureg, {param_prefix}[0], {param_prefix}[2]);\n"
        code += _indent(indent) + f"applyRotateX(qureg, {param_prefix}[1], {param_prefix}[2]);\n"
        code += _indent(indent) + f"applyControlledPhaseGadget(qureg, {param_prefix}[0], ((vector<int>) {{{param_prefix}[1]}}).data(), 1, -(double){param_prefix}[2]);\n"

    elif op_name == 'ryy':
        code += _indent(indent) + f"// RYY门\n"
        code += _indent(indent) + f"applyRotateY(qureg, {param_prefix}[0], {param_prefix}[2]);\n"
        code += _indent(indent) + f"applyRotateY(qureg, {param_prefix}[1], {param_prefix}[2]);\n"
        code += _indent(indent) + f"applyControlledPhaseGadget(qureg, {param_prefix}[0], ((vector<int>) {{{param_prefix}[1]}}).data(), 1, -(double){param_prefix}[2]);\n"

    elif op_name == 'rzz':
        code += _indent(indent) + f"// RZZ门\n"
        code += _indent(indent) + f"applyControlledPhaseGadget(qureg, {param_prefix}[0], ((vector<int>) {{{param_prefix}[1]}}).data(), 1, {param_prefix}[2]);\n"

    elif op_name == 'rzx':
        code += _indent(indent) + f"// RZX门\n"
        code += _indent(indent) + f"applyHadamard(qureg, {param_prefix}[1]);\n"
        code += _indent(indent) + f"applyControlledPhaseGadget(qureg, {param_prefix}[0], ((vector<int>) {{{param_prefix}[1]}}).data(), 1, {param_prefix}[2]);\n"
        code += _indent(indent) + f"applyHadamard(qureg, {param_prefix}[1]);\n"

    # ========================
    # 多量子位门操作
    # ========================
    if op_name == 'ccx':
        code += _indent(indent) + f"// Toffoli门\n"
        code += _indent(indent) + f"applyMultiControlledMultiQubitNot(qureg, ((vector<int>) {param_prefix}[0]).data(), 2, ((vector<int>) {param_prefix}[1]).data(), 1);\n"

    elif op_name == 'cswap':
        code += _indent(indent) + f"// Fredkin门\n"
        code += _indent(indent) + f"applyControlledSwap(qureg, {param_prefix}[0], {param_prefix}[1], {param_prefix}[2]);\n"

    elif op_name in ['mcx', 'mcy', 'mcz']:

        if op_name == 'mcx':
            code += _indent(indent) + f"// 多控制X门\n"
            code += _indent(indent) + f"applyMultiControlledMultiQubitNot(qureg, ((vector<int>) {param_prefix}[0]).data(), {param_prefix}[1], ((vector<int>) {param_prefix}[2]).data(), 1);\n"

        elif op_name == 'mcy':
            code += _indent(indent) + f"// 多控制Y门\n"
            code += _indent(indent) + f"applyMultiControlledPauliY(qureg, ((vector<int>) {param_prefix}[0]).data(), {param_prefix}[1], {param_prefix}[2], 1);\n"

        elif op_name == 'mcz':
            code += _indent(indent) + f"// 多控制Z门\n"
            code += _indent(indent) + f"applyMultiControlledPauliZ(qureg, ((vector<int>) {param_prefix}[0]).data(), {param_prefix}[1], {param_prefix}[2], 1);\n"

    elif op_name == 'mcp':
        code += _indent(indent) + f"// 多控制相位门\n"
        code += _indent(indent) + f"applyMultiControlledDiagMatr1(qureg, ((vector<int>) {param_prefix}[0]).data(), {param_prefix}[1], {param_prefix}[2], getDiagMatr1({{1, exp(1_i*{param_prefix}[3])}}));\n"

    # ========================
    # 特殊操作 (测量、重置等)
    # ========================
    # elif op_name == 'reset':
    #     code += _indent(indent) + f"// 重置量子位\n"
    #     code += _indent(indent) + f"switch(qreal prob0 = calcProbOfQubitOutcome(qureg, {param_prefix}[0], 0)) {{\n"
    #     code += _indent(indent) + f"    case prob0 < 1e-12\n"
    #     code += _indent(indent) + f"        // 概率为零，强制重置为|1>\n"
    #     code += _indent(indent) + f"        applyPauliX(qureg, {param_prefix}[0]);\n"
    #     code += _indent(indent) + f"        break;\n"
    #     code += _indent(indent) + f"    case prob0 < 1.0\n"
    #     code += _indent(indent) + f"        // 叠加态，投影测量后重置\n"
    #     code += _indent(indent) + f"        if (applyQubitMeasurement(qureg, {param_prefix}[0]) == 1) {{\n"
    #     code += _indent(indent) + f"            applyPauliX(qureg, {param_prefix}[0]);\n"
    #     code += _indent(indent) + f"        }}\n"
    #     code += _indent(indent) + f"        break;\n"
    #     code += _indent(indent) + f"}}\n"

    return code
