-- Create answers table
CREATE TABLE answers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    selected_option VARCHAR(255),
    answer_text TEXT,
    user_id UUID NOT NULL,
    question_id UUID NOT NULL,
    question_set_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create question_mappings table
CREATE TABLE question_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question_id UUID NOT NULL,
    campaign_id UUID NOT NULL,
    org_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- -- Add PostgreSQL update trigger for updated_at column
-- CREATE OR REPLACE FUNCTION update_modified_column()
-- RETURNS TRIGGER AS 
-- BEGIN
--     NEW.updated_at = NOW();
--     RETURN NEW;
-- END;
--  LANGUAGE 'plpgsql';

-- -- Apply trigger to answers table
-- CREATE TRIGGER update_answers_modtime
--     BEFORE UPDATE ON answers
--     FOR EACH ROW
--     EXECUTE FUNCTION update_modified_column();

-- -- Apply trigger to question_mappings table
-- CREATE TRIGGER update_question_mappings_modtime
--     BEFORE UPDATE ON question_mappings
--     FOR EACH ROW
--     EXECUTE FUNCTION update_modified_column();

-- Add indexes to improve query performance
CREATE INDEX idx_answers_user_id ON answers(user_id);
CREATE INDEX idx_answers_question_id ON answers(question_id);
CREATE INDEX idx_answers_question_set_id ON answers(question_set_id);
CREATE INDEX idx_question_mappings_question_id ON question_mappings(question_id);
CREATE INDEX idx_question_mappings_campaign_id ON question_mappings(campaign_id);
CREATE INDEX idx_question_mappings_org_id ON question_mappings(org_id);