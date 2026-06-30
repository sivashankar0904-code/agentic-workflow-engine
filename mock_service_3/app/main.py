from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.api.v1 import items, messages
from app.core.consumer import start_consumer


@asynccontextmanager
async def lifespan(app: FastAPI):
    start_consumer()
    yield


app = FastAPI(title="Mock Service 3", lifespan=lifespan)

app.add_middleware(CORSMiddleware, allow_origins=["*"], allow_methods=["*"], allow_headers=["*"])

app.include_router(items.router, prefix="/api/v1")
app.include_router(messages.router, prefix="/api/v1")


@app.get("/health")
def health():
    return {"status": "ok", "service": "mock_service_3"}
