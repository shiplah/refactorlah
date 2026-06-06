import app.services.billing


def generated() -> app.services.billing.InvoiceService:
    return app.services.billing.InvoiceService()
