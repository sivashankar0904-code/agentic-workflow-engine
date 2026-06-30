from kafka import KafkaProducer
from app.core.config import settings
import json

producer = KafkaProducer(
    bootstrap_servers=settings.kafka_broker,
    value_serializer=lambda v: json.dumps(v).encode("utf-8"),
)
