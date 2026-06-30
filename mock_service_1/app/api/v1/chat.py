from fastapi import APIRouter
from app.models.schemas import MessageRequest, MessageResponse

router = APIRouter(prefix="/chat", tags=["chat"])


@router.post("/", response_model=MessageResponse)
def send_message(payload: MessageRequest):
    return MessageResponse(reply=f"Echo: {payload.message}")
