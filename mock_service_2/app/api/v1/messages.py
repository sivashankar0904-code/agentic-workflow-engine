from fastapi import APIRouter
from app.core.consumer import received_messages

router = APIRouter(prefix="/messages", tags=["messages"])


@router.get("/")
def get_messages():
    return {"service": "mock_service_2", "messages": received_messages}
