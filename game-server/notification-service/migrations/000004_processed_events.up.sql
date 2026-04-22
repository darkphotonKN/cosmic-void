  CREATE TABLE processed_events (
      event_id      UUID NOT NULL,                                                                                                                                  
      event_type    TEXT NOT NULL,
      processed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),                                                                                                             
      PRIMARY KEY (event_id, event_type)                                                                                                                            
  );