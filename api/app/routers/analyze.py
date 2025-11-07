from fastapi import APIRouter, HTTPException
from app.models.schemas import QueryAnalysisRequest, QueryAnalysisResponse
from app.core.ollama_client import ollama_client
import logging

logger = logging.getLogger(__name__)
router = APIRouter(prefix="/analyze", tags=["analysis"])

@router.post("/query", response_model=QueryAnalysisResponse)
async def analyze_single_query(request: QueryAnalysisRequest):
    """Analyze a single PostgreSQL query"""
    try:
        logger.info(f"Analyzing query: {request.query[:100]}...")
        
        # Get Ollama analysis
        analysis = await ollama_client.analyze_query(
            query=request.query,
            context=request.context or ""
        )
        
        # TODO: Parse the analysis into structured response
        # For now, return a simplified response
        return QueryAnalysisResponse(
            query=request.query,
            issues=["Detailed analysis pending implementation"],
            index_suggestions=[],
            query_rewrites=[],
            anti_patterns=[],
            analysis_summary=analysis
        )
    except Exception as e:
        logger.error(f"Query analysis failed: {e}")
        raise HTTPException(status_code=500, detail=f"Analysis failed: {str(e)}")