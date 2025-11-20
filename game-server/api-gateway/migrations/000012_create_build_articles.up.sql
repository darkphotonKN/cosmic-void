CREATE TABLE IF NOT EXISTS build_articles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    article_id UUID NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    build_id UUID NOT NULL REFERENCES builds(id) ON DELETE CASCADE,
    -- not repeatable
    UNIQUE(build_id, article_id)
);
-- optimize index 
CREATE INDEX idx_build_articles_build_id ON build_articles(build_id);
CREATE INDEX idx_build_articles_article_id ON build_articles(article_id); 