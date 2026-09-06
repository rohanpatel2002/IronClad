import logging
import numpy as np
from typing import List, Dict, Any, Optional

# Configure logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger("EmbeddingGenerator")

class EmbeddingGenerator:
    """
    Generates semantic embeddings for code snippets and failure logs
    to enable vector-based pattern matching of security incidents.
    """
    def __init__(self, chroma_host: str = "chromadb", chroma_port: int = 8000):
        try:
            import chromadb
            from chromadb.utils import embedding_functions
            logger.info(f"Connecting to ChromaDB at {chroma_host}:{chroma_port}")
            self.client = chromadb.HttpClient(host=chroma_host, port=chroma_port)
            self.embed_fn = embedding_functions.DefaultEmbeddingFunction()
            self.collection = self.client.get_or_create_collection(
                name="security_incidents",
                embedding_function=self.embed_fn
            )
            logger.info("Successfully connected to ChromaDB and initialized collection 'security_incidents'")
        except Exception as e:
            logger.error(f"Failed to initialize ChromaDB client: {e}")
            self.client = None
            self.collection = None


    def store_incident(self, incident_id: str, content: str, metadata: Dict[str, Any]):
        """
        Stores an incident in the vector database with error handling.
        """
        try:
            if not content:
                logger.warning(f"Attempted to store empty content for incident {incident_id}")
                return

            self.collection.add(
                documents=[content],
                metadatas=[metadata],
                ids=[incident_id]
            )
            logger.debug(f"Successfully stored incident {incident_id}")
        except Exception as e:
            logger.error(f"Error storing incident {incident_id}: {e}")

    def find_similar_incidents(self, content: str, n_results: int = 5) -> Dict[str, Any]:
        """
        Finds incidents with similar semantic profiles.
        """
        try:
            if not content:
                return {"ids": [], "distances": [], "documents": []}

            results = self.collection.query(
                query_texts=[content],
                n_results=n_results
            )
            logger.info(f"Found {len(results.get('ids', [[]])[0])} similar incidents")
            return results
        except Exception as e:
            logger.error(f"Error querying similar incidents: {e}")
            return {"ids": [], "distances": [], "documents": [], "error": str(e)}

    @staticmethod
    def generate_dense_vector(text: str, dim: int = 128) -> List[float]:
        """Generate a deterministic normalized vector representation for text."""
        seed = sum(ord(c) for c in text) % (2**32 - 1)
        rng = np.random.RandomState(seed)
        vec = rng.randn(dim)
        norm = np.linalg.norm(vec)
        if norm == 0:
            return [0.0] * dim
        return [float(x) for x in (vec / norm)]


