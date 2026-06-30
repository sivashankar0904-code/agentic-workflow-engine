from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    app_name: str = "mock_service_1"
    debug: bool = False
    database_url: str = "sqlite:///./mock_service_1.db"
    kafka_broker: str = "localhost:9092"
    kafka_topic: str = "mock-service-1"

    class Config:
        env_file = ".env"


settings = Settings()
