import app.services.billing
from app.services import billing as billing_module


def build(service: "app.services.billing.InvoiceService") -> "app.services.billing.InvoiceService":
    created = app.services.billing.InvoiceService()
    return billing_module.InvoiceService()
