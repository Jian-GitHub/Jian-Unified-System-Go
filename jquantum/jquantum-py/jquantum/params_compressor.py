import json
from typing import List, Dict, Any, Tuple, Union

class ParameterCompressor:
    def __init__(self):
        self.patterns: Dict[str, Any] = {}
        self.sequence: List[Dict[str, Union[str, int]]] = []
        self.pattern_cache: Dict[Tuple, str] = {}
        self.pattern_counter = 1
        self.flat_params: List[List[Any]] = []

    def _create_pattern_key(self, segment: List[List[Any]]) -> Tuple:
        # return tuple(tuple(p) for p in segment)
        return self.to_hashable(segment)

    def to_hashable(self, value):
        if isinstance(value, list):
            return tuple(self.to_hashable(v) for v in value)
        elif isinstance(value, dict):
            # dict 也可能存在于你的压缩结果中
            return tuple(sorted((k, self.to_hashable(v)) for k, v in value.items()))
        else:
            return value

    def _pattern_equal(self, a: List[List[Any]], b: List[List[Any]]) -> bool:
        if len(a) != len(b):
            return False
        return all(x == y for x, y in zip(a, b))

    def compress(self, param_list: List[List[Any]]):
        self.patterns.clear()
        self.sequence.clear()
        self.flat_params.clear()
        self.pattern_cache.clear()
        self.pattern_counter = 1

        i = 0
        n = len(param_list)

        while i < n:
            best_value = 1
            best_pattern = None
            best_length = 0
            best_repeat = 0

            max_L = min(50, (n - i) // 2)

            for L in range(2, max_L + 1):
                segment = param_list[i:i + L]
                repeat = 1
                j = i + L
                while j + L <= n and self._pattern_equal(segment, param_list[j:j + L]):
                    repeat += 1
                    j += L

                value = L * repeat
                if repeat > 1 and value > best_value:
                    best_value = value
                    best_pattern = segment
                    best_length = L
                    best_repeat = repeat

            if best_pattern:
                key = self._create_pattern_key(best_pattern)
                if key not in self.pattern_cache:
                    pname = f"p{self.pattern_counter}"
                    self.pattern_counter += 1
                    self.patterns[pname] = best_pattern
                    self.pattern_cache[key] = pname
                else:
                    pname = self.pattern_cache[key]

                self.sequence.append({'ref': pname, 'count': best_repeat})
                i += best_length * best_repeat
            else:
                self.sequence.append({'ref': None, 'count': 1})
                self.flat_params.append(param_list[i])
                i += 1

    def to_json(self) -> str:
        return json.dumps({
            'patterns': self.patterns,
            'sequence': self.sequence,
            'flat_params': self.flat_params
        }, ensure_ascii=False, indent=None)

    def get_params_range(self, start: int, count: int = 1) -> List[List[Any]]:
        result = []
        idx = 0
        flat_idx = 0

        for block in self.sequence:
            ref = block['ref']
            times = block['count']

            if ref is None:
                for _ in range(times):
                    if start <= idx < start + count:
                        result.append(self.flat_params[flat_idx])
                    idx += 1
                    flat_idx += 1
            else:
                pattern = self.patterns[ref]
                plen = len(pattern)
                total = times * plen
                for j in range(total):
                    if start <= idx < start + count:
                        result.append(pattern[j % plen])
                    idx += 1

            if idx >= start + count:
                break

        return result
