from setuptools import setup, find_packages

setup(
    name="ironclad-semantic",
    version="0.1.0",
    description="IRONCLAD semantic intent classifier service",
    author="IRONCLAD Team",
    packages=find_packages(),
    python_requires=">=3.11",
    install_requires=[
        "flask==3.0.0",
        "python-dotenv==1.0.0",
        "psycopg2-binary==2.9.9",
        "anthropic==0.7.8",
        "pydantic==2.4.2",
    ],
    extras_require={
        "dev": ["pytest==7.4.3", "pytest-cov==4.1.0"],
    },
)
