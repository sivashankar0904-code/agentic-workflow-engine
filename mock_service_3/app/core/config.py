from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    app_name: str = "mock_service_3"
    debug: bool = False
    database_url: str = "sqlite:///./mock_service_3.db"

    class Config:
        env_file = ".env"


settings = Settings()
