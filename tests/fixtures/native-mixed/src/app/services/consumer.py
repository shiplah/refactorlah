from . import billing


def build():
    return billing.InvoiceService()
