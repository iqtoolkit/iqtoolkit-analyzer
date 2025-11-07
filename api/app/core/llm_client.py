from abc import ABC, abstractmethod
from typing import Optional, Dict, Any
import httpx
import logging
from app.core.config import settings

logger = logging.getLogger(__name__)

class LLMProvider(ABC):
    @abstractmethod
    async def analyze_query(self, query: str, context: str = "") -> str:
        """Analyze a query using the LLM provider"""
        pass

    @abstractmethod
    async def health_check(self) -> bool:
        """Check if the LLM provider is available"""
        pass

class OpenAIProvider(LLMProvider):
    def __init__(self):
        self.base_url = settings.openai_base_url
        self.model = settings.openai_model
        self.timeout = settings.openai_timeout
        self.api_key = settings.openai_api_key

    async def analyze_query(self, query: str, context: str = "") -> str:
        """Analyze query using OpenAI"""
        prompt = self._build_analysis_prompt(query, context)
        
        async with httpx.AsyncClient(timeout=self.timeout) as client:
            try:
                response = await client.post(
                    f"{self.base_url}/chat/completions",
                    headers={
                        "Authorization": f"Bearer {self.api_key}",
                        "Content-Type": "application/json"
                    },
                    json={
                        "model": self.model,
                        "messages": [
                            {"role": "system", "content": "You are a database performance expert."},
                            {"role": "user", "content": prompt}
                        ],
                        "temperature": settings.llm_temperature,
                        "max_tokens": settings.llm_max_tokens
                    }
                )
                response.raise_for_status()
                result = response.json()
                return result["choices"][0]["message"]["content"]
            except Exception as e:
                logger.error(f"OpenAI request failed: {e}")
                raise

    async def health_check(self) -> bool:
        """Check if OpenAI is accessible"""
        try:
            async with httpx.AsyncClient(timeout=5) as client:
                response = await client.get(
                    f"{self.base_url}/models",
                    headers={"Authorization": f"Bearer {self.api_key}"}
                )
                return response.status_code == 200
        except:
            return False

    def _build_analysis_prompt(self, query: str, context: str) -> str:
        """Build analysis prompt for OpenAI"""
        return f"""Analyze this database query for performance issues and provide specific optimization recommendations.

Query:
{query}

{f'Context:\n{context}' if context else ''}

Provide your analysis in the following format:

1. PERFORMANCE ISSUES
- List specific performance problems identified
- Include any anti-patterns found

2. INDEX RECOMMENDATIONS
- Provide complete CREATE INDEX statements
- Explain why each index will help

3. QUERY REWRITE SUGGESTIONS
- Show complete rewritten queries
- Explain the improvements

4. ESTIMATED IMPACT
- Quantify expected performance gains
- List any tradeoffs to consider

Be specific and actionable."""

class OllamaProvider(LLMProvider):
    def __init__(self):
        self.base_url = settings.ollama_base_url
        self.model = settings.ollama_model
        self.timeout = settings.ollama_timeout
        self.retry_count = settings.ollama_retry_count

    async def analyze_query(self, query: str, context: str = "") -> str:
        """Analyze query using Ollama"""
        prompt = self._build_analysis_prompt(query, context)
        
        for attempt in range(self.retry_count + 1):
            try:
                async with httpx.AsyncClient(timeout=self.timeout) as client:
                    response = await client.post(
                        f"{self.base_url}/api/generate",
                        json={
                            "model": self.model,
                            "prompt": prompt,
                            "stream": False,
                            "options": {
                                "temperature": settings.llm_temperature,
                            }
                        }
                    )
                    response.raise_for_status()
                    result = response.json()
                    return result["response"]
            except Exception as e:
                if attempt == self.retry_count:
                    logger.error(f"Ollama request failed after {self.retry_count} attempts: {e}")
                    raise
                logger.warning(f"Ollama attempt {attempt + 1} failed: {e}")
                continue

    async def health_check(self) -> bool:
        """Check if Ollama is available"""
        try:
            async with httpx.AsyncClient(timeout=5) as client:
                response = await client.get(f"{self.base_url}/api/tags")
                return response.status_code == 200
        except:
            return False

    def _build_analysis_prompt(self, query: str, context: str) -> str:
        """Build analysis prompt for Ollama"""
        return f"""You are a database performance expert. Analyze this query and provide specific optimization recommendations.

Query:
{query}

{f'Context:\n{context}' if context else ''}

Provide your analysis in the following format:

1. PERFORMANCE ISSUES
- List specific performance problems identified
- Include any anti-patterns found

2. INDEX RECOMMENDATIONS
- Provide complete CREATE INDEX statements
- Explain why each index will help

3. QUERY REWRITE SUGGESTIONS
- Show complete rewritten queries
- Explain the improvements

4. ESTIMATED IMPACT
- Quantify expected performance gains
- List any tradeoffs to consider

Be specific and actionable."""

class LLMClient:
    def __init__(self):
        """Initialize LLM client with configured providers"""
        self.providers: Dict[str, LLMProvider] = {}
        self._setup_providers()

    def _setup_providers(self):
        """Set up enabled LLM providers"""
        if settings.openai_enabled:
            self.providers["openai"] = OpenAIProvider()
        
        if settings.ollama_enabled:
            self.providers["ollama"] = OllamaProvider()
        
        if not self.providers:
            raise ValueError("No LLM providers enabled in configuration")

    async def analyze_query(self, query: str, context: str = "") -> str:
        """
        Analyze query using configured providers.
        Falls back to OpenAI if Ollama fails and fallback is enabled.
        """
        primary_provider = settings.llm_provider
        
        try:
            if primary_provider not in self.providers:
                raise ValueError(f"Primary provider '{primary_provider}' not configured")
            
            return await self.providers[primary_provider].analyze_query(query, context)
            
        except Exception as e:
            logger.error(f"Primary provider ({primary_provider}) failed: {e}")
            
            # Try fallback to OpenAI if enabled
            if (primary_provider != "openai" and 
                settings.openai_enabled and 
                settings.openai_fallback and 
                "openai" in self.providers):
                
                logger.info("Falling back to OpenAI")
                return await self.providers["openai"].analyze_query(query, context)
            
            raise

    async def health_check(self) -> Dict[str, bool]:
        """Check health of all configured providers"""
        results = {}
        for name, provider in self.providers.items():
            results[name] = await provider.health_check()
        return results

# Global LLM client instance
llm_client = LLMClient()