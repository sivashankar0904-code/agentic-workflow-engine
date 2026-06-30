from fastapi import HTTPException, status

ROLES: dict[str, list[str]] = {
    "admin": ["read", "write", "delete"],
    "editor": ["read", "write"],
    "viewer": ["read"],
}


def has_permission(role: str, permission: str) -> bool:
    return permission in ROLES.get(role, [])


def require_permission(role: str, permission: str) -> None:
    if not has_permission(role, permission):
        raise HTTPException(status_code=status.HTTP_403_FORBIDDEN, detail="Forbidden")
