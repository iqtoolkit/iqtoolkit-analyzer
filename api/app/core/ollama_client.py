import httpx
from typing import AsyncGenerator, Dict, Any
from app.core.config import settings
import logging

logger = logging.getLogger(__name__)

class OllamaClient:
    def __init__(self):
        self.base_url = settings.ollama_base_url
        self.model = settings.ollama_model
        self.timeout = settings.ollama_timeout
    
    async def analyze_query(self, query: str, context: str = "") -> str:
        """Send query to Ollama for analysis"""
        prompt = self._build_analysis_prompt(query, context)
        
        async with httpx.AsyncClient(timeout=self.timeout) as client:
            try:
                response = await client.post(
                    f"{self.base_url}/api/generate",
                    json={
                        "model": self.model,
                        "prompt": prompt,
                        "stream": False,
                        "options": {
                            "temperature": 0.1,  # Low temperature for consistent analysis
                            "top_p": 0.9
                        }
                    }
                )
                response.raise_for_status()
                result = response.json()
                return result.get("response", "")
            except httpx.HTTPError as e:
                logger.error(f"Ollama request failed: {e}")
                raise Exception(f"Analysis failed: {str(e)}")
    
    def _build_analysis_prompt(self, query: str, context: str) -> str:
        """Build the analysis prompt for Ollama"""
        return f"""You are an expert PostgreSQL database administrator and performance tuner.
        Analyze this slow query and provide specific, actionable optimization recommendations.

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

Focus on practical, implementable changes. Be specific and include actual SQL statements."""

ollama_client = OllamaClient()