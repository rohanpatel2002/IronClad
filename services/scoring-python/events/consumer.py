import json
import os
import pika
import logging

def consume_events():
    url = os.getenv('RABBITMQ_URL', 'amqp://guest:guest@localhost:5672/')
    queue = 'scoring_events'

    params = pika.URLParameters(url)
    connection = pika.BlockingConnection(params)
    channel = connection.channel()

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
    try:
        channel.start_consuming()
    except KeyboardInterrupt:
        channel.stop_consuming()
    connection.close()

if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    consume_events()
