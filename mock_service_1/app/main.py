from fastapi import FastAPI
from app.api.v1 import items, chat

app = FastAPI(title="Mock Service 1")

app.include_router(items.router, prefix="/api/v1")
app.include_router(chat.router, prefix="/api/v1")


@app.get("/health")
def health():
    return {"status": "ok", "service": "mock_service_1"}
