from pydantic import BaseModel, Field
from typing import List, Optional
from datetime import datetime

class QueryAnalysisRequest(BaseModel):
    query: str = Field(..., min_length=1, max_length=50000, description="PostgreSQL query to analyze")
    context: Optional[str] = Field(None, max_length=5000, description="Additional context (table schemas, etc.)")

class IndexSuggestion(BaseModel):
    table: str
    columns: List[str]
    create_statement: str
    estimated_impact: str

class QueryRewrite(BaseModel):
    original: str
    suggested: str
    reason: str

class AntiPattern(BaseModel):
    pattern: str
    severity: str  # "low", "medium", "high"
    description: str
    fix: str

class QueryAnalysisResponse(BaseModel):
    query: str
    issues: List[str]
    index_suggestions: List[IndexSuggestion]
    query_rewrites: List[QueryRewrite]
    anti_patterns: List[AntiPattern]
    analysis_summary: str
    analyzed_at: datetime = Field(default_factory=datetime.utcnow)