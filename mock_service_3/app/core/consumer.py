import json
import logging
import threading
from kafka import KafkaConsumer
from app.core.config import settings

logger = logging.getLogger(__name__)

received_messages: list[dict] = []


def _consume():
    consumer = KafkaConsumer(
        settings.kafka_topic,
        bootstrap_servers=settings.kafka_broker,
        group_id=f"{settings.app_name}-group",
        value_deserializer=lambda v: json.loads(v.decode("utf-8")),
        auto_offset_reset="earliest",
    )
    logger.info("Consuming from topic: %s", settings.kafka_topic)
    for record in consumer:
        logger.info("Received [%s]: %s", settings.kafka_topic, record.value)
        received_messages.append(record.value)


def start_consumer():
    thread = threading.Thread(target=_consume, daemon=True)
    thread.start()
