import json
import os
import pika
import logging
import signal
import sys

def consume_events():
    url = os.getenv('RABBITMQ_URL', 'amqp://guest:guest@localhost:5672/')
    queue = 'scoring_events'

    params = pika.URLParameters(url)
    connection = pika.BlockingConnection(params)
    channel = connection.channel()

    def graceful_exit(signum, frame):
        logging.info("Graceful shutdown initiated...")
        channel.stop_consuming()
        connection.close()
        sys.exit(0)

    signal.signal(signal.SIGINT, graceful_exit)
    signal.signal(signal.SIGTERM, graceful_exit)

    channel.queue_declare(queue=queue, durable=True)

    def callback(ch, method, properties, body):
        logging.info(f"Received scoring event: {body}")
        event = json.loads(body)
        # In a real implementation, we would call the scoring logic here
        # and store the result in the database.
        logging.info(f"Processed event for service: {event.get('service')}")
        ch.basic_ack(delivery_tag=method.delivery_tag)

    channel.basic_qos(prefetch_count=1)
    channel.basic_consume(queue=queue, on_message_callback=callback)

    logging.info('Scoring consumer started. Waiting for events...')
    channel.start_consuming()

if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    consume_events()
