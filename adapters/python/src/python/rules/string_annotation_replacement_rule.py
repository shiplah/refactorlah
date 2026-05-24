from __future__ import annotations

import ast
import tokenize
from dataclasses import dataclass
from io import StringIO

from src.protocol.response import Replacement
from src.python.file_context import PythonFileContext
from src.python.module_mapping import ModuleMapping
from src.python.offsets import byte_slice


@dataclass(frozen=True)
class StringAnnotationReplacementRule:
    replaces_string_tokens: bool = True

    def collect(self, context: PythonFileContext, mappings: tuple[ModuleMapping, ...]) -> tuple[Replacement, ...]:
        try:
            tree = ast.parse(context.content)
        except SyntaxError:
            return ()

        annotation_spans = set(_string_annotation_spans(tree))
        if not annotation_spans:
            return ()

        line_offsets = _line_offsets(context.content)
        replacements: list[Replacement] = []
        for token in tokenize.generate_tokens(StringIO(context.content).readline):
            if token.type != tokenize.STRING:
                continue
            if (token.start, token.end) not in annotation_spans:
                continue

            token_start = line_offsets[token.start[0] - 1] + token.start[1]
            for mapping in mappings:
                search_start = 0
                while True:
                    index = token.string.find(mapping.old_module, search_start)
                    if index < 0:
                        break
                    start, end = byte_slice(
                        context.content,
                        token_start + index,
                        token_start + index + len(mapping.old_module),
                    )
                    replacements.append(
                        Replacement(
                            file=context.file,
                            start=start,
                            end=end,
                            replacement=mapping.new_module,
                            reason="python-string-annotation",
                            rule=self.__class__.__name__,
                        )
                    )
                    search_start = index + len(mapping.old_module)

        return tuple(replacements)


def _string_annotation_spans(tree: ast.AST) -> tuple[tuple[tuple[int, int], tuple[int, int]], ...]:
    spans: list[tuple[tuple[int, int], tuple[int, int]]] = []
    for node in ast.walk(tree):
        for field_name in ("annotation", "returns"):
            annotation = getattr(node, field_name, None)
            if isinstance(annotation, ast.Constant) and isinstance(annotation.value, str):
                spans.append(((annotation.lineno, annotation.col_offset), (annotation.end_lineno, annotation.end_col_offset)))
    return tuple(spans)


def _line_offsets(content: str) -> list[int]:
    offsets = [0]
    for index, char in enumerate(content):
        if char == "\n":
            offsets.append(index + 1)
    return offsets
