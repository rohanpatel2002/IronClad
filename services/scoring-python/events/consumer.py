import json
import os
import pika
import logging
import signal
import sys

class EventProcessor:
    def __init__(self, predictor=None):
        from scorer.ml_predictor import RiskPredictor
        self.predictor = predictor or RiskPredictor()
        self.processed_count = 0

    def process_event_payload(self, body: str) -> dict:
        event = json.loads(body)
        features = [
            event.get('lines_changed', 100),
            event.get('complexity', 5),
            event.get('author_trust', 0.9),
            event.get('blast_radius', 0.5)
        ]
        risk_score = self.predictor.predict_risk(features)
        self.processed_count += 1
        return {
            "service": event.get("service", "unknown"),
            "risk_score": risk_score,
            "status": "processed"
        }

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

    from scorer.ml_predictor import RiskPredictor
    predictor = RiskPredictor()

    def callback(ch, method, properties, body):
        logging.info(f"Received scoring event: {body}")
        event = json.loads(body)
        
        # Extract features for ML prediction
        # Mock features: [lines, complexity, author_trust, blast_radius]
        features = [
            event.get('lines_changed', 100),
            event.get('complexity', 5),
            event.get('author_trust', 0.9),
            event.get('blast_radius', 0.5)
        ]
        
        risk_score = predictor.predict_risk(features)
        logging.info(f"ML Predicted Risk Score for {event.get('service')}: {risk_score:.4f}")
        
        # In a real implementation, store risk_score in DB
        ch.basic_ack(delivery_tag=method.delivery_tag)

    channel.basic_qos(prefetch_count=1)
    channel.basic_consume(queue=queue, on_message_callback=callback)

    logging.info('Scoring consumer started. Waiting for events...')
    channel.start_consuming()

if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    consume_events()
