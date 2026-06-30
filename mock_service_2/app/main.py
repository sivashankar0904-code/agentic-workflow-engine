from fastapi import FastAPI
from app.api.v1 import items

app = FastAPI(title="Mock Service 2")

app.include_router(items.router, prefix="/api/v1")


@app.get("/health")
def health():
    return {"status": "ok", "service": "mock_service_2"}
