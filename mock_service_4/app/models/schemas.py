from pydantic import BaseModel


class ItemBase(BaseModel):
    name: str


class ItemCreate(ItemBase):
    pass


class ItemResponse(ItemBase):
    id: int

    class Config:
        from_attributes = True


class MessageRequest(BaseModel):
    message: str


class MessageResponse(BaseModel):
    reply: str
