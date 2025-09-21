from typing import *
from numpy import double

class ParamsGenerator:
    params = []

    def ensure_list(self, x):
        """返回list"""
        if isinstance(x, (list, tuple)):
            return list(x)
        else:
            return [x]

    # ========================
    # 单量子位门操作
    # ========================
    def handle_single_qubit_gate(self, instruction, qubits) -> None:
        """处理单量子位门操作"""
        op_name = instruction.operation.name.lower()
        target = qubits[0]
        gate_params = []
        # print(op_name)

        # if op_name in ['h', 'x', 'y', 'z', 's', 'sdg', 't', 'tdg', 'sx', 'sxdg', 'measure']:
        #     params.append(target)
        gate_params.append(target)

        # elif op_name in ['rx', 'ry', 'rz', 'p']:
        if op_name in ['rx', 'ry', 'rz', 'p']:
            angle = double(instruction.operation.params[0])
            # params.append(target)
            gate_params.append(angle)

        elif op_name == 'u':
            # theta1, phi2, lam3
            theta, phi, lam = [double(p) for p in instruction.operation.params]
            # params.append(target)
            gate_params.append(theta)
            gate_params.append(phi)
            gate_params.append(lam)

        # return gate_params
        self._set_params(gate_params)

    # ========================
    # 双量子位门操作
    # ========================
    def handle_two_qubit_gate(self, instruction, qubits) -> None:
        """处理双量子位门操作"""
        op_name = instruction.operation.name.lower()
        gate_params = []
        q0, q1 = qubits
        gate_params.append(q0)

        if op_name == 'cx':
            gate_params.append(self.ensure_list(q1))
        else:
            gate_params.append(q1)
            if op_name in ['cz', 'cy', 'ch', 'swap', 'iswap']:
                gate_params.append(q1)

            elif op_name == 'cu3':
                theta, phi, lam = [double(p) for p in instruction.operation.params]
                gate_params.append(theta)
                gate_params.append(phi)
                gate_params.append(lam)

            elif op_name in ['crx', 'cry', 'crz', 'cu1', 'rxx', 'ryy', 'rzz', 'rzx']:
                angle = double(instruction.operation.params[0])
                gate_params.append(angle)

        self._set_params(gate_params)

    # ========================
    # 多量子位门操作
    # ========================
    def handle_multi_qubit_gate(self, instruction, qubits) -> None:
        """处理多量子位门操作"""
        op_name = instruction.operation.name.lower()
        num_qubits = len(qubits)
        gate_params = []

        if op_name == 'ccx':
            controls = qubits[:2]
            target = qubits[2]
            gate_params.append(self.ensure_list(controls))
            gate_params.append(self.ensure_list(target))

        elif op_name == 'cswap':
            gate_params.append(qubits[0])
            gate_params.append(qubits[1])
            gate_params.append(qubits[2])

        elif op_name == 'mcx':
            num_controls = num_qubits - 1
            controls = qubits[:num_controls]
            target = qubits[-1]

            gate_params.append(self.ensure_list(controls))
            gate_params.append(num_controls)
            gate_params.append(self.ensure_list(target))

        elif op_name in ['mcy', 'mcz']:
            num_controls = num_qubits - 1
            controls = qubits[:num_controls]
            target = qubits[-1]

            gate_params.append(self.ensure_list(controls))
            gate_params.append(num_controls)
            gate_params.append(target)

        elif op_name == 'mcp':
            angle = double(instruction.operation.params[0])
            num_controls = num_qubits - 1
            controls = qubits[:num_controls]
            target = qubits[-1]
            gate_params.append(self.ensure_list(controls))
            gate_params.append(num_controls)
            gate_params.append(target)
            gate_params.append(angle)

        # return gate_params
        self._set_params(gate_params)

    def _set_params(self, params):
        self.params.append(params)

    def get_params(self) -> List[Any]:
        return self.params
