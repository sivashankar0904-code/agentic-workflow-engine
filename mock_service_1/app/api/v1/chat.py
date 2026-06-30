from fastapi import APIRouter
from app.models.schemas import MessageRequest, MessageResponse
from app.core.kafka import producer
from app.core.config import settings

router = APIRouter(prefix="/chat", tags=["chat"])


@router.post("/", response_model=MessageResponse)
def send_message(payload: MessageRequest):
    producer.send(settings.kafka_topic, {"message": payload.message})
    producer.flush()
    return MessageResponse(reply=f"Published to {settings.kafka_topic}: {payload.message}")
