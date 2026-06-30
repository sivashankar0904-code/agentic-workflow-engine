from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    app_name: str = "mock_service_3"
    debug: bool = False
    database_url: str = "sqlite:///./mock_service_3.db"
    kafka_broker: str = "localhost:9092"
    kafka_topic: str = "mock-service-3"

    class Config:
        env_file = ".env"


settings = Settings()
