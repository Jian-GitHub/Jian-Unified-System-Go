import base64
import io
import json
import zipfile

import requests
from qiskit import QuantumCircuit
from qiskit.result import Result
from . import converter


class Client:
    def __init__(self, token):
        if token is None or len(token) == 0:
            raise ValueError("token is required")
        self._converter = converter.Converter()
        self._headers = {
            'Authorization': f'{token}'
        }
        server = 'http://jquantum.api.JianUnifiedSystem.com:30501'
        self._submitURL: str = server + '/v1/job/submit'
        self._retrieveResultURL: str = server + '/v1/job/retrieveResult'
        self._clusterInfoURL: str = server + '/v1/job/clusterInfo'

    def info(self) -> dict | None:
        response = requests.post( self._clusterInfoURL, headers=self._headers)
        if response.status_code == 200:
            info = response.json()
            return {
                'cpu': info['totalCpu'],
                'mem': info['totalMem'],
                'max_qubits': info['maxQubits'],
                'nodes_num': info['nodes'],
            }
        return None

    def submit(self, qc: QuantumCircuit) -> str | None:
        code_json, params_json = self._converter.generate_json_files(qc)
        # create zip
        zip_buffer = io.BytesIO()
        with zipfile.ZipFile(zip_buffer, 'w', zipfile.ZIP_DEFLATED) as zf:
            zf.writestr('structure.json', code_json)
            zf.writestr('params.json', params_json)
        zip_buffer.seek(0)
        files = {
            'ariadne': ('thread.zip', zip_buffer, 'application/zip')
        }
        response = requests.post(self._submitURL, files=files, headers=self._headers)
        if response.status_code == 200 and response.json()['code'] == 200:
            return response.json()['jobId']
        print(response.status_code)
        return None

    def retrieve_result(self, job_id: str) -> Result | None:
        data = {
            "jobId": job_id
        }
        response = requests.post(self._retrieveResultURL, json=data, headers=self._headers)
        if response.status_code == 200 and response.json()['code'] == 200:
            result_string = base64.b64decode(response.json().get('result')).decode("utf-8")
            # Load JSON string as Python dictionary
            data_dict = json.loads(result_string)
            # Create Result object using from_dict method
            return Result.from_dict(data_dict)
        return None
