from pydantic_settings import BaseSettings
from typing import Optional

class Settings(BaseSettings):
    # Ollama
    ollama_base_url: str = "http://localhost:11434"
    ollama_model: str = "llama2:13b"  # Using llama2 13B as default
    ollama_timeout: int = 300
    
    # API
    api_host: str = "0.0.0.0"
    api_port: int = 8000
    api_workers: int = 4
    
    # Security
    api_key_enabled: bool = False
    api_key: Optional[str] = None
    
    # Analysis
    max_log_size_mb: int = 50
    max_queries_per_request: int = 100
    
    class Config:
        env_file = ".env"

settings = Settings()