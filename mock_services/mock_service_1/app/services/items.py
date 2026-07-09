from app.models.schemas import ItemCreate, ItemResponse


def get_all_items() -> list[ItemResponse]:
    return []


def get_item_by_id(item_id: int) -> ItemResponse | None:
    return None


def create_item(payload: ItemCreate) -> ItemResponse:
    return ItemResponse(id=1, name=payload.name)
