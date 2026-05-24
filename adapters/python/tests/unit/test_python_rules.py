from __future__ import annotations

import unittest

from src.protocol.response import Replacement
from src.python.file_context import PythonFileContext
from src.python.module_mapping import ModuleMapping
from src.python.rules.imported_module_reference_replacement_rule import ImportedModuleReferenceReplacementRule
from src.python.rules.import_replacement_rule import ImportReplacementRule
from src.python.rules.qualified_reference_replacement_rule import QualifiedReferenceReplacementRule
from src.python.rules.relative_import_replacement_rule import RelativeImportReplacementRule


MAPPING = ModuleMapping(
    old_path="src/app/services/billing.py",
    new_path="src/app/domain/invoicing.py",
    old_module="app.services.billing",
    new_module="app.domain.invoicing",
    old_leaf="billing",
    new_leaf="invoicing",
)


class PythonRulesTest(unittest.TestCase):
    def test_import_rule_updates_import_module(self) -> None:
        content = "import app.services.billing\n"

        replacements = ImportReplacementRule().collect(context(content), (MAPPING,))

        self.assertEqual("import app.domain.invoicing\n", apply_replacements(content, replacements))

    def test_import_rule_updates_from_import_module(self) -> None:
        content = "from app.services.billing import InvoiceService\n"

        replacements = ImportReplacementRule().collect(context(content), (MAPPING,))

        self.assertEqual("from app.domain.invoicing import InvoiceService\n", apply_replacements(content, replacements))

    def test_import_rule_updates_parent_from_import_and_imported_leaf(self) -> None:
        content = "from app.services import billing\n"

        replacements = ImportReplacementRule().collect(context(content), (MAPPING,))

        self.assertEqual("from app.domain import invoicing\n", apply_replacements(content, replacements))

    def test_qualified_reference_rule_updates_exact_module_references(self) -> None:
        content = "value = app.services.billing.InvoiceService()\n"

        replacements = QualifiedReferenceReplacementRule().collect(context(content), (MAPPING,))

        self.assertEqual("value = app.domain.invoicing.InvoiceService()\n", apply_replacements(content, replacements))

    def test_imported_module_reference_rule_updates_visible_module_leaf(self) -> None:
        content = "from app.services import billing\nvalue = billing.InvoiceService()\n"

        replacements = [
            *ImportReplacementRule().collect(context(content), (MAPPING,)),
            *ImportedModuleReferenceReplacementRule().collect(context(content), (MAPPING,)),
        ]

        self.assertEqual("from app.domain import invoicing\nvalue = invoicing.InvoiceService()\n", apply_replacements(content, tuple(replacements)))

    def test_rules_emit_byte_offsets(self) -> None:
        content = "# café\nimport app.services.billing\n"

        replacements = ImportReplacementRule().collect(context(content), (MAPPING,))

        self.assertEqual("# café\nimport app.domain.invoicing\n", apply_replacements(content, replacements))

    def test_relative_import_rule_updates_relative_module_import(self) -> None:
        content = "from .billing import InvoiceService\n"

        replacements = RelativeImportReplacementRule().collect(
            context(content, module="app.services.consumer", package="app.services"),
            (MAPPING,),
        )

        self.assertEqual("from app.domain.invoicing import InvoiceService\n", apply_replacements(content, replacements))

    def test_relative_import_rule_updates_relative_parent_import_and_leaf(self) -> None:
        content = "from . import billing\nvalue = billing.InvoiceService()\n"
        file_context = context(content, module="app.services.consumer", package="app.services")

        replacements = [
            *RelativeImportReplacementRule().collect(file_context, (MAPPING,)),
            *ImportedModuleReferenceReplacementRule().collect(file_context, (MAPPING,)),
        ]

        self.assertEqual("from app.domain import invoicing\nvalue = invoicing.InvoiceService()\n", apply_replacements(content, tuple(replacements)))


def apply_replacements(content: str, replacements: tuple[Replacement, ...]) -> str:
    result = content.encode()
    for replacement in sorted(replacements, key=lambda item: item.start, reverse=True):
        result = result[: replacement.start] + replacement.replacement.encode() + result[replacement.end :]
    return result.decode()


def context(content: str, module: str = "app.use", package: str = "app") -> PythonFileContext:
    return PythonFileContext(file="src/app/use.py", content=content, module=module, package=package)


if __name__ == "__main__":
    unittest.main()
