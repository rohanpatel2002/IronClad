from setuptools import setup, find_packages

setup(
    name="ironclad-scoring",
    version="0.1.0",
    description="IRONCLAD risk scoring and failure grammar learning service",
    author="IRONCLAD Team",
    packages=find_packages(),
    python_requires=">=3.11",
    install_requires=[
        "flask==3.0.0",
        "python-dotenv==1.0.0",
        "psycopg2-binary==2.9.9",
        "numpy==1.24.3",
        "scikit-learn==1.3.0",
        "pydantic==2.4.2",
    ],
    extras_require={
        "dev": ["pytest==7.4.3", "pytest-cov==4.1.0"],
    },
)
