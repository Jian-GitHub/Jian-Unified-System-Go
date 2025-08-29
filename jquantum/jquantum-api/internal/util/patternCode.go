package util

import (
	"strings"
)

// _indent 生成指定数量的 tab 缩进
func _indent(indent int) string {
	return strings.Repeat("\t", indent)
}

// NameToCode 处理单量子位门操作，返回 C++ 代码字符串
func NameToCode(op_name string, indent int, repeats string) string {
	var code strings.Builder
	param_prefix := "params[0]"

	if repeats != "" {
		param_prefix = "params[" + repeats + "]"
	}

	// ========================
	// 单量子位门操作
	// ========================
	switch op_name {
	case "h":
		code.WriteString(_indent(indent) + "// H门\n")
		code.WriteString(_indent(indent) + "applyHadamard(qureg, " + param_prefix + "[0]);\n")
	case "x":
		code.WriteString(_indent(indent) + "// X门\n")
		code.WriteString(_indent(indent) + "applyPauliX(qureg, " + param_prefix + "[0]);\n")
	case "y":
		code.WriteString(_indent(indent) + "// Y门\n")
		code.WriteString(_indent(indent) + "applyPauliY(qureg, " + param_prefix + "[0]);\n")

	case "z":
		code.WriteString(_indent(indent) + "// Z门\n")
		code.WriteString(_indent(indent) + "applyPauliZ(qureg, " + param_prefix + "[0]);\n")

	case "s":
		code.WriteString(_indent(indent) + "// S门\n")
		code.WriteString(_indent(indent) + "applyS(qureg, " + param_prefix + "[0]);\n")

	case "sdg":
		code.WriteString(_indent(indent) + "// S†门\n")
		code.WriteString(_indent(indent) + "applyDiagMatr1(qureg, " + param_prefix + "[0], getDiagMatr1({1, -1_i}));\n")

	case "t":
		code.WriteString(_indent(indent) + "// T门\n")
		code.WriteString(_indent(indent) + "applyT(qureg, " + param_prefix + "[0]);\n")

	case "tdg":
		code.WriteString(_indent(indent) + "// T†门\n")
		code.WriteString(_indent(indent) + "applyDiagMatr1(qureg, " + param_prefix + "[0], getDiagMatr1({1, 1/sqrt(2) - 1_i/sqrt(2)}));\n")

	case "id":
		code.WriteString(_indent(indent) + "// 恒等门\n")

	case "sx":
		code.WriteString(_indent(indent) + "// √X门\n")
		code.WriteString(_indent(indent) + "applyCompMatr1(qureg, " + param_prefix + "[0], CompMatr1 sx_matr = getCompMatr1({ {0.5+0.5_i, 0.5-0.5_i}, {0.5-0.5_i, 0.5+0.5_i} }));\n")

	case "sxdg":
		code.WriteString(_indent(indent) + "// √X†门\n")
		code.WriteString(_indent(indent) + "applyCompMatr1(qureg, " + param_prefix + "[0], getCompMatr1({ {0.5-0.5_i, 0.5+0.5_i}, {0.5+0.5_i, 0.5-0.5_i} }));\n")

	case "rx", "ry", "rz":
		axis := strings.ToUpper(op_name[1:2]) // 提取旋转轴 (X, Y, Z)
		code.WriteString(_indent(indent) + "// R" + axis + "门\n")
		code.WriteString(_indent(indent) + "applyRotate" + axis + "(qureg, " + param_prefix + "[0], " + param_prefix + "[1]);\n")

	case "p":
		code.WriteString(_indent(indent) + "// P相位门\n")
		code.WriteString(_indent(indent) + "applyDiagMatr1(qureg, " + param_prefix + "[0], getDiagMatr1({1, exp(1_i*(double)" + param_prefix + "[1])}));\n")

	case "u":
		code.WriteString(_indent(indent) + "// U门\n")
		code.WriteString("applyCompMatr1(qureg, {param_prefix}[0], getCompMatr1({{cos(((double){param_prefix}[1])/2), -exp(1_i*((double){param_prefix}[3])) * sin(((double){param_prefix}[1])/2)}, {exp(1_i*(double){param_prefix}[2]) * sin(((double){param_prefix}[1])/2), exp(1_i*((double){param_prefix}[2]+(double){param_prefix}[3])) * cos(((double){param_prefix}[1])/2)}}));\n")

	// ========================
	// 双量子位门操作
	// ========================
	case "cx":
		code.WriteString(_indent(indent) + "// CNOT门\n")
		code.WriteString(_indent(indent) + "applyControlledMultiQubitNot(qureg, " + param_prefix + "[0], ((vector<int>) " + param_prefix + "[1]).data(), 1);\n")

	case "cz":
		code.WriteString(_indent(indent) + "// CZ门\n")
		code.WriteString(_indent(indent) + "applyControlledPauliZ(qureg, " + param_prefix + "[0], " + param_prefix + "[1]);\n")

	case "cy":
		code.WriteString(_indent(indent) + "// CY门: 控制位\n")
		code.WriteString(_indent(indent) + "applyControlledPauliY(qureg, " + param_prefix + "[0], " + param_prefix + "[1]);\n")

	case "ch":
		code.WriteString(_indent(indent) + "// CH门: 控制位\n")
		code.WriteString(_indent(indent) + "applyControlledHadamard(qureg, " + param_prefix + "[0], " + param_prefix + "[1]);\n")

	case "swap":
		code.WriteString(_indent(indent) + "// SWAP门\n")
		code.WriteString(_indent(indent) + "applySwap(qureg, " + param_prefix + "[0], " + param_prefix + "[1]);\n")

	case "iswap":
		code.WriteString(_indent(indent) + "// iSWAP门\n")
		code.WriteString(_indent(indent) + "applyCompMatr2(qureg, " + param_prefix + "[0], " + param_prefix + "[1], getCompMatr2({{1, 0, 0, 0},{0, 0, 1_i, 0},{0, 1_i, 0, 0},{0, 0, 0, 1}}));\n")

	case "crx", "cry", "crz":
		axis := strings.ToUpper(op_name[2:3]) // 提取旋转轴 (X, Y, Z)
		code.WriteString(_indent(indent) + "// 控制R" + axis + "门\n")
		code.WriteString(_indent(indent) + "applyControlledRotate" + axis + "(qureg, " + param_prefix + "[0], " + param_prefix + "[1], " + param_prefix + "[2]);\n")

	case "cu1":
		code.WriteString(_indent(indent) + "// 控制U1门\n")
		code.WriteString(_indent(indent) + "applyControlledDiagMatr1(qureg, " + param_prefix + "[0], " + param_prefix + "[1], getDiagMatr1({1, exp(1_i*(double)" + param_prefix + "[2])}));\n")

	case "cu3":
		code.WriteString(_indent(indent) + "// 控制U3门\n")
		code.WriteString(_indent(indent) + "applyControlledCompMatr1(qureg, " + param_prefix + "[0], " + param_prefix + "[1], getCompMatr1({cos(((double)" + param_prefix + "[2])/2), -exp(1_i*(double)" + param_prefix + "[4]) * sin(((double)" + param_prefix + "[2])/2)}, {exp(1_i*(double)" + param_prefix + "[3]) * sin(((double)" + param_prefix + "[2])/2), exp(1_i*((double)" + param_prefix + "[3]+(double)" + param_prefix + "[4])) * cos(((double)" + param_prefix + "[2])/2)}));\n")

	case "rxx":
		code.WriteString(_indent(indent) + "// RXX门\n")
		code.WriteString(_indent(indent) + "applyRotateX(qureg, " + param_prefix + "[0], " + param_prefix + "[2]);\n")
		code.WriteString(_indent(indent) + "applyRotateX(qureg, " + param_prefix + "[1], " + param_prefix + "[2]);\n")
		code.WriteString(_indent(indent) + "applyControlledPhaseGadget(qureg, " + param_prefix + "[0], ((vector<int>) {" + param_prefix + "[1]}).data(), 1, -(double)" + param_prefix + "[2]);\n")

	case "ryy":
		code.WriteString(_indent(indent) + "// RYY门\n")
		code.WriteString(_indent(indent) + "applyRotateY(qureg, " + param_prefix + "[0], " + param_prefix + "[2]);\n")
		code.WriteString(_indent(indent) + "applyRotateY(qureg, " + param_prefix + "[1], " + param_prefix + "[2]);\n")
		code.WriteString(_indent(indent) + "applyControlledPhaseGadget(qureg, " + param_prefix + "[0], ((vector<int>) {" + param_prefix + "[1]}).data(), 1, -(double)" + param_prefix + "[2]);\n")

	case "rzz":
		code.WriteString(_indent(indent) + "// RZZ门\n")
		code.WriteString(_indent(indent) + "applyControlledPhaseGadget(qureg, " + param_prefix + "[0], ((vector<int>) {" + param_prefix + "[1]}).data(), 1, " + param_prefix + "[2]);\n")

	case "rzx":
		code.WriteString(_indent(indent) + "// RZX门\n")
		code.WriteString(_indent(indent) + "applyHadamard(qureg, " + param_prefix + "[1]);\n")
		code.WriteString(_indent(indent) + "applyControlledPhaseGadget(qureg, " + param_prefix + "[0], ((vector<int>) {" + param_prefix + "[1]}).data(), 1, " + param_prefix + "[2]);\n")
		code.WriteString(_indent(indent) + "applyHadamard(qureg, " + param_prefix + "[1]);\n")

	// ========================
	// 多量子位门操作
	// ========================
	case "ccx":
		code.WriteString(_indent(indent) + "// Toffoli门\n")
		code.WriteString(_indent(indent) + "applyMultiControlledMultiQubitNot(qureg, ((vector<int>) " + param_prefix + "[0]).data(), 2, ((vector<int>) " + param_prefix + "[1]).data(), 1);\n")

	case "cswap":
		code.WriteString(_indent(indent) + "// Fredkin门\n")
		code.WriteString(_indent(indent) + "applyControlledSwap(qureg, " + param_prefix + "[0], " + param_prefix + "[1], " + param_prefix + "[2]);\n")

	case "mcx":
		code.WriteString(_indent(indent) + "// 多控制X门\n")
		code.WriteString(_indent(indent) + "applyMultiControlledMultiQubitNot(qureg, ((vector<int>) " + param_prefix + "[0]).data(), " + param_prefix + "[1], ((vector<int>) " + param_prefix + "[2]).data(), 1);\n")

	case "mcy":
		code.WriteString(_indent(indent) + "// 多控制Y门\n")
		code.WriteString(_indent(indent) + "applyMultiControlledPauliY(qureg, ((vector<int>) " + param_prefix + "[0]).data(), " + param_prefix + "[1], " + param_prefix + "[2], 1);\n")

	case "mcz":
		code.WriteString(_indent(indent) + "// 多控制Z门\n")
		code.WriteString(_indent(indent) + "applyMultiControlledPauliZ(qureg, ((vector<int>) " + param_prefix + "[0]).data(), " + param_prefix + "[1], " + param_prefix + "[2], 1);\n")

	case "mcp":
		code.WriteString(_indent(indent) + "// 多控制相位门\n")
		code.WriteString(_indent(indent) + "applyMultiControlledDiagMatr1(qureg, ((vector<int>) " + param_prefix + "[0]).data(), " + param_prefix + "[1], " + param_prefix + "[2], getDiagMatr1({1, exp(1_i*" + param_prefix + "[3])}));\n")

	}

	return code.String()
}
