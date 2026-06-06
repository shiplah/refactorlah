from . import billing
from .billing import InvoiceService


def build_relative() -> InvoiceService:
    return billing.InvoiceService()
