from kafka.admin import KafkaAdminClient, NewTopic
from kafka.errors import TopicAlreadyExistsError

BROKER = "localhost:9092"

TOPICS = [
    "mock-service-1",
    "mock-service-2",
    "mock-service-3",
    "orchestrator",
]

admin = KafkaAdminClient(bootstrap_servers=BROKER)

new_topics = [NewTopic(name=t, num_partitions=1, replication_factor=1) for t in TOPICS]

for topic in new_topics:
    try:
        admin.create_topics([topic])
        print(f"Created topic: {topic.name}")
    except TopicAlreadyExistsError:
        print(f"Already exists: {topic.name}")

admin.close()
