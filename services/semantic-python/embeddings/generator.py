import numpy as np
import chromadb
from chromadb.utils import embedding_functions

class EmbeddingGenerator:
    """
    Generates semantic embeddings for code snippets and failure logs
    to enable vector-based pattern matching of security incidents.
    """
    def __init__(self, chroma_host="chromadb", chroma_port=8000):
        self.client = chromadb.HttpClient(host=chroma_host, port=chroma_port)
        self.embed_fn = embedding_functions.DefaultEmbeddingFunction()
        self.collection = self.client.get_or_create_collection(
            name="security_incidents",
            embedding_function=self.embed_fn
        )

    def store_incident(self, incident_id, content, metadata):
        """
        Stores an incident in the vector database.
        """
        self.collection.add(
            documents=[content],
            metadatas=[metadata],
            ids=[incident_id]
        )

    def find_similar_incidents(self, content, n_results=5):
        """
        Finds incidents with similar semantic profiles.
        """
        results = self.collection.query(
            query_texts=[content],
            n_results=n_results
        )
        return results
