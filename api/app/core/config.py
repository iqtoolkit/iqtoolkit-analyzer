from pydantic_settings import BaseSettings
from typing import Optional, List
from .config_manager import config_manager

class Settings(BaseSettings):
    # Load from config manager with environment variable override
    
    # LLM Common Settings
    llm_provider: str = config_manager.get("llm.provider", "ollama")
    llm_temperature: float = config_manager.get("llm.temperature", 0.1)
    llm_max_tokens: int = config_manager.get("llm.max_tokens", 2000)
    
    # OpenAI (v0.1.5 compatibility)
    openai_enabled: bool = config_manager.get("llm.openai.enabled", True)
    openai_api_key: Optional[str] = config_manager.get("llm.openai.api_key", None)
    openai_base_url: str = config_manager.get("llm.openai.base_url", "https://api.openai.com/v1")
    openai_model: str = config_manager.get("llm.openai.model", "gpt-4")
    openai_timeout: int = config_manager.get("llm.openai.timeout", 300)
    openai_fallback: bool = config_manager.get("llm.openai.fallback", True)
    
    # Ollama (v0.2.0)
    ollama_enabled: bool = config_manager.get("llm.ollama.enabled", True)
    ollama_base_url: str = config_manager.get("llm.ollama.base_url", "http://localhost:11434")
    ollama_model: str = config_manager.get("llm.ollama.model", "llama2:13b")
    ollama_timeout: int = config_manager.get("llm.ollama.timeout", 300)
    ollama_retry_count: int = config_manager.get("llm.ollama.retry_count", 3)
    
    # API
    api_host: str = config_manager.get("api.host", "0.0.0.0")
    api_port: int = config_manager.get("api.port", 8000)
    api_workers: int = config_manager.get("api.workers", 4)
    allowed_origins: List[str] = config_manager.get("api.allowed_origins", ["http://localhost:8000"])
    
    # Security
    api_key_enabled: bool = config_manager.get("security.api_key_enabled", False)
    api_key: Optional[str] = config_manager.get("security.api_key", None)
    cors_enabled: bool = config_manager.get("security.cors_enabled", True)
    
    # Analysis
    max_log_size_mb: int = config_manager.get("analysis.max_log_size_mb", 50)
    max_queries_per_request: int = config_manager.get("analysis.max_queries_per_request", 100)
    min_duration_ms: int = config_manager.get("analysis.min_duration_ms", 1000)
    top_n_queries: int = config_manager.get("analysis.top_n_queries", 10)
    supported_formats: List[str] = config_manager.get("analysis.formats", ["plain", "csv", "json"])
    
    # Reports
    report_formats: List[str] = config_manager.get("reports.formats", ["markdown", "html"])
    syntax_highlighting: bool = config_manager.get("reports.syntax_highlighting", True)
    include_execution_plan: bool = config_manager.get("reports.include_execution_plan", True)
    max_queries_per_report: int = config_manager.get("reports.max_queries_per_report", 50)
    
    class Config:
        env_file = ".env"

    def reload(self):
        """Reload configuration from file and environment"""
        config_manager.reload()
        self.__init__()

settings = Settings()