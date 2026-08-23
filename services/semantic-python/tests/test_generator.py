from embeddings.generator import EmbeddingGenerator

def test_generate_dense_vector():
    text = "feat(auth): add JWT token refresh endpoint"
    vec = EmbeddingGenerator.generate_dense_vector(text, dim=64)
    assert len(vec) == 64
    assert type(vec[0]) == float
