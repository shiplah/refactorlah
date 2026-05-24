import importlib

import app.services.billing
from app.services import billing as billing_module
from app.services.billing import InvoiceService


def build() -> "app.services.billing.InvoiceService":
    service = app.services.billing.InvoiceService()
    alias_service = billing_module.InvoiceService()
    imported_service = InvoiceService()
    literal = "app.services.billing.InvoiceService"
    importlib.import_module("dynamic.name")
    return service or alias_service or imported_service or literal
